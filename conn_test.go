package ezcomm

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v5"
)

var (
	// tstChnClnt is for both start and stop
	tstChnClnt chan struct{}
	// tstChnSvr for start and stop
	tstChnSvr  [2]chan bool
	tstClntSvr bool
)

const tstErrPre = "ERR:"

func TestSvrCln(t *testing.T) {
	tstClntSvr = true
	defer func() {
		tstClntSvr = false
	}()
	Init4Tests(t)
	defer Deinit4Tests()
	if len(*tstRmt) > 0 || len(*tstMsg) > 0 {
		t.Skip("rmt & msg are ignored. msgCount and quan count.")
	}
	AntiFlood.Limit = -1
	//*prot = "tcp"
	tstSvrChan = make(chan struct{})
	tstClntRdMsg = make(chan struct{})
	clnts := *tstClntNo
	go func() {
		TestServer(t)
	}()
	<-tstClntRdMsg // server ready
	chs := make([]chan struct{}, clnts)
	for i := range chs {
		chs[i] = make(chan struct{})
	}
	for i := 0; i < clnts; i++ {
		go func(ind int) {
			TestClient(t)
			//t.Log("done for", ind)
			chs[ind] <- struct{}{}
		}(i)
		<-tstClntRdMsg // next client available
	}
	for i := range chs {
		//t.Log("waiting for", i)
		<-chs[i]
	}
	*tstMsg = tstBye
	if eztools.Verbose > 1 {
		t.Log("client", *tstMsg)
	}
	tstClntRdMsg = nil
	TestClient(t)
	if tstSvrChan != nil {
		<-tstSvrChan
	}
}

func tstClnt(addr [4]string, chn [2]chan RoutCommStruc) {
	for _, chn1 := range chn {
		if chn1 == nil {
			tstT.Error(tstErrPre, "flooding?")
			return
		}
	}
	id := tstClntID
	msg := *tstMsg
	chn2Main := tstChnClnt
	cnt := *tstMsgCount
	chn2Main <- struct{}{}
	defer func() {
		chn2Main <- struct{}{}
	}()

	ids := strconv.Itoa(id)
	var addrStrs [4]string
	addrStrs[0] = "parsed local"
	addrStrs[1] = "remote"
	addrStrs[2] = "requested protocol"
	addrStrs[3] = "requested address"
	for i := range addrStrs {
		if eztools.Verbose > 0 {
			tstT.Log(ids, addrStrs[i], "=", addr[i])
		}
	}
	var msgs []string
	if len(msg) < 1 {
		const (
			CharHd = '\u0393'
			CharTl = '\u03C9'
		)
		msgs = append(msgs, "bonjour")
		if cnt <= 0 {
			cnt = CharTl - CharHd
			if eztools.Verbose > 0 {
				tstT.Log("tail char =", CharTl, "head char =", CharHd, "count =", cnt)
			}
		}
		cnt += CharHd
		for j := CharHd; j < rune(cnt); j++ {
			msgs = append(msgs, string(j)+ids)
		}
	} else {
		msgs = strings.Split(msg, tstSeparator)
	}
ClientLoop:
	for _, msg1 := range msgs {
		if eztools.Verbose > 1 {
			tstT.Log(ids, "sending", msg1)
		}
		chn[0] <- RoutCommStruc{
			Act:  FlowChnSnd,
			Data: []byte(msg1),
		}
		for {
			select {
			case comm := <-chn[1]:
				if comm.Err != nil {
					tstT.Error(tstErrPre, ids, comm.Err)
					break ClientLoop
				}
				switch comm.Act {
				case FlowChnRcv:
					if string(comm.Data) != msg1 {
						tstT.Error(tstErrPre, ids, "sent", msg1, "!=",
							"got", string(comm.Data))
						break ClientLoop
					}
					continue ClientLoop
				case FlowChnSnd:
					if eztools.Verbose > 1 {
						tstT.Log(ids, "sent", comm.Data)
					}
				default:
					tstT.Log(ids, "got action?", comm.Act)
				}
			case <-time.After(tstTO):
				tstT.Log(ids, "client TO")
				tstT.Error(tstErrPre, ids, "reply NOT got from server!", ids)
				break ClientLoop
			}
		}
	}
	return
}

// TestClient uses ConnectedTCP
func TestClient(t *testing.T) {
	Init4Tests(t)
	if !tstClntSvr {
		defer Deinit4Tests()
	}
	AntiFlood.Limit = -1
	if len(*tstRmt) < 1 {
		t.Skip("rmt needed")
	}
	tstChnClnt = make(chan struct{}, 2)
	id := tstClntID
	conn, err := Client(t.Log, tstClnt,
		*tstProt, *tstRmt, ConnectedTCP)
	if err != nil {
		t.Error(tstErrPre, err)
		return
	}
	if conn == nil {
		panic(conn)
	}
	// wait for init
	<-tstChnClnt
	if tstClntRdMsg != nil {
		// init done (for clntId)
		tstClntRdMsg <- struct{}{}
	}
	// wait for end
	<-tstChnClnt
	eztools.Debugging = false // do not let conn routines to print since now
	if eztools.Verbose > 1 {
		t.Log("client died", id)
	}
}

func tstTCPSvr(addr [4]string, chn [2]chan RoutCommStruc) {
	tstChnSvr[0] <- true
	go tstUDPSvr(tstChnSvr[1], chn)
}

func tstUDPSvr(fin chan bool, chn [2]chan RoutCommStruc) {
	wait4Rcvr := true
	done := false
	defer func() {
		if eztools.Verbose > 1 {
			tstT.Log("exiting 1 connection")
		}
		if wait4Rcvr {
			chn[0] <- RoutCommStruc{
				Act: FlowChnEnd,
			}
			for {
				comm := <-chn[1]
				if comm.Act == FlowChnEnd {
					break
				}
			}
		}
		if fin != nil {
			fin <- done
		}
		if eztools.Verbose > 1 {
			tstT.Log("exit 1 connection")
		}
	}()
	for {
		comm := <-chn[1]
		if comm.Err != nil {
			if comm.Err == eztools.ErrAccess {
				tstT.Error(tstErrPre, comm.Err)
			}
			wait4Rcvr = false
			return
		}
		switch comm.Act {
		case FlowChnEnd:
			//tstT.Error(tstErrPre, eztools.ErrIncomplete)
			wait4Rcvr = false
			return
		case FlowChnSnd:
			if string(comm.Data) == tstBye {
				if eztools.Verbose > 1 {
					tstT.Log("exiting server after bye sent")
				}
				done = true
				return
			}
		case FlowChnRcv:
			if eztools.Verbose > 0 {
				var peerAddr string
				if comm.PeerUDP != nil {
					peerAddr = comm.PeerUDP.String()
				}
				tstT.Log("server echoing",
					string(comm.Data), "to", peerAddr)
			}
			comm.Act = FlowChnSnd
			chn[0] <- comm
		}
	}
}

// TestServer uses ConnectedUDP or ConnectedTCP
func TestServer(t *testing.T) {
	Init4Tests(t)
	if !tstClntSvr {
		defer Deinit4Tests()
	}
	var (
		conn  *net.UDPConn
		lstnr net.Listener
		err   error
	)
	AntiFlood.Limit = -1

	id := tstClntID
	if eztools.Verbose > 0 {
		t.Log("server", id, "will TO in", tstTO)
	}
	for i := range tstChnSvr {
		tstChnSvr[i] = make(chan bool, TestDefSimulClients)
	}
	var chnEnd chan error
	if strings.HasPrefix(*tstProt, "udp") {
		conn, err = ListenUDP(*tstProt, *tstLcl)
		if err != nil {
			t.Error(tstErrPre, err)
			return
		}
		var chn [2]chan RoutCommStruc
		for i := range chn {
			chn[i] = make(chan RoutCommStruc, FlowComLen)
		}
		go ConnectedUDP(t.Log, chn, conn)
		tstChnSvr[0] <- true
		go tstUDPSvr(tstChnSvr[1], chn)
	} else {
		chnEnd = make(chan error)
		lstnr, err = ListenTCP(t.Log, tstTCPSvr, *tstProt, *tstLcl, ConnectedTCP, chnEnd)
	}
	if err != nil {
		t.Error(tstErrPre, err)
		return
	}
	const lstn = "listening"
	switch {
	case lstnr != nil:
		if eztools.Verbose > 0 {
			t.Log(lstn, *tstProt, lstnr.Addr())
		}
		*tstRmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		if eztools.Verbose > 0 {
			t.Log(lstn, *tstProt, conn.LocalAddr())
		}
		*tstRmt = conn.LocalAddr().String() // to test clients
	default:
		t.Error(tstErrPre, "no connection info")
	}
	if tstClntRdMsg != nil {
		tstClntRdMsg <- struct{}{}
	}
	// wait for 1 connection
	//<-tstChnSvr[0]
	var everConn bool
ServerLoop:
	for {
		if eztools.Verbose > 0 {
			t.Log("server waiting for end for", tstTO)
		}
		select {
		case <-time.After(tstTO):
			if !everConn {
				t.Skip("server TO")
			}
			break ServerLoop
		case <-tstChnSvr[0]:
			if eztools.Verbose > 0 {
				t.Log("server waiting for 1 more end")
			}
			everConn = true
		case done := <-tstChnSvr[1]:
			if eztools.Verbose > 0 {
				t.Log("server waited for 1 more end")
			}
			if done {
				break ServerLoop
			}
		}
	}
	if eztools.Verbose > 0 {
		t.Log("server to die")
	}
	if lstnr != nil {
		lstnr.Close()
	}
	if conn != nil {
		conn.Close()
	}
	if chnEnd != nil { //TCP
		if eztools.Verbose > 0 {
			t.Log("waiting for server to die")
		}
		<-chnEnd
	}
	eztools.Debugging = false // do not let conn routines to print since now
	if eztools.Verbose > 0 {
		t.Log("server died")
	}
	if tstSvrChan != nil { //TCP
		tstSvrChan <- struct{}{}
	}
}
