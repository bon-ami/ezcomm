package ezcomm

import (
	"strings"
	"testing"
	"time"
)

const tstDefHTTPProt = "tcp"

// TestHttpSvr uses TstProt, TstLcl, TstTimeout and TstClntNo
func TestHttpSvr(t *testing.T) {
	Init4Tests(t)
	if !strings.HasPrefix(*TstProt, "tcp") {
		*TstProt = tstDefHTTPProt
	}
	lstnr, err := ListenTCP(t.Log, nil, *TstProt, *TstLcl, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *tstVerbose > 0 {
		t.Log(lstnr.Addr().String())
	}
	svr := MakeHTTPSvr()
	svr.FS("", *TstRoot, nil)
	ch := svr.Serve(lstnr)
	if *tstVerbose > 1 {
		t.Log("wait for server for", TstTO)
	}
	select {
	case <-time.After(TstTO):
		if *tstVerbose > 1 {
			t.Log("shutting down server")
		}
		err = svr.Shutdown(time.Second * time.Duration(*TstClntNo))
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
