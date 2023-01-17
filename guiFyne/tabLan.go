package main

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	chnLan       chan bool
	httpSvr      *ezcomm.HTTPSvr
	localAddrSlc []string
	lanPrtHTTP   string
	lanWeb       *widget.List
	lanBut       *widget.Button
	lanLbl       *widget.Label
	lanLst       *widget.RadioGroup
)

func httpListNum() int {
	return len(localAddrSlc)
}

func httpListUpdate(id widget.ListItemID, item fyne.CanvasObject) {
	if len(localAddrSlc) <= id {
		item.(*widget.Label).SetText("")
		return
	}
	item.(*widget.Label).SetText(localAddrSlc[id] + lanPrtHTTP)
}

func runHTTP() chan error {
	lstnr, err := ezcomm.ListenTCP(
		Log, nil, "", ":", nil, nil)
	if err != nil {
		eztools.LogFatal(err)
	}
	addr := lstnr.Addr().String()
	ind := strings.LastIndex(addr, ":")
	lanPrtHTTP = addr[ind:]
	lanWeb.Refresh()
	httpSvr = ezcomm.MakeHTTPSvr()
	httpSvr.FS("", "", httpFS(appStorage.RootURI().Path()))
	return httpSvr.Serve(lstnr)
}

func stpHTTP() error {
	ret := httpSvr.Shutdown(1)
	httpSvr = nil
	return ret
}

func runUDP(defLanPrtStr string) *net.UDPConn {
	conn, err := ezcomm.ListenUDP(ezcomm.StrUDP,
		":"+defLanPrtStr)
	if err != nil || conn == nil {
		lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
		Log("discovery failure", err)
		return nil
	}
	lanLbl.SetText(ezcomm.StringTran["StrLst"])
	go ezcomm.ConnectedUDP(Log, chn, conn)
	return conn
}

func sndUDP(conn *net.UDPConn, addrBrd *net.UDPAddr, bonjour []byte) {
	if eztools.Debugging &&
		eztools.Verbose > 1 {
		eztools.Log("discover", *conn)
	}
	pckNt := ezcomm.RoutCommStruc{
		Act:     ezcomm.FlowChnSnd,
		Data:    bonjour,
		PeerUDP: addrBrd}
	chn[0] <- pckNt
}

func switchHTTP(chn chan bool) {
	for {
		switch <-chn {
		case true:
			runHTTP()
		case false:
			stpHTTP()
		}
	}
}

func lanListen(chnHTTP chan bool) {
	var (
		conn  *net.UDPConn
		reqUI bool
		pckNt ezcomm.RoutCommStruc
	)
	for i := range chn {
		chn[i] = make(chan ezcomm.RoutCommStruc,
			ezcomm.FlowComLen)
	}
	end2worker := func() {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("ending discovery")
		}
		pckNt.Act = ezcomm.FlowChnEnd
		chn[0] <- pckNt
		conn = nil
	}
	defLanPrtInt := ezcomm.DefLanPrt()
	defLanPrtStr := strconv.Itoa(defLanPrtInt)
	var bonjour []byte
	for _, c := range ezcomm.EzcName {
		bonjour = append(bonjour, byte(c))
	}
	addrBrd, err := net.ResolveUDPAddr(ezcomm.StrUDP,
		ezcomm.DefBrdAdr+":"+defLanPrtStr)
	if err != nil || addrBrd == nil {
		Log("bad broadcast addr", err)
		return
	}
	if chnHTTP == nil {
		chnHTTP = make(chan bool, 1)
		go switchHTTP(chnHTTP)
	}
	var localAddrMap, peerMap map[string]struct{}
	for {
		select {
		case reqUI = <-chnLan:
			switch reqUI {
			case false:
				if conn != nil {
					end2worker()
					<-chn[1] // wait for ezcomm to end
				}
				if httpSvr != nil {
					Log("false to", chnHTTP)
					chnHTTP <- false
				}
			case true:
				if conn != nil {
					sndUDP(conn, addrBrd, bonjour)
					break
				}
				localAddrs, err := net.InterfaceAddrs()
				if err != nil || localAddrs == nil ||
					len(localAddrs) < 1 {
					Log("bad local addr", err)
					lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
					break
				}
				peerMap = make(map[string]struct{})
				localAddrMap = make(map[string]struct{})
				localAddrSlc = make([]string, 0)
				for _, localAddr1 := range localAddrs {
					str := localAddr1.String()
					ind := strings.LastIndex(str, "/")
					localAddrMap[str[:ind]] = struct{}{}
					var pref, suff string
					if str[(ind+1):] == "64" {
						pref = "["
						suff = "]"
					}
					localAddrSlc = append(localAddrSlc,
						pref+str[:ind]+suff)
				}
				conn = runUDP(defLanPrtStr)
				if conn == nil {
					break
				}
				if eztools.Debugging && eztools.Verbose > 1 {
					eztools.Log("discovery", *conn)
				}
				if httpSvr == nil {
					chnHTTP <- true
				}
			}
		case pckNt = <-chn[1]:
			switch pckNt.Act {
			case ezcomm.FlowChnEnd:
				end2worker()
			case ezcomm.FlowChnSnt:
				if pckNt.Err != nil {
					Log("discovery failure", pckNt.Err)
				}
			case ezcomm.FlowChnRcv:
				if pckNt.PeerUDP == nil {
					Log("discovery failure. NO peer addr!")
					break
				}
				if pckNt.PeerUDP.Port != defLanPrtInt {
					break
				}
				if !bytes.Equal(pckNt.Data, bonjour) {
					break
				}
				peer := pckNt.PeerUDP.IP.String()
				_, ok := localAddrMap[peer]
				if ok {
					break
				}
				if _, ok := peerMap[peer]; ok {
					// duplicate
					break
				}
				peerMap[peer] = struct{}{}
				add2Rmt(0, peer)
				lanLst.Append(peer)
				if eztools.Debugging && eztools.Verbose > 1 {
					eztools.Log("discovered", peer)
				}
			}
		}
	}
}

// toast shows a toast window
// Parameters: index to ezcomm.StringTran and second line string
func toast(id, inf string) {
	const to = time.Second * 3
	go func(inf string) {
		if drv, ok := ezcApp.Driver().(desktop.Driver); ok {
			w := drv.CreateSplashWindow()
			w.SetContent(widget.NewLabel(
				ezcomm.StringTran[id] + "\n" + inf))
			w.Show()
			go func() {
				time.Sleep(to)
				w.Close()
			}()
		}
	}(inf)
}

func makeTabLan(chnHTTP chan bool) *container.TabItem {
	chnLan = make(chan bool, 1)
	lanWeb = widget.NewList(
		httpListNum,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		httpListUpdate)
	lanWeb.OnSelected = func(id widget.ListItemID) {
		if len(localAddrSlc) <= id {
			return
		}
		drv := ezcApp.Driver()
		clipboard := drv.AllWindows()[0].Clipboard()
		clipboard.SetContent(localAddrSlc[id] + lanPrtHTTP)
		toast("StrCopied", localAddrSlc[id]+lanPrtHTTP)
	}
	lanLbl = widget.NewLabel("")
	go lanListen(chnHTTP)
	lanBut = widget.NewButton(ezcomm.StringTran["StrPokePeer"], func() {
		chnLan <- true
	})
	lanLst = widget.NewRadioGroup([]string{},
		func(sel string) {
			sockRmt[0].SetText(sel)
		})

	return container.NewTabItem(ezcomm.StringTran["StrInfLan"],
		container.NewBorder(nil, container.NewVBox(lanBut, lanLst),
			nil, nil, lanWeb))
}

func tabLanShown(yes bool) {
	switch yes {
	case true:
		lanBut.Refresh()
	case false:
	}
	chnLan <- yes
}
