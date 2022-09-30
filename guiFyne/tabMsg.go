package main

import (
	"fmt"
	"io"
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

	sockLcl, sockRmt                      [2]*widget.SelectEntry
	recRcv, recSnd, rowTcpSock2, rowSockF *widget.Select
	//selectCtrls    []*widget.SelectEntry
	rowUdpSock2            *fyne.Container
	protRd                 *widget.RadioGroup
	lstBut, disBut, sndBut *widget.Button
	cntLcl, cntRmt         *widget.Entry
	tabMsg                 *container.TabItem
	sndEnable              bool
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

func Log(inf ...any) {
	if fyneRowLog != nil {
		str := fmt.Sprintf("%s%v\n",
			time.Now().Format("01-02 15:04:05"), inf)
		fyneRowLog.SetText(fyneRowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
			str)
		fyneRowLog.CursorRow++
	}
	eztools.Log(inf)
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
		Log(err)
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
	chkNEnableSnd(tabFil.Content.Visible())
	/*if !filEnable {
		sndBut.Disable()
	} else {
		sndBut.Enable()
	}*/
}

// sckRmt sets remote as single/combined sockets for TCP or UDP
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
	//Log("Listen clicked")
	var (
		err       error
		addrStruc net.Addr
	)
	addr := sockLcl[0].Text + ":" + sockLcl[1].Text
	connDisable()
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
		for i := range chn {
			chn[i] = make(chan ezcomm.RoutCommStruc,
				ezcomm.FlowComLen)
		}
		go ezcomm.ConnectedUdp(Log, chn, udpConn)
		clntRoutine()
	case ezcomm.StrTcp:
		err = svrTcp.Listen("", addr)
		//var lstnr net.Listener
		/*lstnr, err = ezcomm.ListenTcp(Log, Connected, ezcomm.StrTcp, addr,
		ezcomm.ConnectedTcp, nil)*/
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
	/*conn*/ _, err := ezcomm.Client(Log, TcpClnConnected,
		protRd.Selected, pr, ezcomm.ConnectedTcp)
	if err != nil {
		connEnable()
		Log(ezcomm.StringTran["StrConnFail"]+pr, err)
		return
	}
	//lstBut.OnTapped = Stop
	sckRmt(true)
	//guiFyneConnected(conn.LocalAddr().String(), conn.RemoteAddr().String())
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

// TcpClnConnected is TCP client routine
func TcpClnConnected(addr [4]string, chnC [2]chan ezcomm.RoutCommStruc) {
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
	TcpSvrConnected(addr)
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

func TcpSvrConnected(addr [4]string) {
	if len(addr[1]) < 1 {
		svrConnected(addr)
		return
	}
	//butSnd(true)
	disBut.Show()
	if isFilEnable() {
		sndBut.Enable()
	}
	sndEnable = true
	//setLclSck(addr[0])
	rowTcpSock2.Options = append(rowTcpSock2.Options, addr[1])
	if len(rowTcpSock2.Selected) < 1 {
		rowTcpSock2.SetSelectedIndex(0)
	}
	Log(addr, ezcomm.StringTran["StrConnected"])
	rowTcpSock2.Refresh()
}

// Disconnected is for TCP only
func Disconnected(rmt string) {
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
		Log(ezcomm.StringTran["StrDisconnected"],
			rmt, ezcomm.StringTran["StrNotInRec"])
		return
	case ln == 1:
		rowTcpSock2.Options = nil
		rowTcpSock2.Selected = ""
		rowTcpSock2.Refresh()
		disBut.Enable()
		disBut.Hide()
		/*if lstBut.Hidden { //client
			lstBut.Show()
			connEnable()
			//guiFyneSckRmt(false)
			butSnd(false)
			sndBut.Enable()
			Log(rmt, ezcomm.StringTran["StrDisconnected"])
		} else { //server*/
		sndBut.Disable()
		sndEnable = false
		//if protRd.Selected == ezcomm.StrTcp {
		if chn[0] == nil {
			go chkSvrStopped(true)
			Log(rmt, ezcomm.StringTran["StrDisconnected"],
				".", ezcomm.StringTran["StrSvrIdle"])
		} else {
			svrStopped()
		}
		//} /*else {
		//svrStopped()}*/
		//}
		return
	}
	// reorder the records
	if indx != ln-1 {
		rowTcpSock2.Options[indx] = rowTcpSock2.Options[ln-1]
	}
	rowTcpSock2.Options = rowTcpSock2.Options[:ln-1]
	if eztools.Debugging && eztools.Verbose > 2 {
		Log(ezcomm.StringTran["StrClntLft"], rowTcpSock2.Options)
	}
	rowTcpSock2.SetSelectedIndex(0)
	rowTcpSock2.Refresh()
}

// Disconn1 disconnect 1 peer TCP
func Disconn1() {
	rmtTcp := rowTcpSock2.Selected
	if chn[0] == nil {
		svrTcp.Disconnect(rmtTcp)
	} else {
		disBut.Disable()
		chn[0] <- ezcomm.RoutCommStruc{
			Act: ezcomm.FlowChnEnd,
		}
	}
	//Disconnected(rmtTcp)
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
	rowTcpSock2.Options = nil
	for i := range chn {
		chn[i] = nil
	}
}

func chkSvrStopped(clients bool) {
	if eztools.Debugging && eztools.Verbose > 1 {
		Log("entering server stop check routine")
		defer Log("exiting server stop check routine")
	}
	svrTcp.Wait(clients)
	if lstBut.Disabled() {
		if svrTcp.HasStopped() {
			svrStopped()
		}
	}
}

// Stop stops current server
func Stp() {
	lstBut.OnTapped = Lstn
	Log(ezcomm.StringTran["StrStopLstn"])
	if chn[0] != nil {
		chn[0] <- ezcomm.RoutCommStruc{
			Act: ezcomm.FlowChnEnd,
		}
	}
	if protRd.Selected == ezcomm.StrTcp {
		lstBut.Disable()
		//if chn[0] == nil {
		svrTcp.Stop()
		go chkSvrStopped(false)
		//}
		return
	}
	disBut.Disable()
	/*if !disBut.Hidden { //clients still running
		lstBut.Hide()
	} else {
		svrStopped()
	}*/
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
	//Log("send clicked")
	var rmtUdp *net.UDPAddr
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
	var (
		sndFunc     func([]byte) error
		wrapperFunc func(fn string, rdr io.ReadCloser, proc func([]byte) error) error
	)
	switch protRd.Selected {
	case ezcomm.StrUdp: // listen before sending
		//Log("to send UDP")
		rmtUdp = getRmtSck()
		if rmtUdp == nil {
			return
		}
		sndFunc = func(data []byte) error {
			chn[0] <- ezcomm.RoutCommStruc{
				Act:     ezcomm.FlowChnSnd,
				Data:    data,
				PeerUdp: rmtUdp,
			}
			return nil
		}
		wrapperFunc = ezcomm.SndFile
	case ezcomm.StrTcp:
		if chn[0] == nil { //server
			sndFunc = func(buf []byte) error {
				svrTcp.Send(rowTcpSock2.Selected, buf)
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

func Rcv(comm ezcomm.RoutCommStruc) {
	//Log("recv", comm)
	if comm.Err != nil {
		Log(ezcomm.StringTran["StrFl2Rcv"], comm.Err)
		return
	}
	var peer, addr string
	getAddr := func() int {
		ind := strings.LastIndex(peer, ":")
		addr = peer[:ind]
		return ind
	}
	if comm.PeerUdp != nil {
		peer = comm.PeerUdp.String()
		ind := getAddr()
		add2Rmt(0, addr)
		add2Rmt(1, peer[ind+1:])
	} else if comm.PeerTcp != nil {
		peer = comm.PeerTcp.String()
		getAddr()
	}
	if len(peer) > 0 {
		Log(ezcomm.StringTran["StrGotFrom"], peer)
	} else {
		Log(ezcomm.StringTran["StrGotFromSw"])
	}

	var data string
	if ezcomm.IsDataFile(comm.Data) {
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
	if comm.PeerTcp != nil {
		peer := comm.PeerTcp.String()
		if chn[0] == nil {
			svrTcp.Disconnect(peer) // maybe duplicate
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

func Snt(comm ezcomm.RoutCommStruc) {
	var (
		data   string
		isFile bool
	)
	if ezcomm.IsDataFile(comm.Data) {
		data = SntFile(comm)
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
		case comm.PeerUdp != nil:
			peer = comm.PeerUdp.String()
		case comm.PeerTcp != nil:
			peer = comm.PeerTcp.String()
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
			SntFileOk(data)
		}
	}
}

func SntMsg(comm ezcomm.RoutCommStruc) {
	recSnd.Options = append(recSnd.Options, string(comm.Data))
	if comm.PeerUdp != nil {
		addrStr := comm.PeerUdp.String()
		ind := strings.LastIndex(addrStr, ":")
		sockRmt[0].SetText(addrStr[:ind])
		sockRmt[1].SetText(addrStr[ind+1:])
	}
}

func makeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrLcl"]))

	addrLbl := widget.NewLabel(ezcomm.StringTran["StrAdr"])
	portLbl := widget.NewLabel(ezcomm.StringTran["StrPrt"])
	for i := 0; i < 2; i++ {
		sockLcl[i] = widget.NewSelectEntry(nil)
	}
	sockLcl[0].PlaceHolder = ""
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	protRd = widget.NewRadioGroup(
		[]string{ezcomm.StrUdp, ezcomm.StrTcp}, nil)
	protRd.Horizontal = true
	protRd.SetSelected("udp")
	lstBut = widget.NewButton(ezcomm.StringTran["StrLst"], Lstn)
	disBut = widget.NewButton(ezcomm.StringTran["StrDis"], Disconn1)
	disBut.Hide()
	rowProt := container.NewHBox(protRd, lstBut, disBut)
	protRd.OnChanged = func(str string) {
		if len(str) < 1 {
			protRd.SetSelected("udp")
		}
		filLclChk()
		butSndByProt(str)
	}

	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
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
	cntLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrCnt"]))
	rowRec := container.NewGridWithRows(3, recLbl, recSnd, cntLbl)

	cntLcl = widget.NewMultiLineEntry()

	sndBut = widget.NewButton(ezcomm.StringTran["StrSnd"], Snd)
	sndBut.Disable()

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntLcl)
}

func makeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrTo"]))

	for i := 0; i < 2; i++ {
		sockRmt[i] = widget.NewSelectEntry(nil)
	}
	sockRmt[0].PlaceHolder = ezcomm.DefAdr
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
	rowUdpSock2 = container.NewGridWithColumns(2, sockRmt[0], sockRmt[1])
	rowTcpSock2 = widget.NewSelect(nil, func(str string) {
	})
	rowTcpSock2.Hide()

	rowFrm := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrFrm"]))

	for i := 0; i < 2; i++ {
		sndMap[i] = make(map[string]struct{})
	}
	rcvMap = make(map[string]struct{})
	rowSockF = widget.NewSelect(nil, sockF)
	rowSockF.Options = []string{ezcomm.StringTran["StrAll"]}

	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
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
	rowRec := container.NewGridWithRows(3, recLbl, recRcv, cntLbl)

	cntRmt = widget.NewMultiLineEntry()

	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2,
		rowFrm, rowSockF, rowRec)
	return container.NewBorder(tops, nil, nil, nil, cntRmt)
}

func makeTabMsg() *container.TabItem {
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
	protRd.Refresh()
	lstBut.Refresh()
	chkNEnableSnd(false)
}
