package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
)

const (
	MaxRecLen = 10
)

var (
	fyneSockLcl, fyneSockRmt                              [2]*widget.SelectEntry
	fyneRecRcv, fyneRecSnd, fyneRowTcpSock2, fyneRowSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	fyneRowUdpSock2                    *fyne.Container
	fyneProt                           *widget.RadioGroup
	fyneLstBut, fyneDisBut, fyneSndBut *widget.Button
	fyneCntLcl, fyneCntRmt             *widget.Entry
	fyneRowLog                         *Entry // widget.Entry
)

func guiFyneEnable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		fyneSockLcl[i].Enable()
	}
	fyneProt.Enable()
	fyneLstBut.SetText(StrLst)
}

func guiFyneDisable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		fyneSockLcl[i].Disable()
	}
	fyneProt.Disable()
	fyneLstBut.SetText(StrStp)
}

func guiFyneLog(log2File bool, inf ...any) {
	str := fmt.Sprintf("%s%v\n", time.Now().Format("01-02 15:04:05"), inf)
	fyneRowLog.SetText(fyneRowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
		str)
	fyneRowLog.CursorRow++
	if log2File {
		eztools.LogWtTime(inf)
	}
}

func guiFyneSetLclSck(addr string) {
	ind := strings.LastIndex(addr, ":")
	fyneSockLcl[0].SetText(addr[:ind])
	fyneSockLcl[1].SetText(addr[ind+1:])
}

func guiFyneGetRmtSckStr() string {
	return fyneSockRmt[0].Text + ":" + fyneSockRmt[1].Text
}

func guiFyneGetRmtSck() *net.UDPAddr {
	ret, err := net.ResolveUDPAddr(fyneProt.Selected, guiFyneGetRmtSckStr())
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

func guiFyneSockF(str string) {
	if str != StrAll {
		fyneRecRcv.Options = recMap[str]
		return
	}
	fyneRecRcv.Options = recSlc
}

func guiFyneButSnd(snd bool) {
	if snd {
		fyneSndBut.OnTapped = guiFyneSnd
		fyneSndBut.SetText(StrSnd)
	} else {
		fyneSndBut.SetText(StrCon)
		fyneSndBut.OnTapped = guiFyneCon
	}
}

func guiFyneSckRmt(single bool) {
	if single {
		fyneRowUdpSock2.Hide()
		fyneRowTcpSock2.Show()
	} else {
		fyneRowUdpSock2.Show()
		fyneRowTcpSock2.Hide()
	}
}

func guiFyneLst() {
	var (
		err       error
		addrStruc net.Addr
	)
	addr := fyneSockLcl[0].Text + ":" + fyneSockLcl[1].Text
	guiFyneDisable()
	for i := range chanComm {
		chanComm[i] = make(chan RoutCommStruc, FlowComLen)
	}
	switch fyneProt.Selected {
	case StrUdp:
		var udpConn *net.UDPConn
		udpConn, err = ListenUdp(StrUdp, addr)
		if err != nil {
			break
		}
		if udpConn == nil {
			panic("NO connection got")
		}
		addrStruc = udpConn.LocalAddr()
		guiFyneSetLclSck(addrStruc.String())
		go ConnectedUdp(udpConn)
	case StrTcp:
		peeMap = make(map[string]chan RoutCommStruc)
		var lstnr net.Listener
		lstnr, err = ListenTcp(StrTcp, addr, ConnectedTcp, nil)
		if err != nil {
			break
		}
		addrStruc = lstnr.Addr()
		guiFyneSetLclSck(addrStruc.String())
		guiFyneButSnd(true)
		fyneSndBut.Disable()
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
		fyneLstBut.OnTapped = guiFyneStp
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
	/*conn*/ _, err := Client(fyneProt.Selected, pr, ConnectedTcp)
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
	fyneSndBut.OnTapped = guiFyneSnd
	if chanComm[0] == nil { //client
		fyneLstBut.Hide()
		fyneDisBut.Show()
	}
	guiFyneButSnd(true)
	fyneDisBut.Show()
	fyneSndBut.Enable()
	guiFyneSetLclSck(lcl)
	fyneRowTcpSock2.Options = append(fyneRowTcpSock2.Options, rmt)
	if len(fyneRowTcpSock2.Selected) < 1 {
		fyneRowTcpSock2.SetSelectedIndex(0)
	}
	guiFyneLog(true, lcl, "<->", rmt, "connected")
	fyneRowTcpSock2.Refresh()
}

// guiFyneDisconnected is for TCP only
func guiFyneDisconnected(rmt string) {
	indx := -1
	ln := len(fyneRowTcpSock2.Options)
	for i, v := range fyneRowTcpSock2.Options {
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
		fyneRowTcpSock2.Options = nil
		fyneRowTcpSock2.Selected = ""
		fyneRowTcpSock2.Refresh()
		fyneDisBut.Hide()
		if fyneLstBut.Hidden { //client
			fyneLstBut.Show()
			guiFyneEnable()
			//guiFyneSckRmt(false)
			guiFyneButSnd(false)
			fyneSndBut.Enable()
			guiFyneLog(true, rmt, "disconnected")
		} else { //server
			fyneSndBut.Disable()
			guiFyneLog(true, rmt, "disconnected", ". server idle.")
		}
		return
	}
	// reorder the records
	if indx != ln-1 {
		fyneRowTcpSock2.Options[indx] = fyneRowTcpSock2.Options[ln-1]
	}
	fyneRowTcpSock2.Options = fyneRowTcpSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		guiFyneLog(true, "clients left", fyneRowTcpSock2.Options)
	}
	fyneRowTcpSock2.SetSelectedIndex(0)
	fyneRowTcpSock2.Refresh()
}

// guiFyneDis disconnect 1 peer TCP
func guiFyneDis() {
	rmtTcp := fyneRowTcpSock2.Selected
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
	if !fyneDisBut.Hidden { //clients still running
		fyneLstBut.Hide()
	} else {
		guiFyneSckRmt(false)
		guiFyneEnable()
		guiFyneButSnd(false)
	}
	fyneLstBut.OnTapped = guiFyneLst
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
	if _, ok := sndMap[indx][txt]; ok {
		return
	}
	sndMap[indx][txt] = struct{}{}
	snd2Slc[indx] = append(snd2Slc[indx], txt)
	fyneSockRmt[indx].SetOptions(snd2Slc[indx])
}

func guiFyneSnd() {
	var (
		rmtUdp *net.UDPAddr
		rmtTcp string
	)
	if !fyneProt.Disabled() {
		switch fyneProt.Selected {
		case StrUdp: // listen before sending
			rmtUdp = guiFyneGetRmtSck()
			if rmtUdp == nil {
				return
			}
			for i := 0; i < 2; i++ {
				guiFyneAdd2Rmt(i, fyneSockRmt[i].Text)
			}
			guiFyneLst()
			if !fyneProt.Disabled() { // failed
				return
			}
		case StrTcp:
			panic("NOT listening when sending")
		}
	}
	switch fyneProt.Selected {
	case StrUdp: // listen before sending
		rmtUdp = guiFyneGetRmtSck()
		if rmtUdp == nil {
			return
		}
		chanComm[0] <- RoutCommStruc{
			Act:     FlowChnSnd,
			Data:    fyneCntLcl.Text,
			PeerUdp: rmtUdp,
		}
	case StrTcp:
		rmtTcp = fyneRowTcpSock2.Selected
		chn, ok := peeMap[rmtTcp]
		if !ok {
			guiFyneLog(true, "NO peer found for", rmtTcp)
			break
		}
		chn <- RoutCommStruc{
			Act:  FlowChnSnd,
			Data: fyneCntLcl.Text,
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
		fyneCntRmt.SetText(comm.Data)
		var peer string
		if comm.PeerUdp != nil {
			peer = comm.PeerUdp.String()
			ind := strings.LastIndex(peer, ":")
			guiFyneAdd2Rmt(0, peer[:ind])
			guiFyneAdd2Rmt(1, peer[ind+1:])
		} else if comm.PeerTcp != nil {
			peer = comm.PeerTcp.String()
		} else {
			guiFyneLog(true, "got from somewhere")
		}
		if len(peer) > 0 {
			if _, ok := rcvMap[peer]; !ok {
				rcvMap[peer] = struct{}{}
				fyneRowSockF.Options = append(fyneRowSockF.Options, peer)
			}
			fyneRowSockF.SetSelected(peer)
			fyneRowSockF.Refresh()
			recMap[peer] = append(recMap[peer], comm.Data)
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.LogWtTime("<-", peer, comm.Data)
			}
			guiFyneLog(true, "got from", peer)
		}
		fyneRecRcv.Options = append(fyneRecRcv.Options, comm.Data)
		recSlc = append(recSlc, comm.Data)
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
		fyneRecSnd.Options = append(fyneRecSnd.Options, comm.Data)
		if comm.PeerUdp != nil {
			addrStr := comm.PeerUdp.String()
			ind := strings.LastIndex(addrStr, ":")
			fyneSockRmt[0].SetText(addrStr[:ind])
			fyneSockRmt[1].SetText(addrStr[ind+1:])
		}
	}
}

func guiFyneMakeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(StrLcl))

	addrLbl := widget.NewLabel(StrAdr)
	portLbl := widget.NewLabel(StrPrt)
	for i := 0; i < 2; i++ {
		fyneSockLcl[i] = widget.NewSelectEntry(nil)
	}
	fyneSockLcl[0].PlaceHolder = DefAdr
	rowSock := container.NewGridWithRows(2,
		addrLbl, fyneSockLcl[0], portLbl, fyneSockLcl[1])

	fyneProt = widget.NewRadioGroup([]string{StrUdp, StrTcp}, nil)
	fyneProt.Horizontal = true
	fyneProt.SetSelected("udp")
	fyneLstBut = widget.NewButton(StrLst, guiFyneLst)
	fyneDisBut = widget.NewButton(StrDis, guiFyneDis)
	fyneDisBut.Hide()
	rowProt := container.NewHBox(fyneProt, fyneLstBut, fyneDisBut)
	fyneProt.OnChanged = func(str string) {
		switch str {
		case StrUdp:
			guiFyneButSnd(true)
		case StrTcp:
			guiFyneButSnd(false)
			//if len(sockRmt[1].Text) > 0 {
			/*} else {
				//sndBut.Text = STR_SND
				sndBut.OnTapped = guiFyneSnd
			}*/
		}
		fyneSndBut.Refresh()
	}

	recLbl := container.NewCenter(widget.NewLabel(StrRec))
	fyneRecSnd = widget.NewSelect(nil, func(str string) {
		fyneCntLcl.SetText(str)
		escaped := url.PathEscape(str)
		if len(escaped) > MaxRecLen {
			fyneRecSnd.Selected = escaped[0:MaxRecLen]
			fyneRecSnd.Refresh()
		} else {
			if escaped != str {
				fyneRecSnd.Selected = escaped
				fyneRecSnd.Refresh()
			}
		}
	})
	cntLbl := container.NewCenter(widget.NewLabel(StrCnt))
	rowRec := container.NewGridWithRows(3, recLbl, fyneRecSnd, cntLbl)

	fyneCntLcl = widget.NewMultiLineEntry()

	fyneSndBut = widget.NewButton(StrSnd, guiFyneSnd)
	fyneSndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, fyneSndBut, nil, nil, fyneCntLcl)
}

func guiFyneMakeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(StrRmt))
	rowTo := container.NewCenter(widget.NewLabel(StrTo))

	for i := 0; i < 2; i++ {
		fyneSockRmt[i] = widget.NewSelectEntry(nil)
	}
	fyneSockRmt[0].PlaceHolder = StrLcl
	fyneSockRmt[1].OnChanged = func(str string) {
		if len(str) > 0 { //&& len(sockRmt[0].Text) > -1 {
			fyneSndBut.Enable()
		} else {
			fyneSndBut.Disable()
		}
		fyneSndBut.Refresh()
	}
	fyneRowUdpSock2 = container.NewGridWithColumns(2, fyneSockRmt[0], fyneSockRmt[1])
	fyneRowTcpSock2 = widget.NewSelect(nil, func(str string) {
	})
	fyneRowTcpSock2.Hide()

	rowFrm := container.NewCenter(widget.NewLabel(StrFrm))

	for i := 0; i < 2; i++ {
		sndMap[i] = make(map[string]struct{})
	}
	rcvMap = make(map[string]struct{})
	fyneRowSockF = widget.NewSelect(nil, guiFyneSockF)
	fyneRowSockF.Options = []string{StrAll}

	recLbl := container.NewCenter(widget.NewLabel(StrRec))
	fyneRecRcv = widget.NewSelect(nil, func(str string) {
		fyneCntRmt.SetText(str)
		escaped := url.PathEscape(str)
		if len(escaped) > MaxRecLen {
			fyneRecRcv.Selected = escaped[0:MaxRecLen]
			fyneRecRcv.Refresh()
		} else {
			if escaped != str {
				fyneRecRcv.Selected = escaped
				fyneRecRcv.Refresh()
			}
		}
	})
	recMap = make(map[string][]string)
	cntLbl := container.NewCenter(widget.NewLabel(StrCnt))
	rowRec := container.NewGridWithRows(3, recLbl, fyneRecRcv, cntLbl)

	fyneCntRmt = widget.NewMultiLineEntry()

	fyneRowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	fyneRowLog.Disable()

	tops := container.NewVBox(rowLbl, rowTo, fyneRowUdpSock2, fyneRowTcpSock2,
		rowFrm, fyneRowSockF, rowRec)
	return container.NewBorder(tops, fyneRowLog, nil, nil, fyneCntRmt)
}
