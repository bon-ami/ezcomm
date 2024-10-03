package ezcomm

import (
	"flag"
	"strconv"
	"testing"
	"time"

	"gitee.com/bon-ami/eztools/v6"
)

const (
	tstDefSimulClients = 10
	tstDefTimeout      = 10
	tstDefVerbose      = 0
	tstDefRoot         = "testdata"
	tstDefProt         = "tcp4"
	// TstHalo beginning string
	TstHalo = "bonjour"
	// TstBye ending string
	TstBye       = "ciao"
	tstSeparator = ";"
)

var (
	// TstProt protocol for tests
	TstProt *string
	// TstLcl local socket for tests
	TstLcl *string
	// TstRmt remote socket for tests
	TstRmt *string
	// TstMsg messages for tests
	TstMsg *string
	// TstRoot root dir for tests
	TstRoot *string
	// TstClntNo number of clients for tests
	TstClntNo *int
	// TstMsgCount number of messages for tests
	TstMsgCount *int
	// TstTimeout number of seconds as timeout for tests
	TstTimeout *int
	// TstTO timeout duration for tests
	TstTO time.Duration
	// TstT direct pointer to tests. better not to be used
	TstT *testing.T

	tstVerbose   *int
	tstClntRdMsg chan struct{}
	tstSvrChan   chan struct{}
	tstInitDone  bool
	tstClntID    int
)

// InitFlags4Tests needs to be included in init() from a test file,
// to be parsed in Init4Tests()
// prot: protocol, tcp, tcp4, tcp6, unix or unixpacket
// lcl: local address
// rmt: remote address
// msg: messages to send
// root: root dir for http server
// timeout: in seconds
// verbose: verbose level
// msgCount: number of messages to send per client for TestSvrCln
// quan: quantity of clients
func InitFlags4Tests() {
	TstProt = flag.String("prot", tstDefProt,
		"protocol, tcp, tcp4, tcp6, unix or unixpacket")
	TstLcl = flag.String("lcl", Localhost+":", "local address")
	TstRmt = flag.String("rmt", "", "remote address")
	TstMsg = flag.String("msg", "", "messages to send, separated by \""+
		tstSeparator+"\". \""+TstBye+"\" to end udp server")
	TstRoot = flag.String("root", tstDefRoot, "root dir for http server")
	TstTimeout = flag.Int("timer", -1, "in seconds. default="+
		strconv.Itoa(tstDefTimeout)+", NOT=>")
	tstVerbose = flag.Int("verbose", tstDefVerbose, "verbose level. "+
		"default="+strconv.Itoa(tstDefVerbose))
	TstMsgCount = flag.Int("msgCount", 0, "number of messages to send "+
		"per client for TestSvrCln. default=all greek")
	TstClntNo = flag.Int("quan", tstDefSimulClients,
		"quantity of clients or seconds for HTTP shutdown")
}

// Deinit4Tests to use between multiple tests matching Init4Tests
func Deinit4Tests() {
	if eztools.Verbose > 1 {
		eztools.Log("tests vars deiniting")
	}
	tstInitDone = false
	TstT = nil
	*TstProt = ""
	*TstLcl = Localhost + ":" /*"localhost:"*/
	*TstRmt = ""
	*TstMsg = ""
	*TstRoot = ""
	*TstTimeout = -1
	TstTO = 0
	*TstMsgCount = 0
	*TstClntNo = tstDefSimulClients
	*tstVerbose = tstDefVerbose
	eztools.Verbose = *tstVerbose
}

// Init4Tests parse flags, that are introduced from
// InitFlags4Tests() included in init()
func Init4Tests(t *testing.T) {
	tstClntID++
	if tstInitDone {
		return
	}
	if t == nil {
		panic(t)
	}
	t.Cleanup(Deinit4Tests)
	tstInitDone = true
	TstT = t
	//if !flag.Parsed() {
	flag.Parse()
	//}
	if *TstTimeout < 0 {
		*TstTimeout = tstDefTimeout
	}
	TstTO = time.Second * time.Duration(*TstTimeout)
	eztools.SetLogFunc(func(l ...any) {
		func(m ...any) {
			t.Log(m)
		}(eztools.GetCallerLog(), l)
	})
	if *tstVerbose > 0 && *tstVerbose < 4 {
		eztools.Debugging = true
		eztools.Verbose = *tstVerbose
	}
	if eztools.Verbose > 1 {
		eztools.Log("tests vars inited")
	}
	return
}
