package ezcomm

import (
	"runtime"
	"testing"
	"time"
)

var (
	tstSvrT  SvrTCP
	tstSvrTP map[string]struct{}
	tstSvrTD chan struct{}
)

func tstSvrTAct(comm RoutCommStruc) {
	chkDone := func() {
		if lft := len(tstSvrTP); lft < 1 {
			TstT.Log("no remaining clients")
			tstSvrTD <- struct{}{}
		} else {
			TstT.Log("remaining clients", tstSvrTP)
		}
	}
	disc := func(addr string) {
		TstT.Log("disconnecting", addr)
		tstSvrT.Disconnect(addr)
		if len(addr) > 0 {
			delete(tstSvrTP, addr)
			chkDone()
		} else { // TODO: what if new connected() at the same time?
			chkDone()
			tstSvrTP = nil
		}
	}
	TstT.Log("act received", comm)
	if comm.Err != nil && comm.Act != FlowChnEnd {
		TstT.Fatal(comm.Err)
		tstSvrTD <- struct{}{}
	}
	var addr string
	if comm.PeerTCP != nil {
		addr = comm.PeerTCP.String()
		//tstT.Log("addr=", addr)
	}
	switch comm.Act {
	case FlowChnEnd:
		var ok bool
		if tstSvrTP != nil {
			_, ok = tstSvrTP[addr]
		}
		if ok {
			TstT.Log("client gone", addr)
			disc(addr)
		} else {
			TstT.Log("already removed", addr)
			chkDone()
		}
	case FlowChnRcv: // echo it
		tstSvrT.Send(addr, comm.Data)
	case FlowChnSnd:
		if string(comm.Data) == tstBye {
			tstSvrT.Stop()
			disc("")
			/*} else {
			tstT.Log(tstBye, "!=", comm.Data, ".")*/
		}
	}
}

func tstSvrTConn(addr [4]string) {
	if len(addr[1]) < 1 {
		TstT.Log("listening", addr[0])
		return
	}
	TstT.Log("connected", addr)
	tstSvrTP[addr[1]] = struct{}{}
}

func TestSvrTcp(t *testing.T) {
	Init4Tests(t)
	*TstProt = "tcp"
	tstSvrTP = make(map[string]struct{})
	tstSvrTD = make(chan struct{}, 1)
	tstSvrT.ActFunc = tstSvrTAct
	tstSvrT.LogFunc = t.Log
	tstSvrT.ConnFunc = tstSvrTConn
	//t.Log("listen", *tstProt, *tstLcl)
	err := tstSvrT.Listen(*TstProt, *TstLcl)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-time.After(TstTO):
		t.Skip("server TO")
	case <-tstSvrTD:
	}
	t.Log("wait for server")
	tstSvrT.Wait(false)
	t.Log("wait for clients")
	tstSvrT.Wait(true)
	t.Log("wait for all unfinished")
	<-time.After(time.Second)
	if n := runtime.NumGoroutine(); n > 2 {
		<-time.After(time.Second * 5)
		if n := runtime.NumGoroutine(); n > 2 {
			t.Error("routines left (should be 0)", n-2)
		}
	}
}
