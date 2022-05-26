package main

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
	tstBye             = "ciao"
	tstSeparator       = ";"
	TEST_READ_BUF_LEN  = 10
	TEST_SIMUL_CLIENTS = 10
	TEST_SERVER_TO     = 10
)

var (
	tstInitDone                        bool
	tstProt, tstLcl, tstRmt, tstMsg    *string
	tstVerbose, tstClntNo, tstMsgCount *int
	tstClntRdMsg                       chan struct{}
	tstSvrChan                         chan error
)

func init() {
	tstProt = flag.String("prot", "udp4", "protocol, tcp, tcp4, tcp6, unix or unixpacket. default is udp.")
	tstLcl = flag.String("lcl", DEF_ADR /*"localhost:"*/, "local address")
	tstRmt = flag.String("rmt", "", "remote address")
	tstMsg = flag.String("msg", "", "messages to send, separated by \""+tstSeparator+"\". \""+tstBye+"\" to end udp server")
	tstMsgCount = flag.Int("msgCount", 0, "number of messages to send per client for TestSvrCln. default=all greek")
	tstVerbose = flag.Int("verbose", 0, "verbose level. 0-3.")
	tstClntNo = flag.Int("quan", TEST_SIMUL_CLIENTS, "quantity of clients. default="+strconv.Itoa(TEST_SIMUL_CLIENTS))
}

func init4Tests(t *testing.T) {
	if tstInitDone {
		return
	}
	if t == nil {
		panic(t)
	}
	tstInitDone = true
	if !flag.Parsed() {
		flag.Parse()
	}
	eztools.SetLogFunc(t.Log)
	eztools.Debugging = true
	eztools.Verbose = *tstVerbose
	return
}

func TestSvrCln(t *testing.T) {
	init4Tests(t)
	if len(*tstRmt) > 0 || len(*tstMsg) > 0 {
		t.Skip("rmt & msg are ignored. msgCount and quan count.")
	}
	//*prot = "tcp"
	tstSvrChan = make(chan error)
	tstClntRdMsg = make(chan struct{})
	go func() {
		TestServer(t)
	}()
	<-tstClntRdMsg // server ready
	const (
		CharHd = '\u0393'
		CharTl = '\u03C9' + 1 /*bonjour*/
	)
	if *tstMsgCount <= 0 {
		*tstMsgCount = CharTl - CharHd + 1
		t.Log(CharTl, CharHd, *tstMsgCount)
	}
	*tstMsgCount += CharHd
	clnts := *tstClntNo
	if *tstVerbose > 2 {
		t.Log("clnCount", clnts, "msgCount", *tstMsgCount)
	}
	chs := make([]chan struct{}, clnts)
	for i := range chs {
		chs[i] = make(chan struct{})
	}
	for i := 0; i < clnts; i++ {
		a := strconv.Itoa(i)
		*tstMsg = "bonjour"
		for j := CharHd; j < rune(*tstMsgCount); j++ {
			*tstMsg += tstSeparator + string(j)
			*tstMsg += a
		}
		*tstClntNo = i
		if *tstVerbose > 2 {
			t.Log("client", i, *tstMsg)
		}
		go func(ind int) {
			TestClient(t)
			chs[ind] <- struct{}{}
		}(i)
		<-tstClntRdMsg // next client available
	}
	for i := range chs {
		<-chs[i]
	}
	*tstMsg = tstBye
	*tstClntNo = -1
	if *tstVerbose > 2 {
		t.Log("client", *tstMsg)
	}
	tstClntRdMsg = nil
	TestClient(t)
	if !strings.HasPrefix(*tstProt, "udp") {
		<-tstSvrChan
	}
}

func TestClient(t *testing.T) {
	init4Tests(t)
	if len(*tstRmt) < 1 || len(*tstMsg) < 1 {
		t.Skip("rmt & msg needed")
	}
	msgs := strings.Split(*tstMsg, tstSeparator)
	clntId := *tstClntNo
	if tstClntRdMsg != nil {
		tstClntRdMsg <- struct{}{}
	}
	fin := make(chan struct{})
	conn, err := Client(*tstProt, *tstRmt, func(conn net.Conn) {
		defer func() {
			if *tstVerbose > 1 {
				t.Log("client dieing", clntId)
			}
			conn.Close()
			fin <- struct{}{}
		}()
		if *tstVerbose > 0 {
			t.Log(clntId, conn.LocalAddr(), "->", conn.RemoteAddr())
		}
		sndNrcv := func(msg string) error {
			conn.Write([]byte(msg))
			if *tstVerbose > 2 {
				t.Log(clntId, "sent", msg)
			}
			buf := make([]byte, TEST_READ_BUF_LEN)
			ln, err := conn.Read(buf)
			if err != nil {
				t.Log(err)
			} else {
				if string(buf[:ln]) != msg {
					t.Error("sent", msg, "!=", "got", ln, buf[:ln])
				} else {
					if *tstVerbose > 2 {
						t.Log(clntId, string(buf))
					}
				}
			}
			return nil //err
		}
		if *tstVerbose > 2 {
			t.Log(clntId, "messages", msgs)
		}
		for _, msg1 := range msgs {
			if sndNrcv(msg1) != nil {
				return
			}
		}
		if !strings.HasPrefix(*tstProt, "udp") {
			sndNrcv(tstBye)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		panic(conn)
	}
	defer conn.Close()
	<-fin
	if *tstVerbose > 1 {
		t.Log("client died", clntId)
	}
}

func TestServer(t *testing.T) {
	init4Tests(t)
	beg := make(chan struct{}, TEST_SIMUL_CLIENTS)
	fin := make(chan struct{})
	var svrWait bool
	var (
		conn  *net.UDPConn
		lstnr net.Listener
		err   error
	)

	if strings.HasPrefix(*tstProt, "udp") {
		conn, err = ListenUdp(*tstProt, *tstLcl)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			beg <- struct{}{} // once only
			defer func() {
				if *tstVerbose > 2 {
					t.Log("server dieing")
				}
				conn.Close()
				fin <- struct{}{}
			}()
			for {
				bytes := make([]byte, TEST_READ_BUF_LEN)
				ln, addr, err := conn.ReadFromUDP(bytes)
				if err != nil {
					t.Log(err)
					break
				} else {
					if ln < 1 {
						t.Error(ln)
					}
					if *tstVerbose > 2 {
						t.Log("svr:", addr.String(), string(bytes))
					}
				}
				if _, err = conn.WriteToUDP(bytes[:ln], addr); err != nil {
					t.Error(err)
				}
				if string(bytes[:ln]) == tstBye {
					break
				}
			}
		}()
	} else {
		if tstSvrChan == nil {
			tstSvrChan = make(chan error)
			svrWait = true
		}
		lstnr, err = ListenTcp(*tstProt, *tstLcl, func(conn net.Conn) {
			beg <- struct{}{}
			defer func() {
				if *tstVerbose > 2 {
					t.Log("server dieing")
				}
				conn.Close()
				fin <- struct{}{}
			}()
			if *tstVerbose > 2 {
				t.Log(conn.LocalAddr(), "<-", conn.RemoteAddr())
			}
			for {
				bytes := make([]byte, TEST_READ_BUF_LEN)
				ln, err := conn.Read(bytes)
				if err != nil {
					t.Log(err)
					break
				} else {
					if ln < 1 {
						t.Error(ln)
					}
					if *tstVerbose > 2 {
						t.Log("svr:", string(bytes))
					}
				}
				conn.Write(bytes[:ln])
				if string(bytes[:ln]) == tstBye {
					break
				}
			}
		}, tstSvrChan)
	}
	if err != nil {
		t.Fatal(err)
	}
	/*if lstnr == nil && conn == nil {
		panic(lstnr)
	}*/
	//defer conn.Close()
	const lstn = "listening"
	switch {
	case lstnr != nil:
		t.Log(lstn, lstnr.Addr())
		*tstRmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		t.Log(lstn, conn.LocalAddr())
		*tstRmt = conn.LocalAddr().String() // to test clients
	default:
		t.Error("no connection info")
	}
	if tstClntRdMsg != nil {
		tstClntRdMsg <- struct{}{}
	}
	<-beg
	<-fin
	switch {
	case lstnr != nil:
	SERVER_LOOP:
		for {
			select {
			case <-time.After(time.Second * TEST_SERVER_TO):
				break SERVER_LOOP
			case <-beg:
				<-fin
			}
		}
		lstnr.Close()
		/*case conn != nil:
		conn.Close()*/
	}
	if svrWait {
		if *tstVerbose > 1 {
			t.Log("waiting for server to die")
		}
		<-tstSvrChan
	}
	if *tstVerbose > 1 {
		t.Log("server died")
	}
}
