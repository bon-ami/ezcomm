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
	"gitlab.com/bon-ami/ezcomm"
)

const (
	MaxRecLen = 10
)

var (
	// snd2Slc is for UDP peer display
	snd2Slc [2][]string
	// sndMap is for UDP peer match
	sndMap [2]map[string]struct{}
	// recMap maps peer to incoming content
	recMap map[string][]string
	// recSlc records incoming content
	recSlc []string
	// rcvMap records whether anything ever received from a peer
	rcvMap map[string]struct{}
	// peeMap maps peer (addr) to channel of communication
	peeMap map[string]chan ezcomm.RoutCommStruc

	sockLcl, sockRmt                      [2]*widget.SelectEntry
	recRcv, recSnd, rowTcpSock2, rowSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	rowUdpSock2            *fyne.Container
	protRd                 *widget.RadioGroup
	lstBut, disBut, sndBut *widget.Button
	cntLcl, cntRmt         *widget.Entry
)

func connEnable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Enable()
	}
	protRd.Enable()
	lstBut.SetText(ezcomm.StringTran["StrLst"])
}

func connDisable() {
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Disable()
	}
	protRd.Disable()
	lstBut.SetText(ezcomm.StringTran["StrStp"])
}

func (GuiFyne) Log(inf ...any) {
	if fyneRowLog != nil {
		str := fmt.Sprintf("%s%v\n", time.Now().Format("01-02 15:04:05"), inf)
		fyneRowLog.SetText(fyneRowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
			str)
		fyneRowLog.CursorRow++
	}
	if ezcomm.LogWtTime {
		eztools.LogWtTime(inf)
	} else {
		eztools.Log(inf)
	}
}

func setLclSck(addr string) {
	ind := strings.LastIndex(addr, ":")
	sockLcl[0].SetText(addr[:ind])
	sockLcl[1].SetText(addr[ind+1:])
}

func getRmtSckStr() string {
	addr := sockRmt[0].Text
	if len(addr) < 1 {
		addr = ezcomm.DefAdr
	}
	return addr + ":" + sockRmt[1].Text
}

func getRmtSck() *net.UDPAddr {
	ret, err := net.ResolveUDPAddr(protRd.Selected, getRmtSckStr())
	if err != nil {
		f.Log(err)
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

func sockF(str string) {
	if str != ezcomm.StringTran["StrAll"] {
		recRcv.Options = recMap[str]
		return
	}
	recRcv.Options = recSlc
}

func butSnd(snd bool) {
	//GuiLog(true, "but snd", snd)
	if snd {
		sndBut.OnTapped = Snd
		sndBut.SetText(ezcomm.StringTran["StrSnd"])
	} else {
		sndBut.SetText(ezcomm.StringTran["StrCon"])
		sndBut.OnTapped = Connect
	}
	sndBut.Refresh()
}

func butSndByProt(prot string) {
	switch prot {
	case ezcomm.StrUdp:
		butSnd(true)
	case ezcomm.StrTcp:
		butSnd(false)
		//if len(sockRmt[1].Text) > 0 {
		/*} else {
			//sndBut.Text = STR_SND
			sndBut.OnTapped = guiFyneSnd
		}*/
	}
}

func sckRmt(single bool) {
	if single {
		rowUdpSock2.Hide()
		rowTcpSock2.Show()
	} else {
		rowUdpSock2.Show()
		rowTcpSock2.Hide()
	}
}

func Lstn() {
	var (
		err       error
		addrStruc net.Addr
	)
	addr := sockLcl[0].Text + ":" + sockLcl[1].Text
	connDisable()
	for i := range ezcomm.ChanComm {
		ezcomm.ChanComm[i] = make(chan ezcomm.RoutCommStruc, ezcomm.FlowComLen)
	}
	switch protRd.Selected {
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
		setLclSck(addrStruc.String())
		go ezcomm.ConnectedUdp(udpConn)
	case ezcomm.StrTcp:
		peeMap = make(map[string]chan ezcomm.RoutCommStruc)
		var lstnr net.Listener
		lstnr, err = ezcomm.ListenTcp(ezcomm.StrTcp, addr, ezcomm.ConnectedTcp, nil)
		if err != nil {
			break
		}
		addrStruc = lstnr.Addr()
		setLclSck(addrStruc.String())
		butSnd(true)
		sndBut.Disable()
		sckRmt(true)
		go ezcomm.ListeningTcp(lstnr)
	}
	if err != nil {
		for i := range ezcomm.ChanComm {
			ezcomm.ChanComm[i] = nil
		}
		connEnable()
		f.Log(ezcomm.StringTran["StrListeningOn"], err)
	} else {
		lstBut.OnTapped = Stop
		f.Log(ezcomm.StringTran["StrListeningOn"], addrStruc.String())
	}
}

// Connect is for TCP only
func Connect() {
	pr := getRmtSckStr()
	peeMap = make(map[string]chan ezcomm.RoutCommStruc)
	chn := make(chan ezcomm.RoutCommStruc, ezcomm.FlowComLen)
	peeMap[pr] = chn
	connDisable()
	/*conn*/ _, err := ezcomm.Client(protRd.Selected, pr, ezcomm.ConnectedTcp)
	if err != nil {
		connEnable()
		f.Log(ezcomm.StringTran["StrConnFail"]+pr, err)
		return
	}
	//lstBut.OnTapped = guiFyneStp
	sckRmt(true)
	//guiFyneConnected(conn.LocalAddr().String(), conn.RemoteAddr().String())
}

// Connected is for TCP only
func (f GuiFyne) Connected(lcl, rmt string, chn chan ezcomm.RoutCommStruc) {
	peeMap[rmt] = chn
	if ezcomm.ChanComm[0] == nil { //client
		lstBut.Hide()
		disBut.Show()
	}
	butSnd(true)
	disBut.Show()
	sndBut.Enable()
	setLclSck(lcl)
	rowTcpSock2.Options = append(rowTcpSock2.Options, rmt)
	if len(rowTcpSock2.Selected) < 1 {
		rowTcpSock2.SetSelectedIndex(0)
	}
	f.Log(lcl, "<->", rmt, ezcomm.StringTran["StrConnected"])
	rowTcpSock2.Refresh()
}

// Disconnected is for TCP only
func Disconnected(rmt string) {
	indx := -1
	ln := len(rowTcpSock2.Options)
	defer func() {
		delete(peeMap, rmt)
	}()
	for i, v := range rowTcpSock2.Options {
		if v == rmt {
			indx = i
			break
		}
	}
	switch {
	case indx == -1:
		f.Log(ezcomm.StringTran["StrDisconnected"], rmt, ezcomm.StringTran["StrNotInRec"])
		return
	case ln == 1:
		rowTcpSock2.Options = nil
		rowTcpSock2.Selected = ""
		rowTcpSock2.Refresh()
		disBut.Hide()
		if lstBut.Hidden { //client
			lstBut.Show()
			connEnable()
			//guiFyneSckRmt(false)
			butSnd(false)
			sndBut.Enable()
			f.Log(rmt, ezcomm.StringTran["StrDisconnected"])
		} else { //server
			sndBut.Disable()
			f.Log(rmt, ezcomm.StringTran["StrDisconnected"], ".", ezcomm.StringTran["StrSvrIdle"])
		}
		return
	}
	// reorder the records
	if indx != ln-1 {
		rowTcpSock2.Options[indx] = rowTcpSock2.Options[ln-1]
	}
	rowTcpSock2.Options = rowTcpSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		f.Log(ezcomm.StringTran["StrClntLft"], rowTcpSock2.Options)
	}
	rowTcpSock2.SetSelectedIndex(0)
	rowTcpSock2.Refresh()
}

// Disconn1 disconnect 1 peer TCP
func Disconn1() {
	rmtTcp := rowTcpSock2.Selected
	chn, ok := peeMap[rmtTcp]
	if !ok {
		f.Log(ezcomm.StringTran["StrNoPeer4"], rmtTcp)
		return
	}
	f.Log(ezcomm.StringTran["StrDisconnecting"], rmtTcp)
	chn <- ezcomm.RoutCommStruc{
		Act: ezcomm.FlowChnEnd,
	}
	Disconnected(rmtTcp)
}

func Stop() {
	ezcomm.ChanComm[0] <- ezcomm.RoutCommStruc{
		Act: ezcomm.FlowChnEnd,
	}
	if !disBut.Hidden { //clients still running
		lstBut.Hide()
	} else {
		sckRmt(false)
		connEnable()
		butSndByProt(protRd.Selected)
	}
	lstBut.OnTapped = Lstn
	f.Log(ezcomm.StringTran["StrStopLstn"])
}

func add2Rmt(indx int, txt string) {
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

func Snd() {
	var (
		rmtUdp *net.UDPAddr
		rmtTcp string
	)
	//GuiLog(true, "to send")
	if !protRd.Disabled() {
		switch protRd.Selected {
		case ezcomm.StrUdp: // listen before sending
			rmtUdp = getRmtSck()
			if rmtUdp == nil {
				return
			}
			for i := 0; i < 2; i++ {
				add2Rmt(i, sockRmt[i].Text)
			}
			Lstn()
			if !protRd.Disabled() { // failed
				return
			}
		case ezcomm.StrTcp:
			panic("NOT listening when sending")
		}
	}
	switch protRd.Selected {
	case ezcomm.StrUdp: // listen before sending
		//GuiLog(true, "to send UDP")
		rmtUdp = getRmtSck()
		if rmtUdp == nil {
			return
		}
		//GuiLog(true, "requesting to send UDP", ezcomm.ChanComm[0])
		ezcomm.ChanComm[0] <- ezcomm.RoutCommStruc{
			Act:     ezcomm.FlowChnSnd,
			Data:    cntLcl.Text,
			PeerUdp: rmtUdp,
		}
		//GuiLog(true, "requested to send UDP")
	case ezcomm.StrTcp:
		rmtTcp = rowTcpSock2.Selected
		chn, ok := peeMap[rmtTcp]
		if !ok {
			f.Log(ezcomm.StringTran["StrNoPeer4"], rmtTcp)
			break
		}
		chn <- ezcomm.RoutCommStruc{
			Act:  ezcomm.FlowChnSnd,
			Data: cntLcl.Text,
		}
	}
}

func (f GuiFyne) Rcv(comm ezcomm.RoutCommStruc) {
	//GuiLog(true, "recv", comm)
	switch comm.Act {
	case ezcomm.FlowChnRcv:
		if comm.Err != nil {
			f.Log(ezcomm.StringTran["StrFl2Rcv"], comm.Err)
			break
		}
		cntRmt.SetText(comm.Data)
		var peer string
		if comm.PeerUdp != nil {
			peer = comm.PeerUdp.String()
			ind := strings.LastIndex(peer, ":")
			add2Rmt(0, peer[:ind])
			add2Rmt(1, peer[ind+1:])
		} else if comm.PeerTcp != nil {
			peer = comm.PeerTcp.String()
		} else {
			f.Log(ezcomm.StringTran["StrGotFromSw"])
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
				f.Log("<-", peer, comm.Data)
			}
			f.Log(ezcomm.StringTran["StrGotFrom"], peer)
		}
		recRcv.Options = append(recRcv.Options, comm.Data)
		recSlc = append(recSlc, comm.Data)
	}
}

// Ended is run when TCP peer disconnected
func (f GuiFyne) Ended(comm ezcomm.RoutCommStruc) {
	switch comm.Act {
	case ezcomm.FlowChnEnd:
		if comm.PeerTcp != nil {
			peer := comm.PeerTcp.String()
			if ch, ok := peeMap[peer]; !ok {
				f.Log(ezcomm.StringTran["StrUnknownDsc"], peer)
			} else {
				ch <- comm
			}
			Disconnected(peer)
		}
	}
}

func (f GuiFyne) Snt(comm ezcomm.RoutCommStruc) {
	//GuiLog(true, "sent", comm)
	switch comm.Act {
	case ezcomm.FlowChnSnd:
		if comm.Err != nil {
			f.Log(ezcomm.StringTran["StrFl2Snd"], comm.Err)
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
			f.Log(">-", peer, comm.Data)
		}
		f.Log(ezcomm.StringTran["StrSnt2"], peer)
		recSnd.Options = append(recSnd.Options, comm.Data)
		if comm.PeerUdp != nil {
			addrStr := comm.PeerUdp.String()
			ind := strings.LastIndex(addrStr, ":")
			sockRmt[0].SetText(addrStr[:ind])
			sockRmt[1].SetText(addrStr[ind+1:])
		}
	}
}

func makeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLcl"]))

	addrLbl := widget.NewLabel(ezcomm.StringTran["StrAdr"])
	portLbl := widget.NewLabel(ezcomm.StringTran["StrPrt"])
	for i := 0; i < 2; i++ {
		sockLcl[i] = widget.NewSelectEntry(nil)
	}
	sockLcl[0].PlaceHolder = ezcomm.DefAdr
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	protRd = widget.NewRadioGroup([]string{ezcomm.StrUdp, ezcomm.StrTcp}, nil)
	protRd.Horizontal = true
	protRd.SetSelected("udp")
	lstBut = widget.NewButton(ezcomm.StringTran["StrLst"], Lstn)
	disBut = widget.NewButton(ezcomm.StringTran["StrDis"], Disconn1)
	disBut.Hide()
	rowProt := container.NewHBox(protRd, lstBut, disBut)
	protRd.OnChanged = func(str string) {
		butSndByProt(str)
	}

	recLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRec"]))
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
	cntLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, recLbl, recSnd, cntLbl)

	cntLcl = widget.NewMultiLineEntry()

	sndBut = widget.NewButton(ezcomm.StringTran["StrSnd"], Snd)
	sndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntLcl)
}

func makeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrTo"]))

	for i := 0; i < 2; i++ {
		sockRmt[i] = widget.NewSelectEntry(nil)
	}
	sockRmt[0].PlaceHolder = ezcomm.DefAdr
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

	rowFrm := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrFrm"]))

	for i := 0; i < 2; i++ {
		sndMap[i] = make(map[string]struct{})
	}
	rcvMap = make(map[string]struct{})
	rowSockF = widget.NewSelect(nil, sockF)
	rowSockF.Options = []string{ezcomm.StringTran["StrAll"]}

	recLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRec"]))
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
	cntLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, recLbl, recRcv, cntLbl)

	cntRmt = widget.NewMultiLineEntry()

	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2,
		rowFrm, rowSockF, rowRec)
	return container.NewBorder(tops, nil, nil, nil, cntRmt)
}
