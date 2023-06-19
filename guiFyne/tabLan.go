package main

import (
	"bytes"
	"net"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
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
	setLclSck(addr)
	setLclAddrs(localAddrSlc)
	lanWeb.Refresh()
	httpSvr = ezcomm.MakeHTTPSvr()
	httpSvr.GET(httpSmsDir, printIncomingWeb)
	httpSvr.POST(httpSmsDir, postIncomingWeb)
	httpSvr.FS(httpAppDir, "", httpFS(appStorage.RootURI().Path()))
	return httpSvr.Serve(lstnr)
}

func stpHTTP() error {
	lanPrtHTTP = ""
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
	if len(lanPrtHTTP) > 0 {
		for _, c := range lanPrtHTTP {
			bonjour = append(bonjour, byte(c))
		}
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

func end2worker() {
	if eztools.Debugging && eztools.Verbose > 1 {
		eztools.Log("ending discovery")
	}
	var pckNt ezcomm.RoutCommStruc
	pckNt.Act = ezcomm.FlowChnEnd
	chn[0] <- pckNt
}

func lanTabSwitched(reqUI bool, chnHTTP chan bool, connIn *net.UDPConn,
	addrBrd *net.UDPAddr, bonjour []byte, defLanPrtStr string,
	makeLclInf func()) (connOut *net.UDPConn) {
	connOut = connIn
	switch reqUI {
	case false:
		if connIn != nil {
			end2worker()
			connOut = nil
			<-chn[1] // wait for ezcomm to end
		}
		if httpSvr != nil {
			chnHTTP <- false
		}
	case true:
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
		refreshPeerMap(pckNt, peerHTTPPort)
	}
	return
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
	defLanPrtInt := ezcomm.DefLanPrt()
	defLanPrtStr := strconv.Itoa(defLanPrtInt)
	var bonjour []byte
	for _, c := range ezcomm.EzcName {
		bonjour = append(bonjour, byte(c))
	}
	bonjourLen := len(bonjour)
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
		// TODO: afix with [] to IPv6 like localAddrSlc
		peerMap[peer] = struct{}{}
		add2Rmt(0, peer)
		lanLst.Append(peer + ":" + strconv.Itoa(peerHTTPPort))
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("discovered", peer, ":", peerHTTPPort)
		}
	}
	for {
		select {
		case reqUI = <-chnLan:
			conn = lanTabSwitched(reqUI, chnHTTP, conn, addrBrd,
				bonjour, defLanPrtStr, makeLclInf)
		case pckNt = <-chn[1]:
			conn = lanMsgFromNet(pckNt, conn, bonjour, bonjourLen,
				defLanPrtInt, refreshPeerMap)
		}
	}
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
		cp2Clip(localAddrSlc[id] + lanPrtHTTP)
	}
	lanLbl = widget.NewLabel("")
	go lanListen(chnHTTP)
	lanBut = widget.NewButton(ezcomm.StringTran["StrPokePeer"], func() {
		chnLan <- true
	})
	lanLst = widget.NewRadioGroup([]string{},
		func(sel string) {
			cp2Clip(sel)
			// TODO: handle [] w/t or w/o port to IPv6 like localAddrSlc
			i := strings.LastIndex(sel, ":")
			sockRmt[0].SetText(sel[:i])
		})

	return container.NewTabItem(ezcomm.StringTran["StrInfLan"],
		container.NewBorder(widget.NewLabel(ezcomm.StringTran["StrLanHint"]),
			container.NewVBox(lanBut, lanLst), nil, nil, lanWeb))
}

func tabLanShown(yes bool) {
	switch yes {
	case true:
		lanBut.Refresh()
	case false:
	}
	chnLan <- yes
}
