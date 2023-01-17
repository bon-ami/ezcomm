package ezcomm

import (
	"testing"
	"time"
)

func TestHttpSvr(t *testing.T) {
	Init4Tests(t)
	defer Deinit4Tests()
	//*tstProt = "tcp"
	//t.Log("listen", *tstProt, *tstLcl)
	lstnr, err := ListenTCP(nil, nil, "", DefPeerAdr+":", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svr := MakeHTTPSvr()
	svr.FS("", *tstRoot, nil)
	ch := svr.Serve(lstnr)
	t.Log("wait for server for", tstTO, "seconds")
	select {
	case <-time.After(tstTO):
		t.Log("shutting down server")
		err = svr.Shutdown(1)
	case err = <-ch:
	}
	if err != nil {
		t.Fatal(err)
	}
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}
