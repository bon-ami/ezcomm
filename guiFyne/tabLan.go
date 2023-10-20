package main

import (
	"bytes"
	"net"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	lanTabShown     bool
	chnLan, chnHTTP chan bool
	httpSvr         *ezcomm.HTTPSvr
	localAddrSlc    []string
	lanPrtHTTP      string
	lanWeb          *widget.List
	lanBut          *widget.Button
	lanLbl          *widget.Label
	lanLst          *widget.RadioGroup
	// chnWeb is for TCP client and UDP
	chnWeb [2]chan ezcomm.RoutCommStruc
)

func httpListNum() int {
	return len(localAddrSlc)
}

func httpListUpdate(id widget.ListItemID, item fyne.CanvasObject) {
	if len(localAddrSlc) <= id {
		item.(*widget.Label).SetText("")
		return
	}
	item.(*widget.Label).SetText(net.JoinHostPort(localAddrSlc[id],
		lanPrtHTTP))
}

// runHTTP runs HTTP server in a routine
// Return value: chan is closed upon server exit
func runHTTP() chan error {
	lstnr, err := ezcomm.ListenTCP(
		Log, nil, "", ":", nil, nil)
	if err != nil {
		eztools.LogFatal(err)
	}
	setLclSck(lstnr.Addr().String())
	lanPrtHTTP = sockLcl[1].Text
	setLclAddrs(localAddrSlc)
	lanWeb.Refresh()
	httpSvr = ezcomm.MakeHTTPSvr()
	httpSvr.GET(httpSmsDir, printIncomingWeb)
	httpSvr.POST(httpSmsDir, postIncomingWeb)
	httpSvr.FS(httpAppDir, "", httpFS(appStorage.RootURI().Path()))
	return httpSvr.Serve(lstnr)
}

func stpHTTP() (ret error) {
	lanPrtHTTP = ""
	if httpSvr != nil {
		ret = httpSvr.Shutdown(1)
		httpSvr = nil
	}
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
	go ezcomm.ConnectedUDP(Log, chnWeb, conn)
	return conn
}

func sndUDP(conn *net.UDPConn, addrBrd *net.UDPAddr, bonjour []byte) {
	if eztools.Debugging &&
		eztools.Verbose > 1 {
		eztools.Log("discover", *conn)
	}
	if len(lanPrtHTTP) > 0 {
		for _, c := range lanPrtHTTP {
			bonjour = append(bonjour, byte(c))
		}
	}
	pckNt := ezcomm.RoutCommStruc{
		Act:     ezcomm.FlowChnSnd,
		Data:    bonjour,
		PeerUDP: addrBrd}
	chnWeb[0] <- pckNt
}

func switchHTTP() {
	for {
		switch <-chnHTTP {
		case true:
			runHTTP()
		case false:
			stpHTTP()
			return
		}
	}
}

func end2worker() {
	if eztools.Debugging && eztools.Verbose > 1 {
		eztools.Log("ending discovery")
	}
	var pckNt ezcomm.RoutCommStruc
	pckNt.Act = ezcomm.FlowChnEnd
	chnWeb[0] <- pckNt
}

func lanTabSwitched(reqUI bool, connIn *net.UDPConn,
	addrBrd *net.UDPAddr, bonjour []byte, defLanPrtStr string,
	makeLclInf func()) (connOut *net.UDPConn) {
	connOut = connIn
	switch reqUI {
	case false:
		if connIn != nil {
			end2worker()
			connOut = nil
			<-chnWeb[1] // wait for ezcomm to end
		}
		if httpSvr != nil {
			chnHTTP <- false
		}
		for i := range chnWeb {
			if chnWeb[i] == nil {
				continue
			}
			close(chnWeb[i])
			chnWeb[i] = nil
		}
	case true:
		for i := range chnWeb {
			if chnWeb[i] != nil {
				continue
			}
			chnWeb[i] = make(chan ezcomm.RoutCommStruc,
				ezcomm.FlowComLen)
		}
		if connIn != nil {
			sndUDP(connIn, addrBrd, bonjour)
			break
		}
		makeLclInf()
		connOut = runUDP(defLanPrtStr)
		if connOut == nil {
			break
		}
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("discovery", *connOut)
		}
		if httpSvr == nil {
			chnHTTP <- true
		}
	}

	return
}

func lanMsgFromNet(pckNt ezcomm.RoutCommStruc, connIn *net.UDPConn, bonjour []byte,
	bonjourLen, defLanPrtInt int, refreshPeerMap func(pckNt ezcomm.RoutCommStruc,
		peerHTTPPort int)) (connOut *net.UDPConn) {
	connOut = connIn
	switch pckNt.Act {
	case ezcomm.FlowChnEnd:
		end2worker()
		connOut = nil
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
		if !bytes.Equal(pckNt.Data[:bonjourLen], bonjour[:bonjourLen]) {
			break
		}
		var (
			peerHTTPPort int
			err          error
		)
		if len(pckNt.Data) > bonjourLen+1 {
			peerHTTPPort, err = strconv.Atoi(string(pckNt.Data[bonjourLen+1:]))
			if err != nil {
				if eztools.Debugging && eztools.Verbose > 1 {
					Log("not a port in greeting", err)
				}
			}
		}
		for _, ip := range localAddrSlc {
			if ip == pckNt.PeerUDP.IP.String() {
				//Log("echo", ip)
				return
			}
		}
		refreshPeerMap(pckNt, peerHTTPPort)
	}
	return
}

func lanListen() {
	var (
		conn  *net.UDPConn
		reqUI bool
		pckNt ezcomm.RoutCommStruc
	)
	defLanPrtInt := ezcomm.DefLanPrt()
	defLanPrtStr := strconv.Itoa(defLanPrtInt)
	var bonjour []byte
	for _, c := range ezcomm.EzcName {
		bonjour = append(bonjour, byte(c))
	}
	bonjourLen := len(bonjour)
	addrBrd, err := net.ResolveUDPAddr(ezcomm.StrUDP,
		net.JoinHostPort(ezcomm.DefBrdAdr, defLanPrtStr))
	if err != nil || addrBrd == nil {
		Log("bad broadcast addr", err)
		return
	}
	go switchHTTP()
	var localAddrMap, peerMap map[string]struct{}
	makeLclInf := func() {
		localAddrs, err := net.InterfaceAddrs()
		if err != nil || localAddrs == nil ||
			len(localAddrs) < 1 {
			Log("bad local addr", err)
			lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
			return
		}
		peerMap = make(map[string]struct{})
		localAddrMap = make(map[string]struct{})
		localAddrSlc = make([]string, 0)
		for _, localAddr1 := range localAddrs {
			if ipnet, ok := localAddr1.(*net.IPNet); ok {
				ip := ipnet.IP.String()
				localAddrMap[ip] = struct{}{}
				localAddrSlc = append(localAddrSlc, ip)
			}
		}
	}
	refreshPeerMap := func(pckNt ezcomm.RoutCommStruc,
		peerHTTPPort int) {
		peer := pckNt.PeerUDP.IP.String()
		if !eztools.Debugging || eztools.Verbose < 3 {
			// Filter local broadcast
			if _, ok := localAddrMap[peer]; ok {
				return
			}
		}
		if _, ok := peerMap[peer]; ok {
			// duplicate
			return
		}
		peerMap[peer] = struct{}{}
		add2Rmt(0, peer)
		sk := net.JoinHostPort(peer, strconv.Itoa(peerHTTPPort))
		lanLst.Append(sk)
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("discovered", sk)
		}
	}
	for {
		select {
		case reqUI = <-chnLan:
			conn = lanTabSwitched(reqUI, conn, addrBrd,
				bonjour, defLanPrtStr, makeLclInf)
			if !reqUI {
				return
			}
		case pckNt = <-chnWeb[1]:
			conn = lanMsgFromNet(pckNt, conn, bonjour, bonjourLen,
				defLanPrtInt, refreshPeerMap)
		}
	}
}

func makeTabLan(chnHTTPin chan bool) *container.TabItem {
	chnHTTP = chnHTTPin
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
		cp2Clip(net.JoinHostPort(localAddrSlc[id], lanPrtHTTP))
	}
	lanLbl = widget.NewLabel("")
	lanBut = widget.NewButton(ezcomm.StringTran["StrPokePeer"], func() {
		chnLan <- true
	})
	lanLst = widget.NewRadioGroup([]string{},
		func(sel string) {
			if ho, _, err := net.SplitHostPort(sel); err != nil {
				Log("failed to parse socket", sel, err)
			} else {
				cp2Clip(sel)
				sockRmt[0].SetText(ho)
			}
		})

	return container.NewTabItem(ezcomm.StringTran["StrInfLan"],
		container.NewBorder(widget.NewLabel(ezcomm.StringTran["StrLanHint"]),
			container.NewVBox(lanBut, lanLst), nil, nil, lanWeb))
}

func tabLanShown(yes bool) {
	if lanTabShown == yes {
		return
	}
	lanTabShown = yes
	switch yes {
	case true:
		lanBut.Refresh()
		go lanListen()
	case false:
	}
	chnLan <- yes
}
