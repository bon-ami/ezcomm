package ezcomm

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v6"
	"github.com/pkg/errors"
)

var (
	tstClntRdMsg, tstSvrChan chan struct{}
	// tstChnClnt is for both start and stop
	tstChnClnt chan struct{}
	// tstChnSvr for start and stop
	tstChnSvr  [2]chan bool
	tstClntSvr bool
)

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
	defer close(tstSvrChan)
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
	for _, c := range chs {
		//t.Log("waiting for", c)
		<-c
		close(c)
	}
	*TstMsg = TstBye
	if *tstVerbose > 1 {
		t.Log("client", *TstMsg)
	}
	close(tstClntRdMsg)
	tstClntRdMsg = nil
	TestClient(t)
	if tstSvrChan != nil {
		<-tstSvrChan
	}
}

func tstPrintAddr(ids string, addr [4]string) {
	for i, v := range []string{
		"parsed local", "remote",
		"requested protocol",
		"requested address"} {
		if *tstVerbose > 0 {
			TstT.Log(ids, v, "=", addr[i])
		}
	}
}

func tstClnt(addr [4]string, chn [2]chan RoutCommStruc) {
	for _, chn1 := range chn {
		if chn1 == nil {
			TstT.Error("flooding?")
			return
		}
		defer close(chn1)
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
	tstPrintAddr(ids, addr)

	var msgs []string
	if len(msg) < 1 {
		const (
			CharHd = '\u0393'
			CharTl = '\u03C9'
		)
		msgs = append(msgs, TstHalo)
		if cnt <= 0 {
			cnt = CharTl - CharHd
			if *tstVerbose > 0 {
				TstT.Log(ids, "tail char =", CharTl, "head char =", CharHd, "count =", cnt)
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
		if *tstVerbose > 1 {
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
					TstT.Error(errors.Wrap(comm.Err, ids))
					break ClientLoop
				}
				switch comm.Act {
				case FlowChnRcv:
					if string(comm.Data) != msg1 {
						TstT.Error(ids, "sent", msg1, "!=",
							"got", string(comm.Data))
						break ClientLoop
					}
					continue ClientLoop
				case FlowChnSnd:
					if *tstVerbose > 1 {
						TstT.Log(ids, "sent", comm.Data)
					}
				default:
					TstT.Log(ids, "got action?", comm.Act)
				}
			case <-time.After(TstTO):
				TstT.Error(ids, "reply NOT got from server!")
				break ClientLoop
			}
		}
	}
	chn[0] <- RoutCommStruc{Act: FlowChnEnd}
	// wait for End
	<-chn[1]
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
		t.Error(err)
		return
	}
	if conn == nil {
		panic("no connections created")
	}
	// wait for init
	<-chnFromClnt
	if tstClntRdMsg != nil {
		// init done (for clntId)
		tstClntRdMsg <- struct{}{}
	}
	// wait for end
	<-chnFromClnt
	close(chnFromClnt)
	eztools.Debugging = false // do not let conn routines to print since now
	if *tstVerbose > 1 {
		t.Log("client died", id)
	}
}

func tstTCPSvr(addr [4]string, chn [2]chan RoutCommStruc) {
	tstPrintAddr("TCP server", addr)
	tstChnSvr[0] <- true
	go tstUDPSvr(tstChnSvr[1], chn)
}

func tstUDPSvr(fin chan bool, chn [2]chan RoutCommStruc) {
	done := false
	if chn[0] == nil || chn[1] == nil {
		panic("nil chn")
	}
	wait4Rcvr := true
	defer func() {
		if *tstVerbose > 1 {
			TstT.Log("exiting 1 connection", fin, chn)
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
		for _, chn1 := range chn {
			close(chn1)
		}
		if fin != nil {
			fin <- done
		}
		if *tstVerbose > 1 {
			TstT.Log("exit 1 connection", fin, chn)
		}
	}()
	for {
		comm := <-chn[1]
		switch comm.Act {
		case FlowChnEnd:
			if comm.Err != nil {
				//such as eztools.ErrAccess or eztools.ErrIncomplete
				TstT.Log(comm.Err)
			}
			wait4Rcvr = false
			return
		case FlowChnSnd:
			if comm.Err != nil {
				return
			}
			if string(comm.Data) == TstBye {
				if *tstVerbose > 1 {
					TstT.Log("exiting server after bye sent", fin, chn)
				}
				done = true
				return
			}
		case FlowChnRcv: // no comm.Err check
			if *tstVerbose > 0 {
				var peerAddr string
				if comm.PeerUDP != nil {
					peerAddr = comm.PeerUDP.String()
				} else {
					peerAddr = comm.PeerTCP.String()
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
	if *tstVerbose > 0 {
		t.Log("server", id, "will TO in", TstTO)
	}
	var chnEnd chan error
	if strings.HasPrefix(*TstProt, "udp") {
		for i := range tstChnSvr {
			tstChnSvr[i] = make(chan bool, 1)
			defer close(tstChnSvr[i])
		}
		conn, err = ListenUDP(*TstProt, *TstLcl)
		if err != nil {
			t.Error(err)
			return
		}
		var chn [2]chan RoutCommStruc
		for i := range chn {
			chn[i] = make(chan RoutCommStruc, *TstClntNo)
			defer close(chn[i])
		}
		go ConnectedUDP(t.Log, chn, conn)
		tstChnSvr[0] <- true
		go tstUDPSvr(tstChnSvr[1], chn)
	} else {
		for i := range tstChnSvr {
			tstChnSvr[i] = make(chan bool, tstDefSimulClients)
			defer close(tstChnSvr[i])
		}
		chnEnd = make(chan error)
		lstnr, err = ListenTCP(t.Log, tstTCPSvr, *TstProt, *TstLcl, Connected1Peer, chnEnd)
	}
	if err != nil {
		t.Error(err)
		return
	}
	const lstn = "listening"
	switch {
	case lstnr != nil:
		if *tstVerbose > 0 {
			t.Log(lstn, *TstProt, lstnr.Addr())
		}
		*TstRmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		if *tstVerbose > 0 {
			t.Log(lstn, *TstProt, conn.LocalAddr())
		}
		*TstRmt = conn.LocalAddr().String() // to test clients
	default:
		t.Error("no connection info")
	}
	if tstClntRdMsg != nil {
		tstClntRdMsg <- struct{}{}
	}
	// wait for 1 connection
	//<-tstChnSvr[0]
	var everConn bool
ServerLoop:
	for {
		if *tstVerbose > 0 {
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
			if *tstVerbose > 0 {
				t.Log("server got connected")
			}
			everConn = true
		case done := <-tstChnSvr[1]:
			if *tstVerbose > 0 {
				t.Log("server waited for 1 more end")
			}
			if done {
				break ServerLoop
			}
		}
	}
	if *tstVerbose > 0 {
		t.Log("server to die")
	}
	if lstnr != nil {
		lstnr.Close()
	}
	if conn != nil {
		conn.Close()
	}
	if chnEnd != nil { //TCP
		if *tstVerbose > 0 {
			t.Log("waiting for server to die")
		}
		<-chnEnd
	}
	eztools.Debugging = false // do not let conn routines to print since now
	if *tstVerbose > 0 {
		t.Log("server died")
	}
	if tstSvrChan != nil { //TCP
		tstSvrChan <- struct{}{}
	}
}
