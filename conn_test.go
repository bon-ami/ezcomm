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
	bye                = "ciao"
	separator          = ";"
	TEST_READ_BUF_LEN  = 10
	TEST_SIMUL_CLIENTS = 10
	TEST_SERVER_TO     = 10
)

var (
	initDone                  bool
	prot, lcl, rmt, msg       *string
	verbose, clntNo, msgCount *int
	clntRdMsg                 chan struct{}
	svrChan                   chan error
)

func init() {
	prot = flag.String("prot", "udp4", "protocol, tcp, tcp4, tcp6, unix or unixpacket. default is udp.")
	lcl = flag.String("lcl", DEF_ADR /*"localhost:"*/, "local address")
	rmt = flag.String("rmt", "", "remote address")
	msg = flag.String("msg", "", "messages to send, separated by \""+separator+"\". \""+bye+"\" to end udp server")
	msgCount = flag.Int("msgCount", 0, "number of messages to send per client for TestSvrCln. default=all greek")
	verbose = flag.Int("verbose", 0, "verbose level. 0-3.")
	clntNo = flag.Int("quan", TEST_SIMUL_CLIENTS, "quantity of clients. default="+strconv.Itoa(TEST_SIMUL_CLIENTS))
}

func init4Tests(t *testing.T) {
	if initDone {
		return
	}
	if t == nil {
		panic(t)
	}
	initDone = true
	if !flag.Parsed() {
		flag.Parse()
	}
	eztools.SetLogFunc(t.Log)
	eztools.Debugging = true
	eztools.Verbose = *verbose
	return
}

func TestSvrCln(t *testing.T) {
	init4Tests(t)
	if len(*rmt) > 0 || len(*msg) > 0 {
		t.Skip("rmt & msg are ignored. msgCount and quan count.")
	}
	//*prot = "tcp"
	svrChan = make(chan error)
	clntRdMsg = make(chan struct{})
	go func() {
		TestServer(t)
	}()
	<-clntRdMsg // server ready
	const (
		CharHd = '\u0393'
		CharTl = '\u03C9' + 1 /*bonjour*/
	)
	if *msgCount <= 0 {
		*msgCount = CharTl - CharHd + 1
		t.Log(CharTl, CharHd, *msgCount)
	}
	*msgCount += CharHd
	clnts := *clntNo
	if *verbose > 2 {
		t.Log("clnCount", clnts, "msgCount", *msgCount)
	}
	chs := make([]chan struct{}, clnts)
	for i := range chs {
		chs[i] = make(chan struct{})
	}
	for i := 0; i < clnts; i++ {
		a := strconv.Itoa(i)
		*msg = "bonjour"
		for j := CharHd; j < rune(*msgCount); j++ {
			*msg += separator + string(j)
			*msg += a
		}
		*clntNo = i
		if *verbose > 2 {
			t.Log("client", i, *msg)
		}
		go func(ind int) {
			TestClient(t)
			chs[ind] <- struct{}{}
		}(i)
		<-clntRdMsg // next client available
	}
	for i := range chs {
		<-chs[i]
	}
	*msg = bye
	*clntNo = -1
	if *verbose > 2 {
		t.Log("client", *msg)
	}
	clntRdMsg = nil
	TestClient(t)
	if !strings.HasPrefix(*prot, "udp") {
		<-svrChan
	}
}

func TestClient(t *testing.T) {
	init4Tests(t)
	if len(*rmt) < 1 || len(*msg) < 1 {
		t.Skip("rmt & msg needed")
	}
	msgs := strings.Split(*msg, separator)
	clntId := *clntNo
	if clntRdMsg != nil {
		clntRdMsg <- struct{}{}
	}
	fin := make(chan struct{})
	conn, err := Client(*prot, *rmt, func(conn net.Conn) {
		defer func() {
			if *verbose > 1 {
				t.Log("client dieing", clntId)
			}
			conn.Close()
			fin <- struct{}{}
		}()
		if *verbose > 0 {
			t.Log(clntId, conn.LocalAddr(), "->", conn.RemoteAddr())
		}
		sndNrcv := func(msg string) error {
			conn.Write([]byte(msg))
			if *verbose > 2 {
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
					if *verbose > 2 {
						t.Log(clntId, string(buf))
					}
				}
			}
			return nil //err
		}
		if *verbose > 2 {
			t.Log(clntId, "messages", msgs)
		}
		for _, msg1 := range msgs {
			if sndNrcv(msg1) != nil {
				return
			}
		}
		if !strings.HasPrefix(*prot, "udp") {
			sndNrcv(bye)
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
	if *verbose > 1 {
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

	if strings.HasPrefix(*prot, "udp") {
		conn, err = ListenUdp(*prot, *lcl)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			beg <- struct{}{} // once only
			defer func() {
				if *verbose > 2 {
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
					if *verbose > 2 {
						t.Log("svr:", addr.String(), string(bytes))
					}
				}
				if _, err = conn.WriteToUDP(bytes[:ln], addr); err != nil {
					t.Error(err)
				}
				if string(bytes[:ln]) == bye {
					break
				}
			}
		}()
	} else {
		if svrChan == nil {
			svrChan = make(chan error)
			svrWait = true
		}
		lstnr, err = ListenTcp(*prot, *lcl, func(conn net.Conn) {
			beg <- struct{}{}
			defer func() {
				if *verbose > 2 {
					t.Log("server dieing")
				}
				conn.Close()
				fin <- struct{}{}
			}()
			if *verbose > 2 {
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
					if *verbose > 2 {
						t.Log("svr:", string(bytes))
					}
				}
				conn.Write(bytes[:ln])
				if string(bytes[:ln]) == bye {
					break
				}
			}
		}, svrChan)
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
		*rmt = lstnr.Addr().String() // to test clients
	case conn != nil:
		t.Log(lstn, conn.LocalAddr())
		*rmt = conn.LocalAddr().String() // to test clients
	default:
		t.Error("no connection info")
	}
	if clntRdMsg != nil {
		clntRdMsg <- struct{}{}
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
		if *verbose > 1 {
			t.Log("waiting for server to die")
		}
		<-svrChan
	}
	if *verbose > 1 {
		t.Log("server died")
	}
}
