package ezcomm

import (
	"flag"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v4"
)

const (
	tstBye           = "ciao"
	tstSeparator     = ";"
	TestReadBufLen   = 10
	TestSimulClients = 10
	TestTimeout      = 30
)

var (
	tstInitDone                              bool
	tstProt, tstLcl, tstRmt, tstMsg, tstRoot *string
	tstClntNo, tstMsgCount                   *int
	tstClntRdMsg                             chan struct{}
	tstSvrChan                               chan struct{}
)

func init() {
	tstProt = flag.String("prot", "udp4", "protocol, tcp, tcp4, tcp6, unix or unixpacket. default is udp.")
	tstLcl = flag.String("lcl", DefAdr+":" /*"localhost:"*/, "local address")
	tstRmt = flag.String("rmt", "", "remote address")
	tstMsg = flag.String("msg", "", "messages to send, separated by \""+tstSeparator+"\". \""+tstBye+"\" to end udp server")
	tstRoot = flag.String("root", "", "root dir for http server")
	tstMsgCount = flag.Int("msgCount", 0, "number of messages to send per client for TestSvrCln. default=all greek")
	tstClntNo = flag.Int("quan", TestSimulClients, "quantity of clients. default="+strconv.Itoa(TestSimulClients))
}

func init4Tests(t *testing.T) {
	tstClntId++
	if tstInitDone {
		return
	}
	if t == nil {
		panic(t)
	}
	tstInitDone = true
	tstT = t
	if !flag.Parsed() {
		flag.Parse()
	}
	eztools.SetLogFunc(func(l ...any) {
		func(m ...any) {
			t.Log(m)
		}(eztools.GetCallerLog(), l)
	})
	eztools.Debugging = true
	eztools.Verbose = 3
	return
}

var (
	tstClntId int
	tstT      *testing.T
	// tstChnClnt is for both start and stop
	tstChnClnt chan struct{}
	// tstChnSvr for start and stop
	tstChnSvr [2]chan bool
)

func TestSvrCln(t *testing.T) {
	init4Tests(t)
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
	t.Log("client", *tstMsg)
	tstClntRdMsg = nil
	TestClient(t)
	if tstSvrChan != nil {
		<-tstSvrChan
	}
}

func tstClnt(addr [4]string, chn [2]chan RoutCommStruc) {
	id := tstClntId
	msg := *tstMsg
	chn2Main := tstChnClnt
	cnt := *tstMsgCount
	chn2Main <- struct{}{}

	defer func() {
		chn2Main <- struct{}{}
	}()
	tstT.Log(addr)
	ids := strconv.Itoa(id)
	var msgs []string
	if len(msg) < 1 {
		const (
			CharHd = '\u0393'
			CharTl = '\u03C9'
		)
		msgs = append(msgs, "bonjour")
		if cnt <= 0 {
			cnt = CharTl - CharHd
			tstT.Log(CharTl, CharHd, cnt)
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
		chn[0] <- RoutCommStruc{
			Act:  FlowChnSnd,
			Data: []byte(msg1),
		}
		for {
			select {
			case comm := <-chn[1]:
				if comm.Err != nil {
					tstT.Error(comm.Err)
					break ClientLoop
				}
				switch comm.Act {
				case FlowChnRcv:
					if string(comm.Data) != msg1 {
						tstT.Error("sent", msg1, "!=",
							"got", string(comm.Data))
						break ClientLoop
					}
					continue ClientLoop
				case FlowChnSnd:
					tstT.Log("sent", comm.Data)
				default:
					tstT.Log("got action?", comm.Act)
				}
			case <-time.After(time.Second * TestTimeout):
				tstT.Log("client TO")
				tstT.Error("reply NOT got from server!", ids)
				break ClientLoop
			}
		}
	}
	return
}

func TestClient(t *testing.T) {
	init4Tests(t)
	AntiFlood.Limit = -1
	if len(*tstRmt) < 1 {
		t.Skip("rmt needed")
	}
	tstChnClnt = make(chan struct{}, 2)
	id := tstClntId
	conn, err := Client(t.Log, tstClnt,
		*tstProt, *tstRmt, ConnectedTcp)
	if err != nil {
		t.Error(err)
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
	t.Log("client died", id)
}

func tstTCPSvr(addr [4]string, chn [2]chan RoutCommStruc) {
	tstChnSvr[0] <- true
	go tstUDPSvr(tstChnSvr[1], chn)
}

func tstUDPSvr(fin chan bool, chn [2]chan RoutCommStruc) {
	wait4Rcvr := true
	done := false
	defer func() {
		tstT.Log("exiting 1 connection")
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
		tstT.Log("exit 1 connection")
	}()
	for {
		comm := <-chn[1]
		if comm.Err != nil {
			tstT.Error(comm.Err)
			wait4Rcvr = false
			return
		}
		switch comm.Act {
		case FlowChnEnd:
			//tstT.Error(eztools.ErrIncomplete)
			wait4Rcvr = false
			return
		case FlowChnSnd:
			if string(comm.Data) == tstBye {
				tstT.Log("exiting server after bye sent")
				done = true
				return
			}
		case FlowChnRcv:
			tstT.Log("server echoing", string(comm.Data))
			comm.Act = FlowChnSnd
			chn[0] <- comm
		}
	}
}

func TestServer(t *testing.T) {
	init4Tests(t)
	var (
		conn  *net.UDPConn
		lstnr net.Listener
		err   error
	)
	AntiFlood.Limit = -1

	id := tstClntId
	t.Log("server", id, "will TO in 10 minutes!")
	for i := range tstChnSvr {
		tstChnSvr[i] = make(chan bool, TestSimulClients)
	}
	var chnEnd chan error
	if strings.HasPrefix(*tstProt, "udp") {
		conn, err = ListenUdp(*tstProt, *tstLcl)
		if err != nil {
			t.Error(err)
			return
		}
		var chn [2]chan RoutCommStruc
		for i := range chn {
			chn[i] = make(chan RoutCommStruc, TestReadBufLen)
		}
		go ConnectedUdp(t.Log, chn, conn)
		tstChnSvr[0] <- true
		go tstUDPSvr(tstChnSvr[1], chn)
	} else {
		chnEnd = make(chan error)
		lstnr, err = ListenTcp(t.Log, tstTCPSvr, *tstProt, *tstLcl, ConnectedTcp, chnEnd)
	}
	if err != nil {
		t.Error(err)
		return
	}
	const lstn = "listening"
	switch {
	case lstnr != nil:
		t.Log(lstn, *tstProt, lstnr.Addr())
		*tstRmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		t.Log(lstn, *tstProt, conn.LocalAddr())
		*tstRmt = conn.LocalAddr().String() // to test clients
	default:
		t.Error("no connection info")
	}
	if tstClntRdMsg != nil {
		tstClntRdMsg <- struct{}{}
	}
	// wait for 1 connection
	<-tstChnSvr[0]
ServerLoop:
	for {
		t.Log("server waiting for end")
		select {
		case <-time.After(time.Second * TestTimeout):
			t.Skip("server TO")
			break ServerLoop
		case <-tstChnSvr[0]:
			t.Log("server waiting for 1 more end")
		case done := <-tstChnSvr[1]:
			t.Log("server waited for 1 more end")
			if done {
				break ServerLoop
			}
		}
	}
	t.Log("server to die")
	if lstnr != nil {
		lstnr.Close()
	}
	if conn != nil {
		conn.Close()
	}
	if chnEnd != nil { //TCP
		t.Log("waiting for server to die")
		<-chnEnd
	}
	eztools.Debugging = false // do not let conn routines to print since now
	t.Log("server died")
	if tstSvrChan != nil { //TCP
		tstSvrChan <- struct{}{}
	}
}
