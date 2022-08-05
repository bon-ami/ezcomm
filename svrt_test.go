package ezcomm

import (
	"runtime"
	"testing"
	"time"
)

var (
	tstSvrT  SvrTcp
	tstSvrTP map[string]struct{}
	tstSvrTD chan struct{}
)

func tstSvrTAct(comm RoutCommStruc) {
	chkDone := func() {
		if lft := len(tstSvrTP); lft < 1 {
			tstT.Log("no remaining clients")
			tstSvrTD <- struct{}{}
		} else {
			tstT.Log("remaining clients", tstSvrTP)
		}
	}
	disc := func(addr string) {
		tstT.Log("disconnecting", addr)
		tstSvrT.Disconnect(addr)
		if len(addr) > 0 {
			delete(tstSvrTP, addr)
			chkDone()
		} else { // TODO: what if new connected() at the same time?
			chkDone()
			tstSvrTP = nil
		}
	}
	tstT.Log("act received", comm)
	if comm.Err != nil && comm.Act != FlowChnEnd {
		tstT.Fatal(comm.Err)
		tstSvrTD <- struct{}{}
	}
	var addr string
	if comm.PeerTcp != nil {
		addr = comm.PeerTcp.String()
		//tstT.Log("addr=", addr)
	}
	switch comm.Act {
	case FlowChnEnd:
		var ok bool
		if tstSvrTP != nil {
			_, ok = tstSvrTP[addr]
		}
		if ok {
			tstT.Log("client gone", addr)
			disc(addr)
		} else {
			tstT.Log("already removed", addr)
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
		tstT.Log("listening", addr[0])
		return
	}
	tstT.Log("connected", addr)
	tstSvrTP[addr[1]] = struct{}{}
}

func TestSvrTcp(t *testing.T) {
	init4Tests(t)
	*tstProt = "tcp"
	tstSvrTP = make(map[string]struct{})
	tstSvrTD = make(chan struct{}, 1)
	tstSvrT.ActFunc = tstSvrTAct
	tstSvrT.LogFunc = t.Log
	tstSvrT.ConnFunc = tstSvrTConn
	//t.Log("listen", *tstProt, *tstLcl)
	err := tstSvrT.Listen(*tstProt, *tstLcl)
	if err != nil {
		t.Fatal(err)
	}
	<-tstSvrTD
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
