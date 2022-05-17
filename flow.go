package main

import (
	"encoding/xml"
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"strings"

	"gitee.com/bon-ami/eztools/v4"
)

/*const (
	EZCOMM_TYPE_UDP = iota
	EZCOMM_TYPE_TCP
)

type EzComm struct {
	tp   int
	addr string
	port int
}
*/

const (
	FlowActRcv = "receive"
	FlowActSnd = "send"

	FlowVarSign = "%"
	FlowVarSep  = "."
	FlowVarSvr  = "svr"
	FlowVarLst  = "listen"
	FlowVarLcl  = "local"
	FlowVarPee  = "peer"
	FlowVarFil  = "file"

	FlowRcvLen = 256
	FlowComLen = 99
	FlowFilLen = 1024 * 1024
)
const (
	FlowChnLst = iota
	FlowChnMax

	FlowChnEnd
	FlowChnSnd
	FlowChnSndFil
	FlowChnRcv
	FlowChnRcvFil
)

type FlowStepStruc struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt string `xml:",chardata"`

	Act string `xml:"action"`
	// Name if not null, this structure will be mapped to Vals
	Name string `xml:"name"`
	// Dest will be updated upon UDP receive action, if it is a variable
	Dest string `xml:"dest"`
	Data string `xml:"data"`
	// Loop rounds to repeat this step
	// 0, 1: no loop
	// > 1: number of rounds
	// < 0: infinitely
	Loop  int  `xml:"loop"`
	Block bool `xml:"block"`
	// Steps: sub steps triggered
	Steps []FlowStepStruc `xml:"step"`
	// curr: current sub step
	curr int
}

type FlowCommStruc struct {
	// Act=FlowChn* (>FlowChnMax)
	Act  int
	Peer *net.UDPAddr
	Data string
	Resp chan FlowCommStruc
	Err  error
}

type FlowConnStruc struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt string `xml:",chardata"`

	Name     string `xml:"name"`
	Protocol string `xml:"protocol"`
	Addr     string `xml:"address"`
	Peer     string `xml:"peer"`
	Block    bool   `xml:"block"`
	// TODO: use Wait?
	Wait  string          `xml:"wait"`
	Steps []FlowStepStruc `xml:"step"`
	// curr: current step
	//curr  int
	lstnr net.Listener
	conn  *net.UDPConn
	// chanErrs is for Server()
	chanErrs chan error
	chanStrs [FlowChnMax]chan string
	chanComm chan FlowCommStruc
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

const (
	// FlowParseValSimple is a sstring without <FlowVarSign>
	FlowParseValSimple = iota
	// FlowParseValSign is <FlowVarSign><string><FlowVarSign>
	FlowParseValSign
	// FlowParseValVar is <FlowVarSign><xml tag name of FlowConnStruc or FlowStepStruc><FlowVarSep><string><FlowVarSign>
	FlowParseValVar
)

// ParseVar parses a string of a simple string or
//   <FlowVarSign>[<xml tag name of FlowConnStruc or FlowStepStruc><FlowVarSep>]<string><FlowVarSign>
//   fun() is invoked for matched FlowConnStruc,
//	with index of it in FlowStruc.Conns and <string>
// Return values:
//  1st.
//   The simple string
//   If FlowConnStruc is matched, the value of its member whose xml tag is <string>
//   Otherwise, <string>
//  2nd. FlowParseVal*
func (flow FlowStruc) ParseVar(str string,
	fun func(int, string)) (string, int) {
	if len(str) < 1 {
		return str, FlowParseValSimple
	}
	parts := strings.Split(str, FlowVarSign)
	if len(parts) != 3 {
		//eztools.LogWtTime("unrecognized var", str)
		return str, FlowParseValSimple
	}
	vars := strings.Split(parts[1], FlowVarSep)
	if len(vars) != 2 {
		return parts[1], FlowParseValSign
	}
	if fun != nil {
		for i := range flow.Conns {
			if vars[0] != flow.Conns[i].Name {
				continue
			}
			fun(i, vars[1])
		}
	}
	if step1, ok := flow.Vals[vars[0]]; ok && step1 != nil {
		//eztools.Log("parsevar", vars[0], vars[1], *step1)
		stepTp := reflect.TypeOf(*step1)
		for i := 0; i < stepTp.NumField(); i++ {
			if vars[1] != stepTp.Field(i).Tag.Get("xml") {
				//eztools.Log("passing", stepTp.Field(i).Tag.Get("xml"))
				continue
			}
			f := stepTp.FieldByIndex(stepTp.Field(i).Index)
			v := reflect.ValueOf(*step1).FieldByName(f.Name)
			//eztools.Log("got", v.String())
			/*str, ok := v.(string)
			if !ok {
				break
			}*/
			return v.String(), FlowParseValVar
		}
	}

	return parts[1], FlowParseValSign
}

// ParseData parses data in step
// Return values:
//  1st. is one of following
//    data string in form of a simple string
//    the value of a member of FlowConnStruc or FlowStepStruc for <string> in <FlowVarSign><xml tag name of FlowConnStruc or FlowStepStruc><FlowVarSep><string><FlowVarSign>
//    file name for <string> in <FlowVarSign><FlowVarFil><FlowVarSep><string><FlowVarSign>
//  2nd.
//   1: a file name
//   0: a string
func (step FlowStepStruc) ParseData(flow FlowStruc, conn FlowConnStruc) (string, int) {
	retWhole, parseRes := flow.ParseVar(step.Data, nil) // TODO: match a server
	switch parseRes {
	case FlowParseValSign:
		switch {
		case strings.HasPrefix(retWhole, FlowVarFil+FlowVarSep):
			return strings.TrimPrefix(retWhole, FlowVarFil+FlowVarSep), 1
		}
	}
	return retWhole, 0
}

func (step FlowStepStruc) ParseDest(flow FlowStruc, conn FlowConnStruc) *net.UDPAddr {
	var ret string
	retWhole, parseRes := flow.ParseVar(step.Dest, func(svrInd int, varStr string) {
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

func (conn FlowConnStruc) Step1(flow *FlowStruc, step *FlowStepStruc) {
	for {
		if eztools.Verbose > 2 {
			eztools.LogWtTime("step", step.Name, "action", step.Act)
		}
		respChn := make(chan FlowCommStruc, FlowComLen)
		var respStruc FlowCommStruc
		switch step.Act {
		case FlowActSnd:
			dest := step.ParseDest(*flow, conn)
			if dest == nil && eztools.Verbose > 1 {
				eztools.LogWtTime("NO dest parsed for", step.Act, "as", step.Dest)
			}
			data, fil := step.ParseData(*flow, conn)
			conn.chanComm <- FlowCommStruc{
				Act:  FlowChnSnd + fil,
				Peer: dest,
				Data: data,
				Resp: respChn,
			}
		case FlowActRcv:
			data, fil := step.ParseData(*flow, conn)
			conn.chanComm <- FlowCommStruc{
				Act:  FlowChnRcv + fil,
				Data: data,
				Resp: respChn,
			}
		}
		respStruc = <-respChn
		if respStruc.Err != nil {
			eztools.LogWtTime(conn.Name, step.Act, respStruc.Err)
		} else {
			step.Data = respStruc.Data
			if respStruc.Peer != nil {
				if eztools.Verbose > 1 {
					eztools.Log("refreshing dest of", step.Name,
						"from", step.Dest, "to", respStruc.Peer.String())
				}
				step.Dest = respStruc.Peer.String()
			}
			if len(step.Name) > 0 {
				flow.Vals[step.Name] = step
			}
			if eztools.Verbose > 2 {
				eztools.LogWtTime("step done", step.Name, "action", step.Act,
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

func (connStruc *FlowConnStruc) Connected(connTcp net.Conn) {
	defer func() {
		// duplcate for TCP server, but does not matter
		connStruc.chanErrs <- eztools.ErrAbort
	}()
	for {
		if eztools.Verbose > 2 {
			eztools.LogWtTime(connStruc.Name, "waiting")
		}
		// TODO: when to close this connUdp?
		com := <-connStruc.chanComm
		if eztools.Verbose > 2 {
			eztools.LogWtTime(connStruc.Name, "got command", com)
		}
		connUdp := connStruc.conn
		var (
			sndFunc func([]byte) error
			rcvFunc func([]byte, func([]byte, int) error) error
		)
		if connTcp != nil {
			defer connTcp.Close()
			sndFunc = func(buf []byte) (err error) {
				_, err = connTcp.Write(buf)
				return
			}
			rcvFunc = func(buf []byte, fun func([]byte, int) error) error {
				ln, err := connTcp.Read(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						eztools.LogWtTime(connTcp.LocalAddr().String(), "reading", err)
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
			if connUdp == nil {
				eztools.LogFatal("NO connections connected")
				return
			}
			defer connUdp.Close()
			sndFunc = func(buf []byte) (err error) {
				if com.Peer != nil {
					_, err = connUdp.WriteTo(buf, com.Peer)
				} else { // TODO: default to current peer?
					_, err = connUdp.Write(buf)
				}
				return err
			}
			rcvFunc = func(buf []byte, fun func([]byte, int) error) error {
				//eztools.LogPrintWtTime(connStruc.Name, "to recv", connUdp.LocalAddr().String())
				ln, addr, err := connUdp.ReadFromUDP(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						eztools.LogWtTime(connStruc.Name, connUdp.LocalAddr().String(), "reading error", err)
						com.Err = err
						return err
					}
				}
				//eztools.LogPrintWtTime(connStruc.Name, "recv", err)
				if com.Peer == nil {
					com.Peer = addr
					//eztools.Log("setting peer", com.Peer)
				} else {
					eztools.Log("peer not null", com.Peer)
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
				eztools.LogWtTime(connStruc.Name, "ending")
			}
			return
		case FlowChnSnd:
			com.Err = sndFunc([]byte(com.Data))
		case FlowChnSndFil:
			buf := make([]byte, FlowFilLen)
			fil, err := os.Open(com.Data)
			if err != nil {
				eztools.LogWtTime(connUdp.LocalAddr().String, "reading", err)
				com.Err = err
				break
			}
			defer fil.Close()
			// TODO: how to tell of pieces on peer?
			for {
				var ln int
				ln, err = fil.Read(buf)
				if err != nil {
					if !errors.Is(err, io.EOF) {
						eztools.LogWtTime(connUdp.LocalAddr().String, "reading", err)
						com.Err = err
					}
					break
				}
				err = sndFunc([]byte(buf[:ln]))
				if err != nil {
					com.Err = err
					break
				}
			}
		case FlowChnRcv:
			// com.Peer must be empty
			// com.Data is appended
			var byteLen int
			bytes := make([]byte, FlowRcvLen)
			for {
				err := rcvFunc(bytes, func(buf []byte, ln int) error {
					com.Data += string(buf[:ln])
					byteLen = ln
					return nil
				})
				if eztools.Verbose > 2 {
					eztools.LogWtTime(connStruc.Name,
						"received", err, string(bytes[:byteLen]))
				}
				//if err != nil {
				break
				//}
			}
		case FlowChnRcvFil:
			fil, err := os.OpenFile(com.Data, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, eztools.FileCreatePermission)
			if err != nil {
				com.Err = err
				break
			}
			defer fil.Close()
			bytes := make([]byte, FlowFilLen)
			for {
				err := rcvFunc(bytes, func(buf []byte, ln int) error {
					if _, err := fil.Write(buf[:ln]); err != nil {
						eztools.LogWtTime("failed to write to", com.Data, err)
						com.Err = err
						return err
					}
					return nil
				})
				if eztools.Verbose > 2 {
					eztools.LogWtTime(connStruc.Name,
						"saved", err, com.Data)
				}
				//if err != nil {
				break
				//}
			}
		}
		if eztools.Verbose > 2 {
			eztools.LogWtTime(connStruc.Name, "replying", com)
		}
		com.Resp <- com
	}
}

func (conn *FlowConnStruc) ParsePeer(flow FlowStruc) {
	//flow := svr.flow
	flow.ParseVar(conn.Peer, func(svrInd int, varStr string) {
		switch varStr {
		case FlowVarLst:
			conn.wait4Svr = &flow.Conns[svrInd]
			conn.wait4Act = FlowVarLst
			if flow.Conns[svrInd].chanStrs[FlowChnLst] == nil {
				flow.Conns[svrInd].chanStrs[FlowChnLst] = make(chan string, 1)
			} else {
				curr := cap(flow.Conns[svrInd].chanStrs[FlowChnLst])
				flow.Conns[svrInd].chanStrs[FlowChnLst] = make(chan string, curr+1)
			}
			break
		}
	})
}

func (conn *FlowConnStruc) Wait4(flow FlowStruc) (ret string) {
	if conn.wait4Svr == nil {
		return
	}
	conn.LockLog(conn.wait4Svr.Name, true)
	ret = <-conn.wait4Svr.chanStrs[FlowChnLst]
	conn.LockLog(conn.wait4Svr.Name, false)
	return ret
}

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

func (conn *FlowConnStruc) Run(flow *FlowStruc) {
	if conn.chanComm == nil {
		conn.chanComm = make(chan FlowCommStruc, FlowComLen)
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
	conn.chanComm <- FlowCommStruc{Act: FlowChnEnd}
}

func (cln *FlowConnStruc) RunCln(flow FlowStruc) {
	if eztools.Verbose > 2 {
		eztools.LogWtTime("client", cln.Name, cln.Protocol)
	}
	cln.Peer = cln.Wait4(flow)
	parts := strings.Split(cln.Peer, ":")
	cln.Peer = "localhost:" + parts[len(parts)-1]
	if eztools.Verbose > 1 {
		eztools.LogWtTime("client", cln.Name, cln.Protocol, cln.Addr, "->", cln.Peer)
	}
	if strings.HasPrefix(cln.Protocol, "udp") {
		conn, err := ListenUdp(cln.Protocol, cln.Addr)
		if err != nil {
			eztools.LogWtTime(cln.Name, "failed to connect to", cln.Peer)
			return
		}
		cln.conn = conn
		if eztools.Verbose > 0 {
			eztools.LogWtTime("client", cln.Name,
				"local", conn.LocalAddr().String())
		}
		go func() {
			cln.Connected(nil)
		}()
	} else {
		conn, err := Client(cln.Protocol, cln.Peer, cln.Connected)
		if err != nil {
			eztools.LogWtTime(cln.Name, "failed to connect to", cln.Peer)
			return
		}
		if eztools.Verbose > 0 {
			eztools.LogWtTime("client", cln.Name,
				"local", conn.LocalAddr().String(),
				"remote", conn.RemoteAddr().String())
		}
	}
}

// RunSvr supports TCP & UDP only. TODO: IP & Unix
func (svr *FlowConnStruc) RunSvr(flow FlowStruc) {
	if len(svr.Protocol) < 1 {
		return
	}
	if eztools.Verbose > 2 {
		eztools.LogWtTime("server", svr.Name, svr.Protocol)
	}
	//svr.lock.Lock()
	if strings.HasPrefix(svr.Protocol, "udp") {
		var err error
		svr.conn, err = ListenUdp(svr.Protocol, svr.Addr)
		if err != nil {
			eztools.LogWtTime(svr.Name, "failed to listen")
			return
		}
		svr.Addr = svr.conn.LocalAddr().String()
		go func() {
			/*defer func() {
				svr.conn.Close()
			}()*/
			svr.Connected(nil)
		}()
	} else {
		lstnr, err := ListenTcp(svr.Protocol,
			svr.Addr, func(conn net.Conn) {
				svr.Connected(conn)
			}, svr.chanErrs)
		if err != nil {
			eztools.LogWtTime(svr.Name, "failed to listen",
				svr.Protocol, svr.Addr)
			// TODO: how abt svr.chanStrus
			//svr.lock.Unlock()
			return
		}
		svr.lstnr = lstnr
		svr.Addr = lstnr.Addr().String()
	}
	if eztools.Verbose > 0 {
		eztools.LogWtTime("server", svr.Name,
			"local", svr.Addr)
	}
	//svr.lock.Unlock()
	listeners := cap(svr.chanStrs[FlowChnLst])
	for i := 0; i < listeners; i++ {
		svr.chanStrs[FlowChnLst] <- svr.Addr
	}
}

func runFlow(flow FlowStruc) bool {
	if len(flow.Conns) < 1 {
		eztools.LogPrint("NO server defined. NO flow runs.")
		return false
	}
	flow.Vals = make(map[string]*FlowStepStruc, 0)
	for i := range flow.Conns {
		//flow.Servers[i].flow = &flow
		flow.Conns[i].ParsePeer(flow)
	}

	eztools.LogWtTime("flow begins")
	for i := range flow.Conns {
		if flow.Conns[i].chanErrs == nil {
			flow.Conns[i].chanErrs = make(chan error, 1)
		}
		if flow.Conns[i].Block {
			flow.Conns[i].Run(&flow)
		} else {
			go flow.Conns[i].Run(&flow)
		}
	}
	for _, conn := range flow.Conns {
		/*if eztools.Verbose > 2 {
			eztools.Log("waiting for", conn.Name, "to end")
		}*/
		<-conn.chanErrs
	}
	eztools.LogWtTime("flow ends")
	return true
}

func runFlowFile(file string) bool {
	var flow FlowStruc
	if err := eztools.XMLRead(file, &flow); err != nil {
		eztools.LogPrint(file, "failed to be read/parsed", err)
		return false
	}
	return runFlow(flow)
}
