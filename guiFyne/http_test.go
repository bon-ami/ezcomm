package main

import (
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var tstT *testing.T

func tstReadHTTP(chn chan error) {
	if localAddrSlc == nil || len(localAddrSlc) < 1 {
		chn <- eztools.ErrNoValidResults
		return
	}
	addr := "http://" + localAddrSlc[0] + lanPrtHTTP
	tstT.Log("reading from", addr)
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
			tstT.Log(string(buf[:n]))
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

func tstSwitchHTTP(chnHTTP chan bool, chnRes chan error) {
	chnRead := make(chan error, 1)
	chnSvr := make(chan error, 1)
	var err error
	const LoopCnt = 5
	loopCnt := -1
TSTHTTPLOOP:
	for {
		select {
		case <-time.After(time.Second):
			if loopCnt < 0 {
				// wait for server to run
				break
			}
			tstReadHTTP(chnRead)
			loopCnt--
			if loopCnt == 0 {
				err = eztools.ErrAbort
				break TSTHTTPLOOP
			}
		case err = <-chnRead:
			//tstT.Log(err)
			break TSTHTTPLOOP
		case err = <-chnSvr:
			break TSTHTTPLOOP
		case frmHTTP := <-chnHTTP:
			switch frmHTTP {
			case true:
				chnSvr = runHTTP()
				tstReadHTTP(chnRead)
				loopCnt = LoopCnt
			case false:
				stpHTTP()
			}
		}
	}
	chnRes <- err
}

func TestHttp(t *testing.T) {
	tstT = t
	ezcApp = test.NewApp()
	chnHTTP := make(chan bool, 1)
	chnRes := make(chan error, 1)
	go tstSwitchHTTP(chnHTTP, chnRes)
	run(chnHTTP)
	if err := <-chnRes; err != nil {
		t.Error(err)
	}
}
