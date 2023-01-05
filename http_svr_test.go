package ezcomm

import (
	"testing"
	"time"
)

func TestHttpSvr(t *testing.T) {
	init4Tests(t)
	//*tstProt = "tcp"
	//t.Log("listen", *tstProt, *tstLcl)
	lstnr, err := ListenTcp(nil, nil, "", DefAdr+":", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ch := HTTPServ(lstnr, "", *tstRoot, nil)
	t.Log("wait for server")
	select {
	case <-time.After(time.Minute):
		break
	case err := <-ch:
		if err != nil {
			t.Fatal(err)
		}
	}
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}
