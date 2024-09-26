package ezcomm

import (
	"encoding/xml"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"gitee.com/bon-ami/eztools/v6"
)

const (
	// FlowActRcv action receive
	FlowActRcv = "receive"
	// FlowActSnd action send
	FlowActSnd = "send"

	// FlowVarSign sign of vars
	FlowVarSign = "%"
	// FlowVarSep separator for vars
	FlowVarSep = "."
	// FlowVarLst var listen
	FlowVarLst = "listen"
	// FlowVarLcl var local
	FlowVarLcl = "local"
	// FlowVarPee var peer
	FlowVarPee = "peer"
	// FlowVarFil var file
	FlowVarFil = "file"
	// FlowVarScr var script
	FlowVarScr = "script"

	// FlowRcvLen is the size of receive buffer
	FlowRcvLen = 1024 * 1024 // must > FileHdr1stLen+len(EzcName)
	// FlowComLen is the size of the queue between EZComm and UI
	FlowComLen = 99
	// FlowFilLen is the size of send buffer for files
	FlowFilLen = 1024 * 1024
)

const (
	// FlowParseValSimple is a string without <FlowVarSign>
	FlowParseValSimple = iota
	// FlowParseValSign is for <FlowVarSign><string><FlowVarSign>
	FlowParseValSign
	// FlowParseValVar is for <FlowVarSign><xml tag name of FlowConnStruc/FlowStepStruc><FlowVarSign><string><FlowVarSign>
	FlowParseValVar
)

const (
	// FlowChnLst is not used by EZ Comm
	FlowChnLst = iota
	// FlowChnEnd is to end a connection/server, to EZ Comm,
	//	or, a connection ends, from EZ Comm
	FlowChnEnd
	// FlowChnDie is not used by EZ Comm
	FlowChnDie
	// FlowChnSnd is to send sth, to EZ Comm
	//	or, sth is sent, from EZ Comm
	FlowChnSnd
	// FlowChnSndFil is FlowChnSnd for files
	FlowChnSndFil
	// FlowChnSnt is not used by EZ Comm
	FlowChnSnt
	// FlowChnSntFil is FlowChnSnt for files, not used by EZ Comm
	//FlowChnSntFil
	// FlowChnRcv is sth received, from EZ Comm
	FlowChnRcv
	// FlowChnRcvFil is FlowChnRcv for files, not used by EZ Comm. For flow only, currently
	FlowChnRcvFil
)

// FlowStepStruc a step in a flow
type FlowStepStruc struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt string `xml:",chardata"`

	Act string `xml:"action,attr"`
	// Name if not null, this structure will be mapped to Vals
	Name string `xml:"name,attr"`
	// Dest will be updated upon UDP receive action, if it is a variable
	Dest string `xml:"dest,attr"`
	Data string `xml:"data,attr"`
	// Loop rounds to repeat this step
	// 0, 1: no loop
	// > 1: number of rounds
	// < 0: infinitely
	Loop  int  `xml:"loop,attr"`
	Block bool `xml:"block,attr"`
	// Steps: sub steps triggered
	Steps []FlowStepStruc `xml:"step"`
	// curr: current sub step
	curr int
}

// FlowConnStruc a connection in a flow
type FlowConnStruc struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt string `xml:",chardata"`

	Name     string `xml:"name,attr"`
	Protocol string `xml:"protocol,attr"`
	Addr     string `xml:"address,attr"`
	Peer     string `xml:"peer,attr"`
	Block    bool   `xml:"block,attr"`
	// TODO: use Wait?
	Wait  string          `xml:"wait,attr"`
	Steps []FlowStepStruc `xml:"step"`
	lstnr net.Listener
	conn  *net.UDPConn
	// chanErrs is for Server()
	chanErrs chan error
	chanStrs chan string
	chanComm chan RoutCommStruc
	// TODO: for ending
	//chanStrus [FlowChnStrMax]chan struct{}
	//lock      sync.Mutex
	// flow: servers point back to the slice in FlowStruc
	//flow     *FlowStruc
	wait4Svr *FlowConnStruc
	wait4Act string
}

// FlowStruc defines the structure of a flow xml
type FlowStruc struct {
	// Root of the XML
	Root xml.Name `xml:"ezcommFlow"`
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt   string          `xml:",chardata"`
	Conns []FlowConnStruc `xml:"conn"`
	Vals  map[string]*FlowStepStruc
}

// FlowWriterNew creation of a writer
var FlowWriterNew func(string) (io.WriteCloser, error)

// FlowReaderNew creation of a reader
var FlowReaderNew func(string) (io.ReadCloser, error)

// ParseVar parses a string of a simple string or
//
//	  <FlowVarSign>[<xml tag name of FlowConnStruc/FlowStepStruc><FlowVarSign>]<string><FlowVarSign>
//	  fun() is invoked for matched FlowConnStruc,
//		with index of it in FlowStruc.Conns and <string>
//
// Return values:
//
//	1st.
//	 If FlowConnStruc is matched, the value of its member whose xml tag is <string>
//	 Otherwise, the input <string>
//	2nd. FlowParseVal*, that's to say, FlowParseValSimple, FlowParseValSign, or FlowParseValVar
func (flow FlowStruc) ParseVar(str string,
	fun func(int, string)) (string, int) {
	//eztools.Log("parsevar enter", str)
	if len(str) < 1 {
		return str, FlowParseValSimple
	}
	parts := strings.Split(str, FlowVarSign)
	if len(parts) == 3 {
		return parts[1], FlowParseValSign
	}
	if len(parts) != 4 {
		//eztools.LogWtTime("unrecognized var", str)
		return str, FlowParseValSimple
	}
	if fun != nil {
		for i := range flow.Conns {
			if parts[1] != flow.Conns[i].Name {
				continue
			}
			fun(i, parts[2])
		}
	}
	if step1, ok := flow.Vals[parts[1]]; ok && step1 != nil {
		//eztools.Log("parsevar", parts, *step1)
		stepTp := reflect.TypeOf(*step1)
		for i := 0; i < stepTp.NumField(); i++ {
			fld := stepTp.Field(i)
			tg := fld.Tag.Get("xml")
			tg = strings.TrimSuffix(tg, ",attr")
			if parts[2] != tg {
				//eztools.Log("passing", tg)
				continue
			}
			f := stepTp.FieldByIndex(fld.Index)
			v := reflect.ValueOf(*step1).FieldByName(f.Name)
			//eztools.Log("got", v.String())
			/*str, ok := v.(string)
			if !ok {
				break
			}*/
			return v.String(), FlowParseValVar
		}
	}

	return parts[1] + FlowVarSign + parts[2], FlowParseValSign
}

// ParseData parses data in step
// Return values:
//
//		1st. is one of following
//		  data string in form of a simple string
//		  data string where substituted if possible, the value of a member of FlowConnStruc or FlowStepStruc for <string> in <FlowVarSign><xml tag name of FlowConnStruc or FlowStepStruc><FlowVarSep><string><FlowVarSign>
//		  file name for <string> in <FlowVarSign><FlowVarFil><FlowVarSep><string><FlowVarSign>
//		2nd.
//		 0: a string
//		 1(=difference from FlowChnRcv to FlowChnRcvFil): a file name
//	         2: a script name
func (step FlowStepStruc) ParseData(flow FlowStruc,
	conn FlowConnStruc) (string, int) {
	retWhole, parseRes := flow.ParseVar(step.Data, nil) // TODO: match a server
	switch parseRes {
	case FlowParseValSimple, FlowParseValSign:
		const valDiff = FlowChnRcvFil - FlowChnRcv
		switch {
		case strings.HasPrefix(retWhole, FlowVarFil+FlowVarSign):
			return strings.TrimPrefix(retWhole,
				FlowVarFil+FlowVarSign), valDiff
		case strings.HasPrefix(retWhole, FlowVarScr+FlowVarSign):
			return strings.TrimPrefix(retWhole,
				FlowVarScr+FlowVarSign), valDiff + 1
		}
	}
	return retWhole, 0
}

// ParseDest parses destination in flow
func (step FlowStepStruc) ParseDest(flow FlowStruc,
	conn FlowConnStruc) *net.UDPAddr {
	var ret string
	retWhole, parseRes := flow.ParseVar(step.Dest,
		func(svrInd int, varStr string) {
			switch varStr {
			case FlowVarLcl:
				ret = flow.Conns[svrInd].Addr
			}
		})
	if len(ret) < 1 {
		ret = retWhole
		switch parseRes {
		case FlowParseValSimple:
			//eztools.Log("parsedest", ret)
		default:
			switch retWhole {
			case FlowVarPee:
				ret = conn.Peer
			}
			//eztools.Log("parsedest by var", ret, conn)
		}
	} else {
		//eztools.Log("parsedest by server", ret)
	}
	addr, err := net.ResolveUDPAddr(conn.Protocol, ret)
	if err != nil {
		return nil
	}
	return addr
}

// Step1 runs 1 step in flow
func (conn FlowConnStruc) Step1(flow *FlowStruc, step *FlowStepStruc) {
	for {
		if eztools.Verbose > 2 {
			eztools.LogWtTime("step", step.Name, "action", step.Act)
		}
		respChn := make(chan RoutCommStruc, FlowComLen)
		var respStruc RoutCommStruc
		receivingFile4Script := ""
		switch step.Act {
		case FlowActSnd:
			dest := step.ParseDest(*flow, conn)
			if dest == nil && eztools.Verbose > 1 {
				eztools.LogWtTime("NO dest parsed for",
					step.Act, "as", step.Dest)
			}
			data, fil := step.ParseData(*flow, conn)
			if fil > (FlowChnSndFil - FlowChnSnd) {
				if eztools.Verbose > 2 {
					eztools.LogWtTime("exec",
						data, "--local="+conn.Addr,
						"--remote="+dest.String(),
						"--file="+data+FlowVarSep+step.Act)
				}
				cmd := exec.Command(data, "--local="+conn.Addr,
					"--remote="+dest.String(),
					"--file="+data+FlowVarSep+step.Act)
				if errors.Is(cmd.Err, exec.ErrDot) {
					cmd.Err = nil
				}
				if cmd.Err != nil {
					eztools.LogWtTime("failed to make command", data+FlowVarSep+step.Act, cmd.Err)
				} else {
					if err := cmd.Run(); err != nil {
						eztools.LogWtTime("failed to exec", data+FlowVarSep+step.Act, err)
					}
				}
				fil = FlowChnSndFil - FlowChnSnd
				data += FlowVarSep + step.Act
			}
			conn.chanComm <- RoutCommStruc{
				Act:     FlowChnSnd + fil,
				PeerUDP: dest,
				Data:    []byte(data),
				Resp:    respChn,
			}
		case FlowActRcv:
			data, fil := step.ParseData(*flow, conn)
			if fil > (FlowChnRcvFil - FlowChnRcv) {
				receivingFile4Script = data
				data += FlowVarSep + step.Act
				fil = FlowChnRcvFil - FlowChnRcv
			}
			conn.chanComm <- RoutCommStruc{
				Act:  FlowChnRcv + fil,
				Data: []byte(data),
				Resp: respChn,
			}
		}
		respStruc = <-respChn
		close(respChn)
		if respStruc.Err != nil {
			eztools.LogWtTime(conn.Name, step.Act, respStruc.Err)
		} else {
			if respStruc.PeerUDP != nil {
				if eztools.Verbose > 1 {
					eztools.Log("refreshing dest of",
						step.Name, "from", step.Dest,
						"to",
						respStruc.PeerUDP.String())
				}
				step.Dest = respStruc.PeerUDP.String()
			}
			if receivingFile4Script != "" {
				if eztools.Verbose > 2 {
					eztools.LogWtTime("exec",
						receivingFile4Script,
						"--local="+conn.Addr,
						"--remote="+step.Dest,
						"--file="+string(respStruc.Data))
				}
				if data, err := os.ReadFile(string(respStruc.Data)); err != nil {
					eztools.LogWtTime("failed to get received data", err)
				} else {
					step.Data = string(data)
				}
				cmd := exec.Command(receivingFile4Script,
					"--local="+conn.Addr,
					"--remote="+step.Dest,
					"--file="+receivingFile4Script+FlowVarSep+step.Act)
				if errors.Is(cmd.Err, exec.ErrDot) {
					cmd.Err = nil
				}
				if cmd.Err != nil {
					eztools.LogWtTime("failed to make command", receivingFile4Script, cmd.Err)
				} else {
					if err := cmd.Run(); err != nil {
						eztools.LogWtTime("failed to exec", receivingFile4Script, err)
					}
				}
			} else {
				step.Data = string(respStruc.Data)
			}
			if len(step.Name) > 0 {
				flow.Vals[step.Name] = step
			}
			if eztools.Verbose > 2 {
				eztools.LogWtTime("step done",
					step.Name, "action", step.Act,
					"data", step.Data)
			}
			conn.StepAll(flow, step.Steps)
		}
		switch {
		case (step.Loop == 0 || step.Loop == 1):
			return
		case step.Loop < 0:
			continue
		default:
			step.Loop--
		}
	}
}

// StepAll runs steps in flow
func (conn FlowConnStruc) StepAll(flow *FlowStruc, steps []FlowStepStruc) {
	for i, s := range steps {
		if s.Block {
			conn.Step1(flow, &steps[i])
		} else {
			go conn.Step1(flow, &steps[i])
		}
		//svr.curr = i + 1
	}
}

// Connected when connected in flow
func (conn *FlowConnStruc) Connected(logFunc FuncLog,
	connFunc FuncConn, connTCP net.Conn, addr [2]string) {
	defer func() {
		// duplcate for TCP server, but does not matter
		conn.chanErrs <- eztools.ErrAbort
	}()
	for {
		// TODO: replace all with log()
		if eztools.Verbose > 2 {
			logFunc(conn.Name, "waiting")
		}
		// TODO: when to close this connUdp?
		com := <-conn.chanComm
		if eztools.Verbose > 2 {
			logFunc(conn.Name, "got command", com)
		}
		connUDP := conn.conn
		var (
			sndFunc func([]byte) error
			rcvFunc func([]byte, func([]byte, int) error) error
		)
		if connTCP != nil {
			defer connTCP.Close()
			sndFunc = func(buf []byte) (err error) {
				_, err = connTCP.Write(buf)
				return
			}
			rcvFunc = func(buf []byte,
				fun func([]byte, int) error) error {
				ln, err := connTCP.Read(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						logFunc(connTCP.LocalAddr().
							String(), "reading",
							err)
						com.Err = err
						return err
					}
				}
				if err := fun(buf, ln); err != nil {
					return err
				}
				return err
			}
		} else {
			if connUDP == nil {
				eztools.LogFatal("NO connections connected")
				return
			}
			defer connUDP.Close()
			sndFunc = func(buf []byte) (err error) {
				if com.PeerUDP != nil {
					_, err = connUDP.WriteTo(buf,
						com.PeerUDP)
				} else { // TODO: default to current peer?
					_, err = connUDP.Write(buf)
				}
				return err
			}
			rcvFunc = func(buf []byte,
				fun func([]byte, int) error) error {
				//logFunc(connStruc.Name, "to recv", connUdp.LocalAddr().String())
				ln, addr, err := connUDP.ReadFromUDP(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						logFunc(conn.Name,
							connUDP.LocalAddr().
								String(),
							"reading error", err)
						com.Err = err
						return err
					}
				}
				//logFunc(connStruc.Name, "recv", err)
				if com.PeerUDP == nil {
					com.PeerUDP = addr
					//logFunc("setting peer", com.Peer)
				} else {
					logFunc("peer not null", com.PeerUDP)
				}
				if err := fun(buf, ln); err != nil {
					return err
				}
				return err
			}
		}
		switch com.Act {
		case FlowChnEnd:
			if eztools.Verbose > 1 {
				logFunc(conn.Name, "ending")
			}
			return
		case FlowChnSnd:
			com.Err = sndFunc(com.Data)
		case FlowChnSndFil:
			if err := conn.sndFil(logFunc, connUDP, com, sndFunc); err != nil {
				com.Err = err
				break
			}
		case FlowChnRcv:
			// com.Peer must be empty
			// com.Data is appended
			var byteLen int
			bytes := make([]byte, FlowRcvLen)
			for {
				err := rcvFunc(bytes,
					func(buf []byte, ln int) error {
						com.Data = append(com.Data,
							buf[:ln]...)
						byteLen = ln
						return nil
					})
				if eztools.Verbose > 2 {
					logFunc(conn.Name,
						"received", err,
						string(bytes[:byteLen]))
				}
				//if err != nil {
				break
				//}
			}
		case FlowChnRcvFil:
			if err := conn.rcvFil(logFunc, com, rcvFunc); err != nil {
				com.Err = err
				break
			}
		}
		if eztools.Verbose > 2 {
			logFunc(conn.Name, "replying", com)
		}
		com.Resp <- com
	}
}

func (conn *FlowConnStruc) sndFil(logFunc FuncLog, connUDP *net.UDPConn,
	com RoutCommStruc, sndFunc func([]byte) error) error {
	if FlowReaderNew == nil {
		return nil
	}
	buf := make([]byte, FlowFilLen)
	fr, err := FlowReaderNew(string(com.Data))
	if err != nil {
		logFunc("failed to open file to read!", string(com.Data), err)
		return err
	}
	defer fr.Close()
	// TODO: how to tell of pieces on peer?
	for {
		var ln int
		ln, err = fr.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logFunc(connUDP.LocalAddr(),
					"error reading!", err)
			}
			break
		}
		err = sndFunc([]byte(buf[:ln]))
		if err != nil {
			logFunc("failed send!", err)
			com.Err = err
			break
		}
	}
	return nil
}

func (conn *FlowConnStruc) rcvFil(logFunc FuncLog, com RoutCommStruc,
	rcvFunc func([]byte, func([]byte, int) error) error) error {
	if FlowWriterNew == nil {
		logFunc("NO flow writer!")
		return nil
	}
	fw, err := FlowWriterNew(string(com.Data))
	if err != nil {
		logFunc("failed to open file to save!", string(com.Data), err)
		return err
	}
	bytes := make([]byte, FlowFilLen)
	//for {
	err = rcvFunc(bytes, func(buf []byte, ln int) error {
		/*if fw == nil {
			return eztools.ErrOutOfBound
		}*/
		if _, err := fw.Write(buf[:ln]); err != nil {
			logFunc("failed to write to", string(com.Data), err)
			return err
		}
		/*fw.Close()
		fw = nil*/
		return nil
	})
	if eztools.Verbose > 2 {
		logFunc(conn.Name,
			"saved", err, com.Data)
	}
	//if err != nil {
	//return nil
	//}
	//}
	fw.Close()
	return nil
}

// ParsePeer parses peer in struct
func (conn *FlowConnStruc) ParsePeer(flow FlowStruc) {
	//flow := svr.flow
	flow.ParseVar(conn.Peer, func(svrInd int, varStr string) {
		switch varStr {
		case FlowVarLst:
			conn.wait4Svr = &flow.Conns[svrInd]
			conn.wait4Act = FlowVarLst
			if flow.Conns[svrInd].chanStrs == nil {
				flow.Conns[svrInd].chanStrs =
					make(chan string, 1)
			} else {
				curr := cap(flow.Conns[svrInd].chanStrs)
				close(flow.Conns[svrInd].chanStrs)
				flow.Conns[svrInd].chanStrs =
					make(chan string, curr+1)
			}
			break
		}
	})
}

// Wait4 waits for chan from server
func (conn *FlowConnStruc) Wait4(flow FlowStruc) (ret string) {
	if (conn.wait4Svr == nil) || (conn.wait4Svr.chanStrs == nil) {
		return
	}
	conn.LockLog(conn.wait4Svr.Name, true)
	ret = <-conn.wait4Svr.chanStrs
	conn.LockLog(conn.wait4Svr.Name, false)
	return ret
}

// LockLog un-/locks for log
func (conn *FlowConnStruc) LockLog(nm string, lck bool) {
	if eztools.Verbose < 3 {
		return
	}
	var str string
	switch lck {
	case true:
		str = "waiting"
		//flow.Servers[svrW].lock.Lock()
	case false:
		//flow.Servers[svrW].lock.Unlock()
		str = "waited"
	}
	eztools.LogWtTime(conn.Name, str+" for", nm)
}

// Run runs a flow
// to be run by RunFlow, b/c no (de-)init for channels
func (conn *FlowConnStruc) Run(flow *FlowStruc) {
	if conn.chanComm == nil {
		conn.chanComm = make(chan RoutCommStruc, FlowComLen)
		defer func() {
			close(conn.chanComm)
			conn.chanComm = nil
		}()
	}
	if len(conn.Peer) > 1 {
		conn.RunCln(*flow)
	} else {
		conn.RunSvr(*flow)
	}
	conn.StepAll(flow, conn.Steps)
	if eztools.Verbose > 1 {
		eztools.LogWtTime("connection", conn.Name, "ending")
	}
	conn.chanComm <- RoutCommStruc{Act: FlowChnEnd}
}

// RunCln runs a client
// to be run by FlowConnStruc.Run(), b/c no (de-)init for channels
func (conn *FlowConnStruc) RunCln(flow FlowStruc) {
	if eztools.Verbose > 2 {
		eztools.LogWtTime("client", conn.Name, conn.Protocol)
	}
	conn.Peer = conn.Wait4(flow)
	_, po, err := net.SplitHostPort(conn.Peer)
	if err != nil {
		eztools.LogWtTime("failed to parse socket for", conn.Peer, err)
	}
	conn.Peer = net.JoinHostPort(Localhost, po)
	if eztools.Verbose > 1 {
		eztools.LogWtTime("client", conn.Name,
			conn.Protocol, conn.Addr, "->", conn.Peer)
	}
	if strings.HasPrefix(conn.Protocol, "udp") {
		clnt, err := ListenUDP(conn.Protocol, conn.Addr)
		if err != nil {
			eztools.LogWtTime(conn.Name,
				"failed to connect to", conn.Peer)
			return
		}
		conn.conn = clnt
		if eztools.Verbose > 0 {
			eztools.LogWtTime("client", conn.Name,
				"local", clnt.LocalAddr().String())
		}
		go func() {
			conn.Connected(eztools.LogWtTime, nil, nil, [2]string{})
		}()
	} else {
		clnt, err := Client(eztools.LogWtTime, nil,
			conn.Protocol, conn.Peer, conn.Connected)
		if err != nil {
			eztools.LogWtTime(conn.Name,
				"failed to connect to", conn.Peer)
			return
		}
		if eztools.Verbose > 0 {
			eztools.LogWtTime("client", conn.Name,
				"local", clnt.LocalAddr().String(),
				"remote", clnt.RemoteAddr().String())
		}
	}
}

// RunSvr supports TCP & UDP only. TODO: IP & Unix
// to be run by FlowConnStruc.Run(), b/c no (de-)init for channels
func (conn *FlowConnStruc) RunSvr(flow FlowStruc) {
	if len(conn.Protocol) < 1 {
		return
	}
	if eztools.Verbose > 2 {
		eztools.LogWtTime("server", conn.Name, conn.Protocol)
	}
	//svr.lock.Lock()
	if strings.HasPrefix(conn.Protocol, "udp") {
		var err error
		conn.conn, err = ListenUDP(conn.Protocol, conn.Addr)
		if err != nil {
			eztools.LogWtTime(conn.Name, "failed to listen")
			return
		}
		conn.Addr = conn.conn.LocalAddr().String()
		go func() {
			/*defer func() {
				svr.conn.Close()
			}()*/
			conn.Connected(eztools.LogWtTime, nil, nil, [2]string{})
		}()
	} else {
		lstnr, err := ListenTCP(eztools.LogWtTime, nil, conn.Protocol,
			conn.Addr, func(logFunc FuncLog,
				connFunc FuncConn, cnx net.Conn,
				addr [2]string) {
				go conn.Connected(logFunc, connFunc,
					cnx, [2]string{})
			}, conn.chanErrs)
		if err != nil {
			eztools.LogWtTime(conn.Name, "failed to listen",
				conn.Protocol, conn.Addr)
			// TODO: how abt svr.chanStrus
			//svr.lock.Unlock()
			return
		}
		conn.lstnr = lstnr
		conn.Addr = lstnr.Addr().String()
	}
	if eztools.Verbose > 0 {
		eztools.LogWtTime("server", conn.Name,
			"local", conn.Addr)
	}
	//svr.lock.Unlock()
	for i := 0; i < cap(conn.chanStrs); i++ {
		conn.chanStrs <- conn.Addr
	}
}

// RunFlow runs a read flow structure
// Return value: whether succeeded
func RunFlow(flow FlowStruc) bool {
	if len(flow.Conns) < 1 {
		eztools.Log("NO server defined. NO flow runs.")
		return false
	}
	flow.Vals = make(map[string]*FlowStepStruc, 0)
	for i := range flow.Conns {
		flow.Conns[i].ParsePeer(flow)
	}

	if eztools.Verbose > 0 {
		eztools.LogWtTime("flow begins", flow)
	}
	for i := range flow.Conns {
		if flow.Conns[i].chanErrs == nil {
			flow.Conns[i].chanErrs = make(chan error, 1)
		}
		defer func() {
			if flow.Conns[i].chanStrs != nil {
				close(flow.Conns[i].chanStrs)
				flow.Conns[i].chanStrs = nil
			}
			if flow.Conns[i].chanErrs != nil {
				close(flow.Conns[i].chanErrs)
				flow.Conns[i].chanErrs = nil
			}
		}()
		if flow.Conns[i].Block {
			flow.Conns[i].Run(&flow)
		} else {
			go flow.Conns[i].Run(&flow)
		}
	}
	for _, conn := range flow.Conns {
		if eztools.Verbose > 2 {
			eztools.Log("waiting for", conn.Name, "to end")
		}
		<-conn.chanErrs
	}
	if eztools.Verbose > 0 {
		eztools.LogWtTime("flow ends")
	}
	return true
}

// ReadFlowReader reads flow from a reader
func ReadFlowReader(rdr io.ReadCloser) (flow FlowStruc, err error) {
	bytes, err := io.ReadAll(rdr)
	rdr.Close()
	//n, err = uri.Read(bytes)
	if err != nil {
		//eztools.Log("read flow file", err)
		return
	}
	err = xml.Unmarshal(bytes, &flow)
	if err != nil {
		//eztools.Log("parse flow file", err)
		return
	}
	return
}

// ReadFlowFile reads flow from a file
func ReadFlowFile(file string) (flow FlowStruc, err error) {
	if err = eztools.XMLRead(file, &flow); err != nil {
		//eztools.Log(file, "failed to be read/parsed", err)
		return
	}
	return
}
