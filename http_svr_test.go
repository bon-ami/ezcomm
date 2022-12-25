package ezcomm

import (
	"testing"
	"time"
)

func TestHttpSvr(t *testing.T) {
	//init4Tests(t)
	//*tstProt = "tcp"
	//t.Log("listen", *tstProt, *tstLcl)
	lstnr, err := ListenTcp(nil, nil, "", DefAdr+":", nil, nil)
	err = HttpServ(lstnr, ".")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("wait for server")
	<-time.After(time.Second)
	/*if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}*/
}
