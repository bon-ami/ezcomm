package main

import (
	"bytes"
	"net"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	chnLan chan bool
	lanBut *widget.Button
	lanLbl *widget.Label
	lanLst *widget.RadioGroup
)

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
	addrBrd, err := net.ResolveUDPAddr(ezcomm.StrUdp, ezcomm.DefBrd+":"+defLanS)
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
				if conn == nil {
					break
				}
				end2worker()
				<-chn[1] // wait for ezcomm to end
			case true:
				if eztools.Debugging && eztools.Verbose > 1 {
					eztools.Log("discovery", conn)
				}
				if conn != nil {
					pckNt.Act = ezcomm.FlowChnSnd
					pckNt.Data = bonjour
					pckNt.PeerUdp = addrBrd
					chn[0] <- pckNt
					break
				}
				localAddrs, err := net.InterfaceAddrs()
				if err != nil || localAddrs == nil || len(localAddrs) < 1 {
					Log("bad local addr", err)
					lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
					break
				}
				localAddrMap = make(map[string]struct{})
				for _, localAddr1 := range localAddrs {
					str := localAddr1.String()
					ind := strings.LastIndex(str, "/")
					localAddrMap[str[:ind]] = struct{}{}
				}
				conn, err = ezcomm.ListenUdp(ezcomm.StrUdp,
					":"+defLanS)
				if err != nil {
					lanLbl.SetText(ezcomm.StringTran["StrDiscoverFail"])
					Log("discovery failure", err)
					break
				}
				lanLbl.SetText(ezcomm.StringTran["StrLst"])
				peerMap = make(map[string]struct{})
				go ezcomm.ConnectedUdp(Log, chn, conn)
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
					Log("discovery failure", "NO peer addr!")
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

func makeTabLan() *container.TabItem {
	chnLan = make(chan bool, 1)
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
		container.NewVBox(lanBut, lanLst))
}

func tabLanShown(yes bool) {
	switch yes {
	case true:
		lanBut.Refresh()
	case false:
	}
	chnLan <- yes
}
