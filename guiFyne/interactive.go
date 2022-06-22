package guiFyne

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
	"gitlab.com/bon-ami/ezcomm"
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
)

func guiFyneEnable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		fyneSockLcl[i].Enable()
	}
	fyneProt.Enable()
	fyneLstBut.SetText(ezcomm.StringTran["StrLst"])
}

func guiFyneDisable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		fyneSockLcl[i].Disable()
	}
	fyneProt.Disable()
	fyneLstBut.SetText(ezcomm.StringTran["StrStp"])
}

func GuiLog(log2File bool, inf ...any) {
	str := fmt.Sprintf("%s\n%v\n", time.Now().Format("01-02 15:04:05"), inf)
	fyneRowLog.SetText(fyneRowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
		str)
	fyneRowLog.CursorRow++
	if log2File {
		if ezcomm.LogWtTime {
			eztools.LogWtTime(inf)
		} else {
			eztools.Log(inf)
		}
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
		GuiLog(false, err)
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
	if str != ezcomm.StringTran["StrAll"] {
		fyneRecRcv.Options = ezcomm.RecMap[str]
		return
	}
	fyneRecRcv.Options = ezcomm.RecSlc
}

func guiFyneButSnd(snd bool) {
	if snd {
		fyneSndBut.OnTapped = guiFyneSnd
		fyneSndBut.SetText(ezcomm.StringTran["StrSnd"])
	} else {
		fyneSndBut.SetText(ezcomm.StringTran["StrCon"])
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
	for i := range ezcomm.ChanComm {
		ezcomm.ChanComm[i] = make(chan ezcomm.RoutCommStruc, ezcomm.FlowComLen)
	}
	switch fyneProt.Selected {
	case ezcomm.StrUdp:
		var udpConn *net.UDPConn
		udpConn, err = ezcomm.ListenUdp(ezcomm.StrUdp, addr)
		if err != nil {
			break
		}
		if udpConn == nil {
			panic("NO connection got")
		}
		addrStruc = udpConn.LocalAddr()
		guiFyneSetLclSck(addrStruc.String())
		go ezcomm.ConnectedUdp(udpConn)
	case ezcomm.StrTcp:
		ezcomm.PeeMap = make(map[string]chan ezcomm.RoutCommStruc)
		var lstnr net.Listener
		lstnr, err = ezcomm.ListenTcp(ezcomm.StrTcp, addr, ezcomm.ConnectedTcp, nil)
		if err != nil {
			break
		}
		addrStruc = lstnr.Addr()
		guiFyneSetLclSck(addrStruc.String())
		guiFyneButSnd(true)
		fyneSndBut.Disable()
		guiFyneSckRmt(true)
		go ezcomm.ListeningTcp(lstnr)
	}
	if err != nil {
		for i := range ezcomm.ChanComm {
			ezcomm.ChanComm[i] = nil
		}
		guiFyneEnable()
		GuiLog(true, err)
	} else {
		fyneLstBut.OnTapped = guiFyneStp
		GuiLog(true, ezcomm.StringTran["StrListeningOn"], addrStruc.String())
	}
}

// guiFyneCon is for TCP only
func guiFyneCon() {
	pr := guiFyneGetRmtSckStr()
	ezcomm.PeeMap = make(map[string]chan ezcomm.RoutCommStruc)
	chn := make(chan ezcomm.RoutCommStruc, ezcomm.FlowComLen)
	ezcomm.PeeMap[pr] = chn
	guiFyneDisable()
	/*conn*/ _, err := ezcomm.Client(fyneProt.Selected, pr, ezcomm.ConnectedTcp)
	if err != nil {
		guiFyneEnable()
		GuiLog(true, ezcomm.StringTran["StrConnFail"]+pr, err)
		return
	}
	//lstBut.OnTapped = guiFyneStp
	guiFyneSckRmt(true)
	//guiFyneConnected(conn.LocalAddr().String(), conn.RemoteAddr().String())
}

// GuiConnected is for TCP only
func GuiConnected(lcl, rmt string) {
	fyneSndBut.OnTapped = guiFyneSnd
	if ezcomm.ChanComm[0] == nil { //client
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
	GuiLog(true, lcl, "<->", rmt, ezcomm.StringTran["StrConnected"])
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
		GuiLog(true, ezcomm.StringTran["StrDisconnected"], rmt, ezcomm.StringTran["StrNotInRec"])
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
			GuiLog(true, rmt, ezcomm.StringTran["StrDisconnected"])
		} else { //server
			fyneSndBut.Disable()
			GuiLog(true, rmt, ezcomm.StringTran["StrDisconnected"], ".", ezcomm.StringTran["StrSvrIdle"])
		}
		return
	}
	// reorder the records
	if indx != ln-1 {
		fyneRowTcpSock2.Options[indx] = fyneRowTcpSock2.Options[ln-1]
	}
	fyneRowTcpSock2.Options = fyneRowTcpSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		GuiLog(true, ezcomm.StringTran["StrClntLft"], fyneRowTcpSock2.Options)
	}
	fyneRowTcpSock2.SetSelectedIndex(0)
	fyneRowTcpSock2.Refresh()
}

// guiFyneDis disconnect 1 peer TCP
func guiFyneDis() {
	rmtTcp := fyneRowTcpSock2.Selected
	chn, ok := ezcomm.PeeMap[rmtTcp]
	if !ok {
		GuiLog(true, ezcomm.StringTran["StrNoPeer3"], rmtTcp)
		return
	}
	GuiLog(true, ezcomm.StringTran["StrDisconnecting"], rmtTcp)
	chn <- ezcomm.RoutCommStruc{
		Act: ezcomm.FlowChnEnd,
	}
	guiFyneDisconnected(rmtTcp)
}

func guiFyneStp() {
	ezcomm.ChanComm[0] <- ezcomm.RoutCommStruc{
		Act: ezcomm.FlowChnEnd,
	}
	if !fyneDisBut.Hidden { //clients still running
		fyneLstBut.Hide()
	} else {
		guiFyneSckRmt(false)
		guiFyneEnable()
		guiFyneButSnd(false)
	}
	fyneLstBut.OnTapped = guiFyneLst
	GuiLog(true, ezcomm.StringTran["StrStopLstn"])
}

func guiFyneAdd2Rmt(indx int, txt string) {
	if len(txt) < 1 {
		return
	}
	/*if _, ok := sndMap[indx][txt]; ok {
		return
	}
	sndMap[indx][txt] = struct{}{}*/
	if _, ok := ezcomm.SndMap[indx][txt]; ok {
		return
	}
	ezcomm.SndMap[indx][txt] = struct{}{}
	ezcomm.Snd2Slc[indx] = append(ezcomm.Snd2Slc[indx], txt)
	fyneSockRmt[indx].SetOptions(ezcomm.Snd2Slc[indx])
}

func guiFyneSnd() {
	var (
		rmtUdp *net.UDPAddr
		rmtTcp string
	)
	if !fyneProt.Disabled() {
		switch fyneProt.Selected {
		case ezcomm.StrUdp: // listen before sending
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
		case ezcomm.StrTcp:
			panic("NOT listening when sending")
		}
	}
	switch fyneProt.Selected {
	case ezcomm.StrUdp: // listen before sending
		rmtUdp = guiFyneGetRmtSck()
		if rmtUdp == nil {
			return
		}
		ezcomm.ChanComm[0] <- ezcomm.RoutCommStruc{
			Act:     ezcomm.FlowChnSnd,
			Data:    fyneCntLcl.Text,
			PeerUdp: rmtUdp,
		}
	case ezcomm.StrTcp:
		rmtTcp = fyneRowTcpSock2.Selected
		chn, ok := ezcomm.PeeMap[rmtTcp]
		if !ok {
			GuiLog(true, ezcomm.StringTran["StrNoPeer3"], rmtTcp)
			break
		}
		chn <- ezcomm.RoutCommStruc{
			Act:  ezcomm.FlowChnSnd,
			Data: fyneCntLcl.Text,
		}
	}
}

func GuiRcv(comm ezcomm.RoutCommStruc) {
	switch comm.Act {
	case ezcomm.FlowChnRcv:
		if comm.Err != nil {
			GuiLog(true, ezcomm.StringTran["StrFl1Rcv"], comm.Err)
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
			GuiLog(true, ezcomm.StringTran["StrGotFromSw"])
		}
		if len(peer) > 0 {
			if _, ok := ezcomm.RcvMap[peer]; !ok {
				ezcomm.RcvMap[peer] = struct{}{}
				fyneRowSockF.Options = append(fyneRowSockF.Options, peer)
			}
			fyneRowSockF.SetSelected(peer)
			fyneRowSockF.Refresh()
			ezcomm.RecMap[peer] = append(ezcomm.RecMap[peer], comm.Data)
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.LogWtTime("<-", peer, comm.Data)
			}
			GuiLog(true, ezcomm.StringTran["StrGotFrom"], peer)
		}
		fyneRecRcv.Options = append(fyneRecRcv.Options, comm.Data)
		ezcomm.RecSlc = append(ezcomm.RecSlc, comm.Data)
	}
}

// GuiEnded is run when TCP peer disconnected
func GuiEnded(comm ezcomm.RoutCommStruc) {
	switch comm.Act {
	case ezcomm.FlowChnEnd:
		if comm.PeerTcp != nil {
			peer := comm.PeerTcp.String()
			if ch, ok := ezcomm.PeeMap[peer]; !ok {
				GuiLog(true, ezcomm.StringTran["StrUnknownDsc"], peer)
			} else {
				ch <- comm
			}
			guiFyneDisconnected(peer)
		}
	}
}

func GuiSnt(comm ezcomm.RoutCommStruc) {
	switch comm.Act {
	case ezcomm.FlowChnSnd:
		if comm.Err != nil {
			GuiLog(true, ezcomm.StringTran["StrFl1Snd"], comm.Err)
			break
		}
		var peer string
		switch {
		case comm.PeerUdp != nil:
			peer = comm.PeerUdp.String()
		case comm.PeerTcp != nil:
			peer = comm.PeerTcp.String()
		default:
			peer = ezcomm.StringTran["StrSw"]
		}
		if eztools.Debugging && eztools.Verbose > 2 {
			eztools.LogWtTime(">-", peer, comm.Data)
		}
		GuiLog(true, ezcomm.StringTran["StrSnt1"], peer)
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
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLcl"]))

	addrLbl := widget.NewLabel(ezcomm.StringTran["StrAdr"])
	portLbl := widget.NewLabel(ezcomm.StringTran["StrPrt"])
	for i := 0; i < 2; i++ {
		fyneSockLcl[i] = widget.NewSelectEntry(nil)
	}
	fyneSockLcl[0].PlaceHolder = ezcomm.DefAdr
	rowSock := container.NewGridWithRows(2,
		addrLbl, fyneSockLcl[0], portLbl, fyneSockLcl[1])

	fyneProt = widget.NewRadioGroup([]string{ezcomm.StrUdp, ezcomm.StrTcp}, nil)
	fyneProt.Horizontal = true
	fyneProt.SetSelected("udp")
	fyneLstBut = widget.NewButton(ezcomm.StringTran["StrLst"], guiFyneLst)
	fyneDisBut = widget.NewButton(ezcomm.StringTran["StrDis"], guiFyneDis)
	fyneDisBut.Hide()
	rowProt := container.NewHBox(fyneProt, fyneLstBut, fyneDisBut)
	fyneProt.OnChanged = func(str string) {
		switch str {
		case ezcomm.StrUdp:
			guiFyneButSnd(true)
		case ezcomm.StrTcp:
			guiFyneButSnd(false)
			//if len(sockRmt[1].Text) > 0 {
			/*} else {
				//sndBut.Text = STR_SND
				sndBut.OnTapped = guiFyneSnd
			}*/
		}
		fyneSndBut.Refresh()
	}

	recLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRec"]))
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
	cntLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, recLbl, fyneRecSnd, cntLbl)

	fyneCntLcl = widget.NewMultiLineEntry()

	fyneSndBut = widget.NewButton(ezcomm.StringTran["StrSnd"], guiFyneSnd)
	fyneSndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, fyneSndBut, nil, nil, fyneCntLcl)
}

func guiFyneMakeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrTo"]))

	for i := 0; i < 2; i++ {
		fyneSockRmt[i] = widget.NewSelectEntry(nil)
	}
	fyneSockRmt[0].PlaceHolder = ezcomm.StringTran["StrLcl"]
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

	rowFrm := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrFrm"]))

	for i := 0; i < 2; i++ {
		ezcomm.SndMap[i] = make(map[string]struct{})
	}
	ezcomm.RcvMap = make(map[string]struct{})
	fyneRowSockF = widget.NewSelect(nil, guiFyneSockF)
	fyneRowSockF.Options = []string{ezcomm.StringTran["StrAll"]}

	recLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRec"]))
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
	ezcomm.RecMap = make(map[string][]string)
	cntLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, recLbl, fyneRecRcv, cntLbl)

	fyneCntRmt = widget.NewMultiLineEntry()

	tops := container.NewVBox(rowLbl, rowTo, fyneRowUdpSock2, fyneRowTcpSock2,
		rowFrm, fyneRowSockF, rowRec)
	return container.NewBorder(tops, nil, nil, nil, fyneCntRmt)
}
