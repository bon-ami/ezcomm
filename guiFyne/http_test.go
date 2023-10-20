package main

import (
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

func tstReadHTTP(t *testing.T, chn chan error) {
	defer close(chn)
	if localAddrSlc == nil || len(localAddrSlc) < 1 {
		chn <- eztools.ErrNoValidResults
		return
	}
	addr := "http://" + net.JoinHostPort(localAddrSlc[0], lanPrtHTTP)
	if eztools.Verbose > 0 {
		t.Log("reading from", addr)
	}
	resp, err := http.Get(addr)
	if err != nil || resp == nil {
		if err == nil {
			err = eztools.ErrNoValidResults
		}
		chn <- err
		return
	}
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		chn <- eztools.ErrAccess
		return
	}
	bdy := resp.Body
	var sthRd bool
	if bdy != nil {
		defer bdy.Close()
		buf := make([]byte, ezcomm.FlowRcvLen)
		var n int
		sthRdFun := func() {
			sthRd = true
			if eztools.Verbose > 2 {
				t.Log(string(buf[:n]))
			}
		}
		for {
			n, err = bdy.Read(buf)
			if errors.Is(err, io.EOF) {
				sthRdFun()
				err = nil
				break
			}
			if err != nil || n < 1 {
				break
			}
			sthRdFun()
		}
	}
	if err == nil && !sthRd {
		err = eztools.ErrIncomplete
	}
	chn <- err
}

func tstSwitchHTTP(t *testing.T, chnHTTP chan bool, chnRes chan error) {
	chnRead := make(chan error, 1)
	chnSvr := make(chan error, 1)
	var err error
	loopCnt := -1
TSTHTTPLOOP:
	for {
		select {
		case <-time.After(time.Second):
			if loopCnt < 0 {
				// wait for server to run
				break
			}
			tstReadHTTP(t, chnRead)
			loopCnt--
			if loopCnt == 0 {
				err = eztools.ErrAbort
				break TSTHTTPLOOP
			}
		case err = <-chnRead:
			//t.Log(err)
			break TSTHTTPLOOP
		case err = <-chnSvr:
			//t.Log("from server", err)
			break TSTHTTPLOOP
		case frmHTTP := <-chnHTTP:
			switch frmHTTP {
			case true:
				chnSvr = runHTTP()
				tstReadHTTP(t, chnRead)
				loopCnt = *ezcomm.TstTimeout
			case false:
				stpHTTP()
			}
		}
	}
	chnRes <- err
}

// TestHttp uses TstTimeout
func TestHttp(t *testing.T) {
	ezcomm.Init4Tests(t)
	ezcApp = test.NewApp()
	chnHTTP := make(chan bool, 1) // closed by run()
	chnRes := make(chan error, 1)
	defer close(chnRes)
	go tstSwitchHTTP(t, chnHTTP, chnRes)
	run(chnHTTP)
	if err := <-chnRes; err != nil {
		t.Error(err)
	}
}
