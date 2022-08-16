package main

import (
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	filLcl, filRmt *widget.Button
	filNmO, filNmI string
	tabFil         *container.TabItem
	// filPieces maps peer address to former file name to append to
	//filPieces      map[string]string
	refRcv, refSnd *widget.Select
	filEnable      bool
)

func filLclChk() {
	if len(filNmO) < 1 {
		return
	}
	cap := filepath.Base(filNmO)
	if inf, err := os.Stat(filNmO); err != nil {
		cap += "\n" + err.Error()
		filEnable = false

	} else {
		filEnable = true
		if inf.Size() >= ezcomm.FlowRcvLen {
			switch protRd.Selected {
			case ezcomm.StrUdp:
				cap += "\n>" + strconv.Itoa(ezcomm.FlowRcvLen) + "\n" + ezcomm.StringTran["StrTooLarge2Rcv"]
				filEnable = false
			case ezcomm.StrTcp:
			}
		}
	}
	filLcl.SetText(cap)
	if filEnable {
		chkNEnableSnd(true)
	} else {
		if sndBut.Text == ezcomm.StringTran["StrSnd"] {
			sndBut.Disable()
		}
		//filNmO = ""
	}
}

func filButLcl() {
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err == nil && uri != nil {
			filNmO = uri.URI().Path()
			filLclChk()
		}
	}, ezcWin)
}

func filButRmt() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err == nil && uri != nil {
			filNmI = uri.Path()
			filRmt.SetText(uri.Name())
		}
	}, ezcWin)
}

func isSndFile(wrapperFunc func(fn string, proc func([]byte) error) error,
	fun func(buf []byte) error) bool {
	if tabFil.Content.Visible() {
		if len(filNmO) > 0 {
			if err := wrapperFunc(filNmO, fun); err != nil {
				Log(ezcomm.StringTran["StrFl2Snd"], err)
			}
		}
		return true
	}
	return false
}

// SntFile checks whether sent data is of a file
// Return values:
//	"", false: not file
//	"", true: file, but in progress
//	file name, true: sending finished
func SntFile(comm ezcomm.RoutCommStruc) string {
	fn, _, _, end := ezcomm.BulkFile("", "", comm.Data)
	if end {
		return filepath.Base(fn)
	} else {
		return ""
	}
}

func SntFileOk(fn string) {
	if len(fn) > MaxRecLen {
		refSnd.Options = append(refSnd.Options, fn[0:MaxRecLen])
	} else {
		refSnd.Options = append(refSnd.Options, fn)
	}
}

func RcvFile(comm ezcomm.RoutCommStruc, addr string) (string, string) {
	/*var (
		appending bool
		fn        string
	)*/
	if len(filNmI) < 1 {
		return addr, ezcomm.StringTran["StrDirI"]
	}
	wr := eztools.FileAppend
	fn, offset, data, end := ezcomm.BulkFile(filNmI, addr, comm.Data)
	if data == nil {
		return "", ""
	}
	if offset == 0 {
		wr = eztools.FileWrite
	}
	/*if fn, appending = filPieces[addr]; !appending {
		wr = eztools.FileWrite
	} else {
		if eztools.Debugging && eztools.Verbose > 1 {
			Log("appending to", fn)
		}
	}*/
	//eztools.Log(wr, fnFull, data)
	if err := wr(fn, data); err != nil {
		Log(ezcomm.StringTran["StrFl2Rcv"], err)
	} else {
		/*if len(comm.Data) == ezcomm.FlowRcvLen {
		filPieces[addr] = fn
		appending = false // do not delete it*/
		/*Log("to append")
		} else {
			Log("not to append", len(comm.DataI))*/
		/*}*/
		if end {
			fn = filepath.Base(fn)
			if len(fn) > MaxRecLen {
				refRcv.Options = append(refRcv.Options, fn[0:MaxRecLen])
			} else {
				refRcv.Options = append(refRcv.Options, fn[0:])
			}
			return addr, fn
		}
	}
	/*if appending {
		delete(filPieces, addr)
	}*/
	return "", ""
}

func makeControlsLF() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLcl"]))

	addrLbl := widget.NewLabel(ezcomm.StringTran["StrAdr"])
	portLbl := widget.NewLabel(ezcomm.StringTran["StrPrt"])
	rowSock := container.NewGridWithRows(2,
		addrLbl, sockLcl[0], portLbl, sockLcl[1])

	rowProt := container.NewHBox(protRd, lstBut, disBut)

	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
	refSnd = widget.NewSelect(nil, nil)
	rowRec := container.NewGridWithRows(2, recLbl, refSnd)

	filLbl := widget.NewLabel(ezcomm.StringTran["StrFilO"])
	filLbl.Wrapping = fyne.TextWrapWord
	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec, filLbl)

	filLcl = widget.NewButton(ezcomm.StringTran["StrFil"], filButLcl)
	return container.NewBorder(tops, sndBut, nil, nil, filLcl)
}

func makeControlsRF() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrTo"]))

	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
	refRcv = widget.NewSelect(nil, nil)
	rowRec := container.NewGridWithRows(2, recLbl, refRcv)

	filLbl := widget.NewLabel(ezcomm.StringTran["StrDirI"])
	filLbl.Wrapping = fyne.TextWrapWord
	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2, rowRec, filLbl)
	filRmt = widget.NewButton(ezcomm.StringTran["StrDir"], filButRmt)
	return container.NewBorder(tops, nil, nil, nil, filRmt)
}

func makeTabFil() *container.TabItem {
	//filPieces = make(map[string]string)
	tabFil = container.NewTabItem(ezcomm.StringTran["StrFil"],
		container.NewGridWithColumns(2, makeControlsLF(), makeControlsRF()))
	return tabFil
}

func isFilEnable() bool {
	if tabFil.Content.Visible() {
		return filEnable
	}
	return true
}

func tabFilShown() {
	if !filEnable {
		if sndBut.Text == ezcomm.StringTran["StrSnd"] {
			sndBut.Disable()
		}
	} else {
		chkNEnableSnd(true)
	}
}
