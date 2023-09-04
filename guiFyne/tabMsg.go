package main

import (
	"io"
	"net"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

const (
	// MaxRecLen max length of a text to show in records
	MaxRecLen = 10
)

var (
	// svrTCP is for TCP server only
	svrTCP ezcomm.SvrTCP
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

	sockLcl, sockRmt                      [2]*widget.SelectEntry
	recRcv, recSnd, rowTCPSock2, rowSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	rowUDPSock2            *fyne.Container
	protRd                 *widget.RadioGroup
	lstBut, disBut, sndBut *widget.Button
	cntLcl, cntRmt         *widget.Entry
	tabMsg                 *container.TabItem
	sndEnable, tabMsgInit  bool
)

func connEnable() {
	tabs.DisableItem(tabLan)
	tabs.DisableItem(tabWeb)
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Enable()
	}
	protRd.Enable()
	lstBut.SetText(ezcomm.StringTran["StrLst"])
}

func connDisable() {
	tabs.EnableItem(tabLan)
	tabs.EnableItem(tabWeb)
	//for _, i := range selectCtrls {
	for i := 0; i < 2; i++ {
		sockLcl[i].Disable()
	}
	protRd.Disable()
	lstBut.SetText(ezcomm.StringTran["StrStp"])
}

func setLclAddrs(addr []string) {
	sockLcl[0].SetOptions(addr)
}

func setLclSck(addr string) {
	ho, po := parseSck(addr)
	sockLcl[0].SetText(ho)
	sockLcl[1].SetText(po)
}

func getRmtSckStr() string {
	addr := sockRmt[0].Text
	if len(addr) < 1 {
		addr = ezcomm.DefPeerAdr
	}
	return addr + ":" + sockRmt[1].Text
}

func getRmtSck() *net.UDPAddr {
	ret, err := net.ResolveUDPAddr(protRd.Selected, getRmtSckStr())
	if err != nil {
		Log(err)
		return nil
	}
	return ret
}

func sockF(str string) {
	if str != ezcomm.StringTran["StrAll"] {
		recRcv.Options = recMap[str]
		return
	}
	recRcv.Options = recSlc
}

// butSnd set the send/conn button as Send
func butSnd(snd bool) {
	//Log("but snd", snd)
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
	case ezcomm.StrUDP:
		butSnd(true)
	case ezcomm.StrTCP:
		butSnd(false)
	}
	chkNEnableSnd(tabFil.Content.Visible())
}

// sckRmt sets remote as single/combined sockets for TCP or UDP
func sckRmt(single bool) {
	if single {
		rowUDPSock2.Hide()
		rowTCPSock2.Show()
	} else {
		rowUDPSock2.Show()
		rowTCPSock2.Hide()
	}
}

// Lstn listens to local socket
func Lstn() {
	//Log("Listen clicked")
	var (
		err       error
		addrStruc net.Addr
	)
	addr := sockLcl[0].Text + ":" + sockLcl[1].Text
	connDisable()
	switch protRd.Selected {
	case ezcomm.StrUDP:
		var udpConn *net.UDPConn
		udpConn, err = ezcomm.ListenUDP(ezcomm.StrUDP, addr)
		if err != nil {
			break
		}
		if udpConn == nil {
			panic("NO connection got")
		}
		addrStruc = udpConn.LocalAddr()
		setLclSck(addrStruc.String())
		for i := range chn {
			chn[i] = make(chan ezcomm.RoutCommStruc,
				ezcomm.FlowComLen)
		}
		go ezcomm.ConnectedUDP(Log, chn, udpConn)
		clntRoutine()
	case ezcomm.StrTCP:
		err = svrTCP.Listen("", addr)
		if err != nil {
			//Log(ezcomm.StringTran["StrListeningOn"], err)
			break
		}
	}
	if err != nil {
		connEnable()
		Log(ezcomm.StringTran["StrListeningOn"], err)
	} else {
		lstBut.OnTapped = Stp
		if addrStruc != nil {
			Log(ezcomm.StringTran["StrListeningOn"],
				addrStruc.String())
		}
	}
}

// Connect is for TCP only
func Connect() {
	//Log("Connect clicked")
	pr := getRmtSckStr()
	connDisable()
	_, err := ezcomm.Client(Log, TCPClnConnected,
		protRd.Selected, pr, ezcomm.Connected1Peer)
	if err != nil {
		connEnable()
		Log(ezcomm.StringTran["StrConnFail"]+pr, err)
		return
	}
	sckRmt(true)
}

func tcpConnAct(comm ezcomm.RoutCommStruc) {
	//Log("act from connection", comm)
	switch comm.Act {
	case ezcomm.FlowChnRcv:
		if len(comm.Data) < 1 {
			break
		}
		Rcv(comm)
	case ezcomm.FlowChnEnd:
		Ended(comm)
	case ezcomm.FlowChnSnd:
		Snt(comm)
		/*case ezcomm.FlowChnSndFil:
		SntFil(comm)*/
	}
}

// TCPClnConnected is TCP client routine
func TCPClnConnected(addr [4]string, chnC [2]chan ezcomm.RoutCommStruc) {
	if eztools.Debugging && eztools.Verbose > 1 {
		Log("entering TCP client routine", addr)
		defer func() {
			Log("exiting TCP client routine", addr)
		}()
	}
	for _, ch := range chnC {
		if ch == nil {
			connEnable()
			Log(ezcomm.StringTran["StrConnFail"],
				addr[3], "flooding")
			return
		}
	}
	chn = chnC
	svrConnected(addr)
	TCPSvrConnected(addr)
	lstBut.Hide()
	clntRoutine()
}

// clntRoutine is for TCP client and UDP
func clntRoutine() {
	go func(chn chan ezcomm.RoutCommStruc) {
		if eztools.Debugging && eztools.Verbose > 1 {
			Log("entering client routine")
			defer Log("exiting client routine")
		}
		for {
			comm := <-chn
			tcpConnAct(comm)
			if comm.Act == ezcomm.FlowChnEnd {
				break
			}
		}
	}(chn[1])
}

func svrConnected(addr [4]string) {
	Log(ezcomm.StringTran["StrListeningOn"], addr[0])
	//Log(addr)
	setLclSck(addr[0])
	butSnd(true)
	sndBut.Disable()
	sndEnable = false
	sckRmt(true)
}

// TCPSvrConnected on connected for TCP server
func TCPSvrConnected(addr [4]string) {
	if len(addr[1]) < 1 {
		svrConnected(addr)
		return
	}
	disBut.Show()
	if isFilEnable() {
		sndBut.Enable()
	}
	sndEnable = true
	rowTCPSock2.Options = append(rowTCPSock2.Options, addr[1])
	if len(rowTCPSock2.Selected) < 1 {
		rowTCPSock2.SetSelectedIndex(0)
	}
	Log(addr, ezcomm.StringTran["StrConnected"])
	rowTCPSock2.Refresh()
}

// Disconnected is for TCP only
func Disconnected(rmt string) {
	indx := -1
	ln := len(rowTCPSock2.Options)
	for i, v := range rowTCPSock2.Options {
		if v == rmt {
			indx = i
			break
		}
	}
	switch {
	case indx == -1:
		Log(ezcomm.StringTran["StrDisconnected"],
			rmt, ezcomm.StringTran["StrNotInRec"])
		return
	case ln == 1:
		rowTCPSock2.Options = nil
		rowTCPSock2.Selected = ""
		rowTCPSock2.Refresh()
		disBut.Enable()
		disBut.Hide()
		sndBut.Disable()
		sndEnable = false
		if chn[0] == nil {
			go chkSvrStopped(true)
			Log(rmt, ezcomm.StringTran["StrDisconnected"],
				".", ezcomm.StringTran["StrSvrIdle"])
		} else {
			svrStopped()
		}
		return
	}
	// reorder the records
	if indx != ln-1 {
		rowTCPSock2.Options[indx] = rowTCPSock2.Options[ln-1]
	}
	rowTCPSock2.Options = rowTCPSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		Log(ezcomm.StringTran["StrClntLft"], rowTCPSock2.Options)
	}
	rowTCPSock2.SetSelectedIndex(0)
	rowTCPSock2.Refresh()
}

// Disconn1 disconnect 1 peer TCP
func Disconn1() {
	rmtTCP := rowTCPSock2.Selected
	if chn[0] == nil {
		svrTCP.Disconnect(rmtTCP)
	} else {
		disBut.Disable()
		chn[0] <- ezcomm.RoutCommStruc{
			Act: ezcomm.FlowChnEnd,
		}
	}
}

func svrStopped() {
	sckRmt(false)
	connEnable()
	if len(sockRmt[1].Text) > 0 {
		if isFilEnable() {
			sndBut.Enable()
		}
		sndEnable = true
	}
	butSndByProt(protRd.Selected)
	lstBut.Enable()
	lstBut.Show()
	disBut.Enable()
	disBut.Hide()
	rowTCPSock2.Options = nil
	for i := range chn {
		chn[i] = nil
	}
}

func chkSvrStopped(clients bool) {
	if eztools.Debugging && eztools.Verbose > 1 {
		Log("entering server stop check routine")
		defer Log("exiting server stop check routine")
	}
	svrTCP.Wait(clients)
	if lstBut.Disabled() {
		if svrTCP.HasStopped() {
			svrStopped()
		}
	}
}

// Stp stops current server
func Stp() {
	lstBut.OnTapped = Lstn
	Log(ezcomm.StringTran["StrStopLstn"])
	if chn[0] != nil {
		chn[0] <- ezcomm.RoutCommStruc{
			Act: ezcomm.FlowChnEnd,
		}
	}
	if protRd.Selected == ezcomm.StrTCP {
		lstBut.Disable()
		svrTCP.Stop()
		go chkSvrStopped(false)
		return
	}
	disBut.Disable()
}

func add2Rmt(indx int, txt string) {
	if len(txt) < 1 {
		return
	}
	if _, ok := sndMap[indx][txt]; ok {
		return
	}
	sndMap[indx][txt] = struct{}{}
	snd2Slc[indx] = append(snd2Slc[indx], txt)
	sockRmt[indx].SetOptions(snd2Slc[indx])
}

// Snd sends sth.
func Snd() {
	//Log("send clicked")
	var rmtUDP *net.UDPAddr
	if !protRd.Disabled() {
		switch protRd.Selected {
		case ezcomm.StrUDP: // listen before sending
			rmtUDP = getRmtSck()
			if rmtUDP == nil {
				return
			}
			for i := 0; i < 2; i++ {
				add2Rmt(i, sockRmt[i].Text)
			}
			Lstn()
			if !protRd.Disabled() { // failed
				return
			}
		case ezcomm.StrTCP:
			panic("NOT listening when sending")
		}
	}
	var (
		sndFunc     func([]byte) error
		wrapperFunc func(fn string, rdr io.ReadCloser, proc func([]byte) error) error
	)
	switch protRd.Selected {
	case ezcomm.StrUDP: // listen before sending
		//Log("to send UDP")
		rmtUDP = getRmtSck()
		if rmtUDP == nil {
			return
		}
		sndFunc = func(data []byte) error {
			chn[0] <- ezcomm.RoutCommStruc{
				Act:     ezcomm.FlowChnSnd,
				Data:    data,
				PeerUDP: rmtUDP,
			}
			return nil
		}
		wrapperFunc = ezcomm.SndFile
	case ezcomm.StrTCP:
		if chn[0] == nil { //server
			sndFunc = func(buf []byte) error {
				svrTCP.Send(rowTCPSock2.Selected, buf)
				return nil
			}
		} else { //client
			sndFunc = func(buf []byte) error {
				chn[0] <- ezcomm.RoutCommStruc{
					Act:  ezcomm.FlowChnSnd,
					Data: buf,
				}
				return nil
			}
		}
		wrapperFunc = ezcomm.SplitFile
	}
	if !isSndFile(wrapperFunc, sndFunc) {
		sndFunc([]byte(cntLcl.Text))
	}
}

// Rcv receives sth.
func Rcv(comm ezcomm.RoutCommStruc) {
	//Log("recv", comm)
	if comm.Err != nil {
		Log(ezcomm.StringTran["StrFl2Rcv"], comm.Err)
		return
	}
	var peer, addr string
	if comm.PeerUDP != nil {
		peer = comm.PeerUDP.String()
		var po string
		addr, po = parseSck(peer)
		add2Rmt(0, addr)
		add2Rmt(1, po)
	} else if comm.PeerTCP != nil {
		peer = comm.PeerTCP.String()
		addr, _ = parseSck(peer)
	}
	if len(peer) > 0 {
		Log(ezcomm.StringTran["StrGotFrom"], peer)
	} else {
		Log(ezcomm.StringTran["StrGotFromSw"])
	}

	var data string
	if ok, _ := ezcomm.IsDataFile(comm.Data); ok {
		peer, data = RcvFile(comm, addr)
	} else {
		peer, data = RcvMsg(comm, peer)
	}
	if eztools.Debugging && eztools.Verbose > 2 {
		if len(peer) > 0 {
			Log("<-", peer, data)
		}
	}
}

// RcvMsg receives a text message
func RcvMsg(comm ezcomm.RoutCommStruc, peer string) (string, string) {
	data := string(comm.Data)
	cntRmt.SetText(data)
	recRcv.Options = append(recRcv.Options, data)
	recSlc = append(recSlc, data)
	if len(peer) > 0 {
		recMap[peer] = append(
			recMap[peer], data)
		if _, ok := rcvMap[peer]; !ok {
			rcvMap[peer] = struct{}{}
			rowSockF.Options = append(
				rowSockF.Options, peer)
		}
		rowSockF.SetSelected(peer)
		rowSockF.Refresh()
	}
	return peer, data
}

// Ended is run when peer disconnected
func Ended(comm ezcomm.RoutCommStruc) {
	if comm.PeerTCP != nil {
		peer := comm.PeerTCP.String()
		if chn[0] == nil {
			svrTCP.Disconnect(peer) // maybe duplicate
		} else {
			chn[0] <- ezcomm.RoutCommStruc{
				Act: ezcomm.FlowChnEnd,
			}
		}
		Disconnected(peer)
	} else {
		svrStopped()
	}
}

// Snt sth. sent
func Snt(comm ezcomm.RoutCommStruc) {
	var (
		data            string
		isFile, endFile bool
	)
	if ok, _ := ezcomm.IsDataFile(comm.Data); ok {
		data, endFile = SntFile(comm)
		isFile = true
	} else {
		data = string(comm.Data)
	}
	if comm.Err != nil {
		Log(ezcomm.StringTran["StrFl2Snd"], comm.Err)
		return
	}
	if len(data) > 0 {
		var peer string
		switch {
		case comm.PeerUDP != nil:
			peer = comm.PeerUDP.String()
		case comm.PeerTCP != nil:
			peer = comm.PeerTCP.String()
		default:
			peer = ezcomm.StringTran["StrSw"]
		}
		if eztools.Debugging && eztools.Verbose > 2 {
			Log(">-", peer, data)
		}
		Log(ezcomm.StringTran["StrSnt2"], peer)
		if !isFile {
			SntMsg(comm)
		} else {
			SntFileOk(data, endFile)
		}
	}
}

// SntMsg a text message was sent
func SntMsg(comm ezcomm.RoutCommStruc) {
	recSnd.Options = append(recSnd.Options, string(comm.Data))
	if comm.PeerUDP != nil {
		addrStr := comm.PeerUDP.String()
		ho, po := parseSck(addrStr)
		sockRmt[0].SetText(ho)
		sockRmt[1].SetText(po)
	}
}

func makeControlsLclSocks(prot bool) *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrLcl"]))

	addrLbl := widget.NewLabel(ezcomm.StringTran["StrAdr"])
	portLbl := widget.NewLabel(ezcomm.StringTran["StrPrt"])
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	if !prot {
		return container.NewVBox(rowLbl, rowSock)
	}
	rowProt := container.NewHBox(protRd, lstBut, disBut)
	return container.NewVBox(rowLbl, rowSock, rowProt)
}

func makeControlsSocks() {
	// local/left part
	for i := 0; i < 2; i++ {
		sockLcl[i] = widget.NewSelectEntry(nil)
	}
	sockLcl[0].PlaceHolder = ""

	protRd = widget.NewRadioGroup(
		[]string{ezcomm.StrUDP, ezcomm.StrTCP}, nil)
	protRd.Horizontal = true
	protRd.SetSelected("udp")
	protRd.OnChanged = func(str string) {
		if len(str) < 1 {
			protRd.SetSelected("udp")
		}
		filLclChk(nil, "")
		butSndByProt(str)
	}
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

	// remote/right part
	for i := 0; i < 2; i++ {
		sockRmt[i] = widget.NewSelectEntry(nil)
	}
	sockRmt[0].PlaceHolder = ezcomm.DefPeerAdr
	sockRmt[1].OnChanged = func(str string) {
		if len(str) > 0 { //&& len(sockRmt[0].Text) > -1 {
			if isFilEnable() {
				sndBut.Enable()
			}
			sndEnable = true
		} else {
			sndBut.Disable()
			sndEnable = false
		}
		sndBut.Refresh()
	}
}

func makeControlRecLbl() *fyne.Container {
	return container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
}

func makeControlsLcl() *fyne.Container {
	lstBut = widget.NewButton(ezcomm.StringTran["StrLst"], Lstn)
	disBut = widget.NewButton(ezcomm.StringTran["StrDis"], Disconn1)
	disBut.Hide()

	cntLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, makeControlRecLbl(), recSnd, cntLbl)

	cntLcl = widget.NewMultiLineEntry()

	sndBut = widget.NewButton(ezcomm.StringTran["StrSnd"], Snd)
	sndBut.Disable()

	tops := container.NewVBox(makeControlsLclSocks(true), rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntLcl)
}

func makeControlsRmtSocks() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrTo"]))

	return container.NewVBox(rowLbl, rowTo, rowUDPSock2, rowTCPSock2)
}

func makeControlsRmt() *fyne.Container {
	rowUDPSock2 = container.NewGridWithColumns(2, sockRmt[0], sockRmt[1])
	rowTCPSock2 = widget.NewSelect(nil, func(str string) {
	})
	rowTCPSock2.Hide()

	rowFrm := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrFrm"]))

	for i := 0; i < 2; i++ {
		sndMap[i] = make(map[string]struct{})
	}
	rcvMap = make(map[string]struct{})
	rowSockF = widget.NewSelect(nil, sockF)
	rowSockF.Options = []string{ezcomm.StringTran["StrAll"]}

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
	cntLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, makeControlRecLbl(), recRcv, cntLbl)

	cntRmt = widget.NewMultiLineEntry()

	tops := container.NewVBox(makeControlsRmtSocks(),
		rowFrm, rowSockF, rowRec)
	return container.NewBorder(tops, nil, nil, nil, cntRmt)
}

func makeTabMsg() *container.TabItem {
	svrTCP.ActFunc = tcpConnAct
	svrTCP.ConnFunc = TCPSvrConnected
	svrTCP.LogFunc = Log

	tabMsg = container.NewTabItem(ezcomm.StringTran["StrInt"],
		container.NewGridWithColumns(2,
			makeControlsLcl(), makeControlsRmt()))
	return tabMsg
}

func chkNEnableSnd(filShown bool) {
	filEnabled := true
	if filShown {
		filEnabled = filEnable
	}
	if filEnabled && sndEnable {
		sndBut.Enable()
	} else {
		sndBut.Disable()
	}
}

func tabMsgShown() {
	if tabMsgInit {
		return
	}
	tabMsgInit = true
	tabWebShown()
	sndBut.Refresh()
	chkNEnableSnd(false)
}
