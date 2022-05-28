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
	sockLcl, sockRmt                      [2]*widget.SelectEntry
	recRcv, recSnd, rowTcpSock2, rowSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	rowUdpSock2            *fyne.Container
	prot                   *widget.RadioGroup
	lstBut, disBut, sndBut *widget.Button
	cntLcl, cntRmt         *widget.Entry
	rowLog                 *Entry // widget.Entry
)

func guiFyneEnable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Enable()
	}
	prot.Enable()
	lstBut.SetText(StrLst)
}

func guiFyneDisable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Disable()
	}
	prot.Disable()
	lstBut.SetText(StrStp)
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

func guiFyneSockF(str string) {
	if str != StrAll {
		recRcv.Options = recMap[str]
		return
	}
	recRcv.Options = recSlc
}

func guiFyneButSnd(snd bool) {
	if snd {
		sndBut.OnTapped = guiFyneSnd
		sndBut.SetText(StrSnd)
	} else {
		sndBut.SetText(StrCon)
		sndBut.OnTapped = guiFyneCon
	}
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
	if _, ok := sndMap[indx][txt]; ok {
		return
	}
	sndMap[indx][txt] = struct{}{}
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
		case StrUdp: // listen before sending
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
		case StrTcp:
			panic("NOT listening when sending")
		}
	}
	switch prot.Selected {
	case StrUdp: // listen before sending
		rmtUdp = guiFyneGetRmtSck()
		if rmtUdp == nil {
			return
		}
		chanComm[0] <- RoutCommStruc{
			Act:     FlowChnSnd,
			Data:    cntLcl.Text,
			PeerUdp: rmtUdp,
		}
	case StrTcp:
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
		cntRmt.SetText(comm.Data)
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
				rowSockF.Options = append(rowSockF.Options, peer)
			}
			rowSockF.SetSelected(peer)
			rowSockF.Refresh()
			recMap[peer] = append(recMap[peer], comm.Data)
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.LogWtTime("<-", peer, comm.Data)
			}
			guiFyneLog(true, "got from", peer)
		}
		recRcv.Options = append(recRcv.Options, comm.Data)
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
		recSnd.Options = append(recSnd.Options, comm.Data)
		if comm.PeerUdp != nil {
			addrStr := comm.PeerUdp.String()
			ind := strings.LastIndex(addrStr, ":")
			sockRmt[0].SetText(addrStr[:ind])
			sockRmt[1].SetText(addrStr[ind+1:])
		}
	}
}

func guiFyneMakeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(StrLcl))

	addrLbl := widget.NewLabel(StrAdr)
	portLbl := widget.NewLabel(StrPrt)
	for i := 0; i < 2; i++ {
		sockLcl[i] = widget.NewSelectEntry(nil)
	}
	sockLcl[0].PlaceHolder = DefAdr
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	prot = widget.NewRadioGroup([]string{StrUdp, StrTcp}, nil)
	prot.Horizontal = true
	prot.SetSelected("udp")
	lstBut = widget.NewButton(StrLst, guiFyneLst)
	disBut = widget.NewButton(StrDis, guiFyneDis)
	disBut.Hide()
	rowProt := container.NewHBox(prot, lstBut, disBut)
	prot.OnChanged = func(str string) {
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
		sndBut.Refresh()
	}

	recLbl := container.NewCenter(widget.NewLabel(StrRec))
	recSnd = widget.NewSelect(nil, func(str string) {
		cntLcl.SetText(str)
		escaped := url.PathEscape(str)
		if len(escaped) > MaxRecLen {
			recSnd.Selected = escaped[0:MaxRecLen]
			recSnd.Refresh()
		} else {
			if escaped != str {
				recSnd.Selected = escaped
				recSnd.Refresh()
			}
		}
	})
	cntLbl := container.NewCenter(widget.NewLabel(StrCnt))
	rowRec := container.NewGridWithRows(3, recLbl, recSnd, cntLbl)

	cntLcl = widget.NewMultiLineEntry()

	sndBut = widget.NewButton(StrSnd, guiFyneSnd)
	sndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntLcl)
}

func guiFyneMakeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(StrRmt))
	rowTo := container.NewCenter(widget.NewLabel(StrTo))

	for i := 0; i < 2; i++ {
		sockRmt[i] = widget.NewSelectEntry(nil)
	}
	sockRmt[0].PlaceHolder = StrLcl
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

	rowFrm := container.NewCenter(widget.NewLabel(StrFrm))

	for i := 0; i < 2; i++ {
		sndMap[i] = make(map[string]struct{})
	}
	rcvMap = make(map[string]struct{})
	rowSockF = widget.NewSelect(nil, guiFyneSockF)
	rowSockF.Options = []string{StrAll}

	recLbl := container.NewCenter(widget.NewLabel(StrRec))
	recRcv = widget.NewSelect(nil, func(str string) {
		cntRmt.SetText(str)
		escaped := url.PathEscape(str)
		if len(escaped) > MaxRecLen {
			recRcv.Selected = escaped[0:MaxRecLen]
			recRcv.Refresh()
		} else {
			if escaped != str {
				recRcv.Selected = escaped
				recRcv.Refresh()
			}
		}
	})
	recMap = make(map[string][]string)
	cntLbl := container.NewCenter(widget.NewLabel(StrCnt))
	rowRec := container.NewGridWithRows(3, recLbl, recRcv, cntLbl)

	cntRmt = widget.NewMultiLineEntry()

	rowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	rowLog.Disable()

	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2,
		rowFrm, rowSockF, rowRec)
	return container.NewBorder(tops, rowLog, nil, nil, cntRmt)
}
