package ezcomm

import (
	"flag"
	"strconv"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v5"
)

const (
	tstBye       = "ciao"
	tstSeparator = ";"
	// TestDefSimulClients default number of simultaneous clients
	TestDefSimulClients = 10
	// TestDefTimeout default timeout for a test
	TestDefTimeout = 10
	// TestDefVerbose default verbose level during tests
	TestDefVerbose = 0
)

var (
	tstInitDone                                    bool
	tstClntID                                      int
	tstProt, tstLcl, tstRmt, tstMsg, tstRoot       *string
	tstClntNo, tstMsgCount, tstTimeout, tstVerbose *int
	tstClntRdMsg                                   chan struct{}
	tstSvrChan                                     chan struct{}
	tstTO                                          time.Duration
	tstT                                           *testing.T
)

func init() {
	tstProt = flag.String("prot", "udp4", "protocol, tcp, tcp4, tcp6, unix or unixpacket. default is udp.")
	tstLcl = flag.String("lcl", DefPeerAdr+":" /*"localhost:"*/, "local address")
	tstRmt = flag.String("rmt", "", "remote address")
	tstMsg = flag.String("msg", "", "messages to send, separated by \""+tstSeparator+"\". \""+tstBye+"\" to end udp server")
	tstRoot = flag.String("root", "", "root dir for http server")
	tstTimeout = flag.Int("timer", -1, "in seconds. default="+strconv.Itoa(TestDefTimeout))
	tstVerbose = flag.Int("verbose", TestDefVerbose, "verbose level. default="+strconv.Itoa(TestDefVerbose))
	tstMsgCount = flag.Int("msgCount", 0, "number of messages to send per client for TestSvrCln. default=all greek")
	tstClntNo = flag.Int("quan", TestDefSimulClients, "quantity of clients. default="+strconv.Itoa(TestDefSimulClients))
}

// Deinit4Tests to use between multiple tests matching Init4Tests
func Deinit4Tests() {
	tstInitDone = false
	tstT = nil
	*tstProt = ""
	*tstLcl = DefPeerAdr + ":" /*"localhost:"*/
	*tstRmt = ""
	*tstMsg = ""
	*tstRoot = ""
	*tstTimeout = -1
	tstTO = 0
	*tstVerbose = TestDefVerbose
	eztools.Verbose = *tstVerbose
	*tstMsgCount = 0
	*tstClntNo = TestDefSimulClients
}

// Init4Tests inits flags
// prot: protocol, tcp, tcp4, tcp6, unix or unixpacket
// lcl: local address
// rmt: remote address
// msg: messages to send
// root: root dir for http server
// timeout: in seconds
// verbose: verbose level
// msgCount: number of messages to send per client for TestSvrCln
// quan: quantity of clients
func Init4Tests(t *testing.T) {
	tstClntID++
	if tstInitDone {
		return
	}
	if t == nil {
		panic(t)
	}
	tstInitDone = true
	tstT = t
	//if !flag.Parsed() {
	flag.Parse()
	//}
	if *tstTimeout < 0 {
		*tstTimeout = TestDefTimeout
	}
	tstTO = time.Second * time.Duration(*tstTimeout)
	eztools.SetLogFunc(func(l ...any) {
		func(m ...any) {
			t.Log(m)
		}(eztools.GetCallerLog(), l)
	})
	eztools.Debugging = true
	if *tstVerbose > 0 && *tstVerbose < 4 {
		eztools.Verbose = *tstVerbose
	}
	return
}
