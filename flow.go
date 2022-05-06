package main

import (
	"encoding/xml"
	"errors"
	"io"
	"net"
	"os"
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

	FlowRcvLen = 256
	FlowComLen = 99
	FlowFilLen = 1024 * 1024
)
const (
	FlowChnLst = iota
	FlowChnMax

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

	Act   string `xml:"action"`
	Name  string `xml:"name"`
	Dest  string `xml:"dest"`
	Data  string `xml:"data"`
	Loop  int    `xml:"loop"`
	Block bool   `xml:"block"`
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
}

// ParseVar parses a string of a simple string or
//   <FlowVarSign>[<FlowConnStruc.Name><FlowVarSep>]<string><FlowVarSign>
//   fun() is invoked for matched FlowConnStruc,
//	with index of it in FlowStruc.Conns and <string>
//   A simple string and false are returned.
//   Otherwise, <string> and true are returned.
//   Both <FlowVarSign>'s are stripped before returned.
func (flow FlowStruc) ParseVar(str string,
	fun func(int, string)) (string, bool) {
	if len(str) < 1 {
		return "", false
	}
	parts := strings.Split(str, FlowVarSign)
	if len(parts) != 3 {
		//eztools.LogWtTime("unrecognized var", str)
		return str, false
	}
	vars := strings.Split(parts[1], FlowVarSep)
	if len(vars) != 2 {
		return parts[1], true
	}
	if fun != nil {
		for i := range flow.Conns {
			if vars[0] != flow.Conns[i].Name {
				continue
			}
			fun(i, vars[1])
		}
	}
	return parts[1], true
}

func (step FlowStepStruc) ParseDest(flow FlowStruc, conn FlowConnStruc) *net.UDPAddr {
	var ret string
	retWhole, ok := flow.ParseVar(step.Dest, func(svrInd int, varStr string) {
		switch varStr {
		case FlowVarLcl:
			ret = flow.Conns[svrInd].Addr
		}
	})
	if len(ret) < 1 {
		if !ok {
			ret = retWhole
		} else {
			switch retWhole {
			case FlowVarPee:
				ret = conn.Peer
			}
		}
	}
	addr, err := net.ResolveUDPAddr(conn.Protocol, ret)
	if err != nil {
		return nil
	}
	return addr
}

func (conn FlowConnStruc) Step1(flow FlowStruc, step FlowStepStruc) {
	respChn := make(chan FlowCommStruc, FlowComLen)
	var respStruc FlowCommStruc
	switch step.Act {
	case FlowActSnd:
		dest := step.ParseDest(flow, conn)
		if dest == nil && eztools.Verbose > 1 {
			eztools.LogWtTime("NO dest parsed for", step.Act, "as", step.Dest)
		}
		conn.chanComm <- FlowCommStruc{
			Act:  FlowChnSnd,
			Peer: dest,
			Data: step.Data,
			Resp: respChn,
		}
	case FlowActRcv:
		conn.chanComm <- FlowCommStruc{
			Act:  FlowChnRcv,
			Resp: respChn,
		}
	}
	respStruc = <-respChn
	if respStruc.Err != nil {
		eztools.LogWtTime(conn.Name, step.Act, respStruc.Err)
		return
	}
	conn.StepAll(flow, step.Steps)
}

func (conn FlowConnStruc) StepAll(flow FlowStruc, steps []FlowStepStruc) {
	for _, s := range steps {
		if s.Block {
			conn.Step1(flow, s)
		} else {
			go conn.Step1(flow, s)
		}
		//svr.curr = i + 1
	}
}

func (connStruc *FlowConnStruc) Connected(connTcp net.Conn) {
	// TODO: when to close this connUdp?
	com := <-connStruc.chanComm
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
			ln, addr, err := connUdp.ReadFromUDP(buf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					eztools.LogWtTime(connUdp.LocalAddr().String(), "reading", err)
					com.Err = err
					return err
				}
			}
			if err := fun(buf, ln); err != nil {
				return err
			}
			if com.Peer == nil {
				com.Peer = addr
			}
			return err
		}
	}
	switch com.Act {
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
		// com.Data and com.Peer must be empty
		bytes := make([]byte, FlowRcvLen)
		for {
			err := rcvFunc(bytes, func(buf []byte, ln int) error {
				com.Data += string(buf[:ln])
				return nil
			})
			if err != nil {
				break
			}
		}
	case FlowChnRcvFil:
		fil, err := os.OpenFile(com.Data, os.O_WRONLY|os.O_TRUNC, eztools.FileCreatePermission)
		if err != nil {
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
			if err != nil {
				break
			}
		}
	}
	com.Resp <- com
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
	eztools.LogWtTime(conn.Name, str+" on server", nm)
}

func (conn *FlowConnStruc) Run(flow FlowStruc) {
	if conn.chanComm == nil {
		conn.chanComm = make(chan FlowCommStruc, FlowComLen)
	}
	if len(conn.Peer) > 1 {
		conn.RunCln(flow)
	} else {
		conn.RunSvr(flow)
	}
	conn.StepAll(flow, conn.Steps)
}

func (cln *FlowConnStruc) RunCln(flow FlowStruc) {
	cln.Peer = cln.Wait4(flow)
	if strings.HasPrefix(cln.Protocol, "udp") {
		conn, err := ListenUdp(cln.Protocol, cln.Peer)
		if err != nil {
			eztools.LogWtTime(cln.Name, "failed to connect to", cln.Peer)
			return
		}
		cln.conn = conn
		go func() {
			cln.Connected(nil)
		}()
	} else {
		_, err := Client(cln.Protocol, cln.Peer, cln.Connected)
		if err != nil {
			eztools.LogWtTime(cln.Name, "failed to connect to", cln.Peer)
			return
		}
	}
}

// RunSvr supports TCP & UDP only. TODO: IP & Unix
func (svr *FlowConnStruc) RunSvr(flow FlowStruc) {
	//svr.lock.Lock()
	if strings.HasPrefix(svr.Protocol, "udp") {
		var err error
		svr.conn, err = ListenUdp(svr.Protocol, svr.Addr)
		if err != nil {
			eztools.LogWtTime(svr.Name, "failed to listen")
			return
		}
		svr.Addr = svr.conn.LocalAddr().String()
		/* TODO: go func() {
			defer func() {
				svr.conn.Close()
			}()
		}()*/
	} else {
		if svr.chanErrs == nil {
			svr.chanErrs = make(chan error, 1)
		}
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
	if eztools.Verbose > 1 {
		eztools.LogWtTime("listening", svr.Addr)
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
	for i := range flow.Conns {
		//flow.Servers[i].flow = &flow
		flow.Conns[i].ParsePeer(flow)
	}

	eztools.LogWtTime("flow begins")
	for i := range flow.Conns {
		if flow.Conns[i].Block {
			flow.Conns[i].Run(flow)
		} else {
			go flow.Conns[i].Run(flow)
		}
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
