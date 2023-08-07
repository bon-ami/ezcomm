package ezcomm

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v6"
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
	if len(*TstRmt) > 0 || len(*TstMsg) > 0 {
		t.Skip("rmt & msg are ignored. msgCount and quan count.")
	}
	AntiFlood.Limit = -1
	//*prot = "tcp"
	tstSvrChan = make(chan struct{}, 1)
	tstClntRdMsg = make(chan struct{})
	clnts := *TstClntNo
	go func() {
		TestServer(t)
	}()
	<-tstClntRdMsg // server ready
	chs := make([]chan struct{}, clnts)
	for i := range chs {
		chs[i] = make(chan struct{}, 1)
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
	*TstMsg = TstBye
	if eztools.Verbose > 1 {
		t.Log("client", *TstMsg)
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
			TstT.Error(tstErrPre, "flooding?")
			return
		}
	}
	id := tstClntID
	msg := *TstMsg
	chn2Main := tstChnClnt
	cnt := *TstMsgCount
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
			TstT.Log(ids, addrStrs[i], "=", addr[i])
		}
	}
	var msgs []string
	if len(msg) < 1 {
		const (
			CharHd = '\u0393'
			CharTl = '\u03C9'
		)
		msgs = append(msgs, TstHalo)
		if cnt <= 0 {
			cnt = CharTl - CharHd
			if eztools.Verbose > 0 {
				TstT.Log("tail char =", CharTl, "head char =", CharHd, "count =", cnt)
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
			TstT.Log(ids, "sending", msg1)
		}
		chn[0] <- RoutCommStruc{
			Act:  FlowChnSnd,
			Data: []byte(msg1),
		}
		for {
			select {
			case comm := <-chn[1]:
				if comm.Err != nil {
					TstT.Error(tstErrPre, ids, comm.Err)
					break ClientLoop
				}
				switch comm.Act {
				case FlowChnRcv:
					if string(comm.Data) != msg1 {
						TstT.Error(tstErrPre, ids, "sent", msg1, "!=",
							"got", string(comm.Data))
						break ClientLoop
					}
					continue ClientLoop
				case FlowChnSnd:
					if eztools.Verbose > 1 {
						TstT.Log(ids, "sent", comm.Data)
					}
				default:
					TstT.Log(ids, "got action?", comm.Act)
				}
			case <-time.After(TstTO):
				TstT.Log(ids, "client TO")
				TstT.Error(tstErrPre, ids, "reply NOT got from server!", ids)
				break ClientLoop
			}
		}
	}
	return
}

// TestClient uses Connected1Peer
func TestClient(t *testing.T) {
	Init4Tests(t)
	AntiFlood.Limit = -1
	if len(*TstRmt) < 1 {
		t.Skip("rmt needed")
	}
	tstChnClnt = make(chan struct{}, 2)
	chnFromClnt := tstChnClnt
	id := tstClntID
	conn, err := Client(t.Log, tstClnt,
		*TstProt, *TstRmt, Connected1Peer)
	if err != nil {
		t.Error(tstErrPre, err)
		return
	}
	if conn == nil {
		panic(conn)
	}
	// wait for init
	<-chnFromClnt
	if tstClntRdMsg != nil {
		// init done (for clntId)
		tstClntRdMsg <- struct{}{}
	}
	// wait for end
	<-chnFromClnt
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
			TstT.Log("exiting 1 connection")
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
			TstT.Log("exit 1 connection")
		}
	}()
	for {
		comm := <-chn[1]
		if comm.Err != nil {
			if comm.Err == eztools.ErrAccess {
				TstT.Error(tstErrPre, comm.Err)
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
			if string(comm.Data) == TstBye {
				if eztools.Verbose > 1 {
					TstT.Log("exiting server after bye sent")
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
				TstT.Log("server echoing",
					string(comm.Data), "to", peerAddr)
			}
			comm.Act = FlowChnSnd
			chn[0] <- comm
		}
	}
}

// TestServer uses ConnectedUDP or Connected1Peer
func TestServer(t *testing.T) {
	Init4Tests(t)
	var (
		conn  *net.UDPConn
		lstnr net.Listener
		err   error
	)
	AntiFlood.Limit = -1

	id := tstClntID
	if eztools.Verbose > 0 {
		t.Log("server", id, "will TO in", TstTO)
	}
	var chnEnd chan error
	if strings.HasPrefix(*TstProt, "udp") {
		for i := range tstChnSvr {
			tstChnSvr[i] = make(chan bool, 1)
		}
		conn, err = ListenUDP(*TstProt, *TstLcl)
		if err != nil {
			t.Error(tstErrPre, err)
			return
		}
		var chn [2]chan RoutCommStruc
		for i := range chn {
			chn[i] = make(chan RoutCommStruc, *TstClntNo)
		}
		go ConnectedUDP(t.Log, chn, conn)
		tstChnSvr[0] <- true
		go tstUDPSvr(tstChnSvr[1], chn)
	} else {
		for i := range tstChnSvr {
			tstChnSvr[i] = make(chan bool, tstDefSimulClients)
		}
		chnEnd = make(chan error)
		lstnr, err = ListenTCP(t.Log, tstTCPSvr, *TstProt, *TstLcl, Connected1Peer, chnEnd)
	}
	if err != nil {
		t.Error(tstErrPre, err)
		return
	}
	const lstn = "listening"
	switch {
	case lstnr != nil:
		if eztools.Verbose > 0 {
			t.Log(lstn, *TstProt, lstnr.Addr())
		}
		*TstRmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		if eztools.Verbose > 0 {
			t.Log(lstn, *TstProt, conn.LocalAddr())
		}
		*TstRmt = conn.LocalAddr().String() // to test clients
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
			t.Log("server waiting for end for", TstTO)
		}
		select {
		case <-time.After(TstTO):
			if !everConn {
				t.Skip("server TO")
			} else {
				t.Log("server TO")
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
