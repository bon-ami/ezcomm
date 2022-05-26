package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
)

func guiFyne() {
	ezcApp := app.New()
	/*icon, err := LoadResourceFromPath("icon.ico")
	if err == nil {
		ezcApp.SetIcon(icon)
	}*/
	ezcWin := ezcApp.NewWindow(ezcName)

	contLcl := guiFyneMakeControlsLcl()
	contRmt := guiFyneMakeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)
	//lay.Layout(prot)

	ezcWin.SetContent(cont) //prot)
	//ezcWin.SetContent(widget.NewLabel("Hello"))

	/*selectCtrls = []*widget.SelectEntry{ sockLcl[0], sockLcl[1], //sockRmt[0], sockRmt[1],
		//sockRcv[0], sockRcv[0], //recRmt,
	}*/

	ezcWin.Show()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("to show UI")
	}
	ezcApp.Run()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("UI done")
	}
}

var (
	sockLcl, sockRmt, sockRcv                [2]*widget.SelectEntry
	recRmt, recInp, rowTcpSock2, rowTcpSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	rowUdpSock2, rowUdpSockF *fyne.Container
	prot                     *widget.RadioGroup
	lstBut, disBut, sndBut   *widget.Button
	cntLcl, cntRmt           *widget.Entry
	rowLog                   *Entry // widget.Entry
)

const (
	strUdp = "udp"
	strTcp = "tcp"
)

func guiFyneEnable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Enable()
	}
	prot.Enable()
	lstBut.SetText(STR_LST)
}

func guiFyneDisable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Disable()
	}
	prot.Disable()
	lstBut.SetText(STR_STP)
}

func guiFyneLog(log2File bool, inf ...any) {
	str := fmt.Sprintf("%s%v\n", time.Now().Format("01-02 15:04:05"), inf)
	rowLog.SetText(rowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
		str)
	rowLog.CursorRow++
	if log2File {
		eztools.LogWtTime(inf)
	}
}

func guiFyneSetLclSck(addr string) {
	ind := strings.LastIndex(addr, ":")
	sockLcl[0].SetText(addr[:ind])
	sockLcl[1].SetText(addr[ind+1:])
}

func guiFyneGetRmtSckStr() string {
	return sockRmt[0].Text + ":" + sockRmt[1].Text
}

func guiFyneGetRmtSck() *net.UDPAddr {
	ret, err := net.ResolveUDPAddr(prot.Selected, guiFyneGetRmtSckStr())
	if err != nil {
		guiFyneLog(false, err)
		return nil
	}
	return ret
}

/*func guiFyneGetLclSck() *net.UDPAddr {
	ret, err := net.ResolveUDPAddr(prot.Selected, sockLcl[0].Text+":"+sockLcl[1].Text)
	if err != nil {
		guiFyneLog(false, err)
		return nil
	}
	return ret
}*/

func guiFyneButSnd(snd bool) {
	if snd {
		sndBut.OnTapped = guiFyneSnd
		sndBut.Text = STR_SND
	} else {
		sndBut.Text = STR_CON
		sndBut.OnTapped = guiFyneCon
	}
	sndBut.Refresh()
}

func guiFyneSckRmt(single bool) {
	if single {
		rowUdpSock2.Hide()
		rowTcpSock2.Show()
	} else {
		rowUdpSock2.Show()
		rowTcpSock2.Hide()
	}
}

func guiFyneLst() {
	var (
		err       error
		addrStruc net.Addr
	)
	addr := sockLcl[0].Text + ":" + sockLcl[1].Text
	guiFyneDisable()
	for i := range chanComm {
		chanComm[i] = make(chan RoutCommStruc, FlowComLen)
	}
	switch prot.Selected {
	case strUdp:
		var udpConn *net.UDPConn
		udpConn, err = ListenUdp(strUdp, addr)
		if err != nil {
			break
		}
		if udpConn == nil {
			panic("NO connection got")
		}
		addrStruc = udpConn.LocalAddr()
		guiFyneSetLclSck(addrStruc.String())
		go ConnectedUdp(udpConn)
	case strTcp:
		peeMap = make(map[string]chan RoutCommStruc)
		var lstnr net.Listener
		lstnr, err = ListenTcp(strTcp, addr, ConnectedTcp, nil)
		if err != nil {
			break
		}
		addrStruc = lstnr.Addr()
		guiFyneSetLclSck(addrStruc.String())
		guiFyneButSnd(true)
		sndBut.Disable()
		guiFyneSckRmt(true)
		go ListeningTcp(lstnr)
	}
	if err != nil {
		for i := range chanComm {
			chanComm[i] = nil
		}
		guiFyneEnable()
		guiFyneLog(true, err)
	} else {
		lstBut.OnTapped = guiFyneStp
		guiFyneLog(true, "listening on", addrStruc.String())
	}
}

// guiFyneCon is for TCP only
func guiFyneCon() {
	pr := guiFyneGetRmtSckStr()
	peeMap = make(map[string]chan RoutCommStruc)
	chn := make(chan RoutCommStruc, FlowComLen)
	peeMap[pr] = chn
	guiFyneDisable()
	/*conn*/ _, err := Client(prot.Selected, pr, ConnectedTcp)
	if err != nil {
		guiFyneEnable()
		guiFyneLog(true, "fail to connect to "+pr, err)
		return
	}
	//lstBut.OnTapped = guiFyneStp
	guiFyneSckRmt(true)
	//guiFyneConnected(conn.LocalAddr().String(), conn.RemoteAddr().String())
}

// guiFyneConnected is for TCP only
func guiFyneConnected(lcl, rmt string) {
	sndBut.OnTapped = guiFyneSnd
	if chanComm[0] == nil { //client
		lstBut.Hide()
		disBut.Show()
	}
	guiFyneButSnd(true)
	disBut.Show()
	sndBut.Enable()
	guiFyneSetLclSck(lcl)
	rowTcpSock2.Options = append(rowTcpSock2.Options, rmt)
	if len(rowTcpSock2.Selected) < 1 {
		rowTcpSock2.SetSelectedIndex(0)
	}
	guiFyneLog(true, lcl, "<->", rmt, "connected")
	rowTcpSock2.Refresh()
}

// guiFyneDisconnected is for TCP only
func guiFyneDisconnected(rmt string) {
	indx := -1
	ln := len(rowTcpSock2.Options)
	for i, v := range rowTcpSock2.Options {
		if v == rmt {
			indx = i
			break
		}
	}
	switch {
	case indx == -1:
		guiFyneLog(true, "disconnected", rmt, "NOT in record!")
		return
	case ln == 1:
		rowTcpSock2.Options = nil
		rowTcpSock2.Selected = ""
		rowTcpSock2.Refresh()
		disBut.Hide()
		if lstBut.Hidden { //client
			lstBut.Show()
			guiFyneEnable()
			//guiFyneSckRmt(false)
			guiFyneButSnd(false)
			sndBut.Enable()
			guiFyneLog(true, rmt, "disconnected")
		} else { //server
			sndBut.Disable()
			guiFyneLog(true, rmt, "disconnected", ". server idle.")
		}
		return
	}
	// reorder the records
	if indx != ln-1 {
		rowTcpSock2.Options[indx] = rowTcpSock2.Options[ln-1]
	}
	rowTcpSock2.Options = rowTcpSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		guiFyneLog(true, "clients left", rowTcpSock2.Options)
	}
	rowTcpSock2.SetSelectedIndex(0)
	rowTcpSock2.Refresh()
}

// guiFyneDis disconnect 1 peer TCP
func guiFyneDis() {
	rmtTcp := rowTcpSock2.Selected
	chn, ok := peeMap[rmtTcp]
	if !ok {
		guiFyneLog(true, "NO peer found for", rmtTcp)
		return
	}
	guiFyneLog(true, "disconnecting", rmtTcp)
	chn <- RoutCommStruc{
		Act: FlowChnEnd,
	}
	guiFyneDisconnected(rmtTcp)
}

func guiFyneStp() {
	chanComm[0] <- RoutCommStruc{
		Act: FlowChnEnd,
	}
	if !disBut.Hidden { //clients still running
		lstBut.Hide()
	} else {
		guiFyneSckRmt(false)
		guiFyneEnable()
		guiFyneButSnd(false)
	}
	lstBut.OnTapped = guiFyneLst
	guiFyneLog(true, "stopped listening")
}

func guiFyneAdd2Rmt(indx int, txt string) {
	if len(txt) < 1 {
		return
	}
	/*if _, ok := sndMap[indx][txt]; ok {
		return
	}
	sndMap[indx][txt] = struct{}{}*/
	snd2Slc[indx] = append(snd2Slc[indx], txt)
	sockRmt[indx].SetOptions(snd2Slc[indx])
}

func guiFyneSnd() {
	var (
		rmtUdp *net.UDPAddr
		rmtTcp string
	)
	if !prot.Disabled() {
		switch prot.Selected {
		case strUdp: // listen before sending
			rmtUdp = guiFyneGetRmtSck()
			if rmtUdp == nil {
				return
			}
			for i := 0; i < 2; i++ {
				guiFyneAdd2Rmt(i, sockRmt[i].Text)
			}
			guiFyneLst()
			if !prot.Disabled() { // failed
				return
			}
		case strTcp:
			panic("NOT listening when sending")
		}
	}
	switch prot.Selected {
	case strUdp: // listen before sending
		rmtUdp = guiFyneGetRmtSck()
		if rmtUdp == nil {
			return
		}
		chanComm[0] <- RoutCommStruc{
			Act:     FlowChnSnd,
			Data:    cntLcl.Text,
			PeerUdp: rmtUdp,
		}
	case strTcp:
		rmtTcp = rowTcpSock2.Selected
		chn, ok := peeMap[rmtTcp]
		if !ok {
			guiFyneLog(true, "NO peer found for", rmtTcp)
			break
		}
		chn <- RoutCommStruc{
			Act:  FlowChnSnd,
			Data: cntLcl.Text,
		}
	}
}

func guiFyneRcv(comm RoutCommStruc) {
	switch comm.Act {
	case FlowChnRcv:
		if comm.Err != nil {
			guiFyneLog(true, "failed to receive", comm.Err)
			break
		}
		cntRmt.Text = comm.Data
		cntRmt.Refresh()
		if comm.PeerUdp != nil {
			addrStr := comm.PeerUdp.String()
			ind := strings.LastIndex(addrStr, ":")
			var sockStrs [2]string
			sockStrs[0] = addrStr[:ind]
			sockStrs[1] = addrStr[ind+1:]
			for i := 0; i < 2; i++ {
				if len(sockStrs[0]) < 1 {
					continue
				}
				guiFyneAdd2Rmt(i, sockStrs[i])
				if _, ok := rcvMap[i][sockStrs[i]]; ok {
					continue
				}
				rcvMap[i][sockStrs[i]] = struct{}{}
				rcvSlc[i] = append(rcvSlc[i], sockStrs[i])
				sockRcv[i].SetOptions(rcvSlc[i])
				sockRcv[i].Text = sockStrs[i]
				sockRcv[i].Refresh()
			}
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.LogWtTime("<-", comm.PeerUdp.String(), comm.Data)
			}
			guiFyneLog(true, "got from", comm.PeerUdp.String())
		} else if comm.PeerTcp != nil {
			peer := comm.PeerTcp.String()
			if _, ok := rcvMap[0][peer]; !ok {
				rcvMap[0][peer] = struct{}{}
				rowTcpSockF.Options = append(rowTcpSockF.Options, peer)
			}
			rowTcpSockF.Selected = peer
			rowTcpSockF.Refresh()
			guiFyneLog(true, "got from", peer)
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.LogWtTime("<-", peer, comm.Data)
			}
		} else {
			guiFyneLog(true, "got from somewhere")
		}
		recRmt.Options = append(recRmt.Options, comm.Data)
	}
}

// guiFyneEnded is run when TCP peer disconnected
func guiFyneEnded(comm RoutCommStruc) {
	switch comm.Act {
	case FlowChnEnd:
		if comm.PeerTcp != nil {
			peer := comm.PeerTcp.String()
			if ch, ok := peeMap[peer]; !ok {
				guiFyneLog(true, "UNKNOWN peer disconnected", peer)
			} else {
				ch <- comm
			}
			guiFyneDisconnected(peer)
		}
	}
}

func guiFyneSnt(comm RoutCommStruc) {
	switch comm.Act {
	case FlowChnSnd:
		if comm.Err != nil {
			guiFyneLog(true, "failed to send", comm.Err)
			break
		}
		var peer string
		switch {
		case comm.PeerUdp != nil:
			peer = comm.PeerUdp.String()
		case comm.PeerTcp != nil:
			peer = comm.PeerTcp.String()
		default:
			peer = "somewhere"
		}
		if eztools.Debugging && eztools.Verbose > 2 {
			eztools.LogWtTime(">-", peer, comm.Data)
		}
		guiFyneLog(true, "sent to", peer)
		recInp.Options = append(recInp.Options, comm.Data)
		if comm.PeerUdp != nil {
			addrStr := comm.PeerUdp.String()
			ind := strings.LastIndex(addrStr, ":")
			sockRmt[0].SetText(addrStr[:ind])
			sockRmt[1].SetText(addrStr[ind+1:])
		}
	}
}

func guiFyneMakeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(STR_LCL))

	addrLbl := widget.NewLabel(STR_ADR)
	portLbl := widget.NewLabel(STR_PRT)
	for i := 0; i < 2; i++ {
		sockLcl[i] = widget.NewSelectEntry(nil)
	}
	sockLcl[0].PlaceHolder = DEF_ADR
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	prot = widget.NewRadioGroup([]string{strUdp, strTcp}, nil)
	prot.Horizontal = true
	prot.SetSelected("udp")
	lstBut = widget.NewButton(STR_LST, guiFyneLst)
	disBut = widget.NewButton(STR_DIS, guiFyneDis)
	disBut.Hide()
	rowProt := container.NewHBox(prot, lstBut, disBut)
	prot.OnChanged = func(str string) {
		switch str {
		case strUdp:
			guiFyneButSnd(true)
			rowUdpSockF.Show()
			rowTcpSockF.Hide()
		case strTcp:
			guiFyneButSnd(false)
			rowUdpSockF.Hide()
			rowTcpSockF.Show()
			//if len(sockRmt[1].Text) > 0 {
			/*} else {
				//sndBut.Text = STR_SND
				sndBut.OnTapped = guiFyneSnd
			}*/
		}
		sndBut.Refresh()
	}

	recLbl := container.NewCenter(widget.NewLabel(STR_REC))
	recInp = widget.NewSelect(nil, func(str string) {
		cntLcl.Text = str
		cntLcl.Refresh()
	})
	cntLbl := container.NewCenter(widget.NewLabel(STR_CNT))
	rowRec := container.NewGridWithRows(3, recLbl, recInp, cntLbl)

	cntLcl = widget.NewMultiLineEntry()

	sndBut = widget.NewButton(STR_SND, guiFyneSnd)
	sndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntLcl)
}

func guiFyneMakeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(STR_RMT))
	rowTo := container.NewCenter(widget.NewLabel(STR_TO))

	for i := 0; i < 2; i++ {
		sockRmt[i] = widget.NewSelectEntry(nil)
		//sndMap[i] = make(map[string]struct{})
	}
	sockRmt[0].PlaceHolder = STR_LCL
	/*sockRmt[0].OnChanged = func(str string) {
		//guiFyneLog(false, str, sockRmt[1].Text, sockRmt[1].SelectedText(), sockLcl[1].Disabled())
		if len(str) > 0 && len(sockRmt[1].Text) > 0 {
			sndBut.Enable()
		} else {
			sndBut.Disable()
		}
		sndBut.Refresh()
	}*/
	sockRmt[1].OnChanged = func(str string) {
		if len(str) > 0 { //&& len(sockRmt[0].Text) > -1 {
			sndBut.Enable()
		} else {
			sndBut.Disable()
		}
		sndBut.Refresh()
	}
	rowUdpSock2 = container.NewGridWithColumns(2, sockRmt[0], sockRmt[1])
	rowTcpSock2 = widget.NewSelect(nil, func(str string) {
	})
	rowTcpSock2.Hide()

	rowFrm := container.NewCenter(widget.NewLabel(STR_FRM))

	for i := 0; i < 2; i++ {
		sockRcv[i] = widget.NewSelectEntry(nil)
		rcvMap[i] = make(map[string]struct{})
	}
	rowUdpSockF = container.NewGridWithColumns(2, sockRcv[0], sockRcv[1])
	rowTcpSockF = widget.NewSelect(nil, func(str string) {
	})
	rowTcpSockF.Hide()

	recLbl := container.NewCenter(widget.NewLabel(STR_REC))
	recRmt = widget.NewSelect(nil, func(str string) {
		cntRmt.Text = str
	})
	cntLbl := container.NewCenter(widget.NewLabel(STR_CNT))
	rowRec := container.NewGridWithRows(3, recLbl, recRmt, cntLbl)

	cntRmt = widget.NewMultiLineEntry()

	rowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	rowLog.Disable()

	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2,
		rowFrm, rowUdpSockF, rowTcpSockF, rowRec)
	return container.NewBorder(tops, rowLog, nil, nil, cntRmt)
}
