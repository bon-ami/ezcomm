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
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	chnLan       chan bool
	httpSvr      *ezcomm.HTTPSvr
	localAddrSlc []string
	lanPrtHTTP   string
	lanWeb       *widget.List
	//lanWeb     *widget.TextGrid
	lanBut *widget.Button
	lanLbl *widget.Label
	lanLst *widget.RadioGroup
)

func httpListNum() int {
	return len(localAddrSlc)
}

func httpListUpdate(id widget.ListItemID, item fyne.CanvasObject) {
	item.(*widget.Label).SetText(localAddrSlc[id] + lanPrtHTTP)
}

func lanListen() {
	var (
		conn  *net.UDPConn
		reqUI bool
		pckNt ezcomm.RoutCommStruc
		chn   [2]chan ezcomm.RoutCommStruc
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
	defLanI := ezcomm.DefLan()
	defLanS := strconv.Itoa(defLanI)
	var bonjour []byte
	for _, c := range ezcomm.EzcName {
		bonjour = append(bonjour, byte(c))
	}
	addrBrd, err := net.ResolveUDPAddr(ezcomm.StrUdp,
		ezcomm.DefBrd+":"+defLanS)
	if err != nil || addrBrd == nil {
		Log("bad broadcast addr", err)
		return
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
					httpSvr.Shutdown(1)
					httpSvr = nil
				}
			case true:
				if conn != nil {
					if eztools.Debugging &&
						eztools.Verbose > 1 {
						eztools.Log("discover", *conn)
					}
					pckNt.Act = ezcomm.FlowChnSnd
					pckNt.Data = bonjour
					pckNt.PeerUdp = addrBrd
					chn[0] <- pckNt
					break
				}
				localAddrs, err := net.InterfaceAddrs()
				if err != nil || localAddrs == nil ||
					len(localAddrs) < 1 {
					Log("bad local addr", err)
					lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
					break
				}
				localAddrMap = make(map[string]struct{})
				localAddrSlc = make([]string, 0)
				for _, localAddr1 := range localAddrs {
					str := localAddr1.String()
					ind := strings.LastIndex(str, "/")
					localAddrMap[str[:ind]] = struct{}{}
					localAddrSlc = append(localAddrSlc, str[:ind])
				}
				conn, err = ezcomm.ListenUDP(ezcomm.StrUdp,
					":"+defLanS)
				if err != nil || conn == nil {
					lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
					Log("discovery failure", err)
					break
				}
				if eztools.Debugging && eztools.Verbose > 1 {
					eztools.Log("discovery", *conn)
				}
				lanLbl.SetText(ezcomm.StringTran["StrLst"])
				peerMap = make(map[string]struct{})
				go ezcomm.ConnectedUDP(Log, chn, conn)
				if httpSvr == nil {
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
					httpSvr.FS("", "",
						httpFS(appStorage.RootURI().Path()))
					httpSvr.Serve(lstnr)
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
				if pckNt.PeerUdp == nil {
					Log("discovery failure. NO peer addr!")
					break
				}
				if pckNt.PeerUdp.Port != defLanI {
					break
				}
				if !bytes.Equal(pckNt.Data, bonjour) {
					break
				}
				peer := pckNt.PeerUdp.IP.String()
				_, ok := localAddrMap[peer]
				if ok {
					break
				}
				if _, ok := peerMap[peer]; ok {
					// duplicate
					break
				}
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

func makeTabLan() *container.TabItem {
	chnLan = make(chan bool, 1)
	//lanWeb = widget.NewTextGrid()
	lanWeb = widget.NewList(
		httpListNum,
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		httpListUpdate)
	lanWeb.OnSelected = func(id widget.ListItemID) {
		drv := ezcApp.Driver()
		clipboard := drv.AllWindows()[0].Clipboard()
		clipboard.SetContent(localAddrSlc[id] + lanPrtHTTP)
		toast("StrCopied", localAddrSlc[id]+lanPrtHTTP)
	}
	lanLbl = widget.NewLabel("")
	go lanListen()
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
