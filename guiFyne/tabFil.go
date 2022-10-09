package main

import (
	"io"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	filLcl *widget.Button
	filRmt *widget.Entry
	filNmI string
	// filPieces maps peer address to former file name to append to
	//filPieces      map[string]string
	refRcv, refSnd *widget.Select
	filEnable      bool
	rc             fyne.URIReadCloser
	dldUri         fyne.URI
)

func filLclChk() {
	if rc == nil {
		return
	}
	cap := rc.URI().Name()
	filEnable = true
	switch protRd.Selected {
	case ezcomm.StrUdp:
		r, err := storage.Reader(rc.URI())
		if err != nil {
			cap += "\n" + err.Error()
			filEnable = false
			break
		}
		cap, _, err = ezcomm.TryOnlyChunk(cap, r)
		if err == nil {
			break
		}
		if err == eztools.ErrOutOfBound ||
			err == eztools.ErrInvalidInput {
			ml, _ := ezcomm.Sz41stChunk(cap)
			cap += "\n>" + strconv.Itoa(ml) +
				"\n" + ezcomm.StringTran["StrTooLarge2Rcv"]
		} else {
			cap += "\n" + err.Error()
		}
		filEnable = false
	case ezcomm.StrTcp:
	}
	filLcl.SetText(cap)
	if filEnable {
		chkNEnableSnd(true)
	} else {
		if sndBut.Text == ezcomm.StringTran["StrSnd"] {
			sndBut.Disable()
		}
	}
}

func filButLcl() {
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err == nil && uri != nil {
			rc = uri
			uri.Close()
			filLclChk()
		}
	}, ezcWin)
}

/*func filButRmt() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err == nil && uri != nil {
			filNmI = uri.Path()
			filRmt.SetText(uri.Name())
		}
	}, ezcWin)
}*/

func isSndFile(wrapperFunc func(string, io.ReadCloser, func([]byte) error) error,
	fun func(buf []byte) error) bool {
	if tabFil.Content.Visible() {
		if rc != nil {
			r, err := storage.Reader(rc.URI())
			if err != nil {
				Log(ezcomm.StringTran["StrFl2Snd"], err)
				return true
			}
			if err = wrapperFunc(rc.URI().Name(), r, fun); err != nil {
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
	filRmt.SetText(fn + " ->")
}

// RcvFile saves incoming file piece
//   To avoid asking user for every file, when permission needed,
//   all files saved to app directory.
// Return values: fn=file name or error string
func RcvFile(comm ezcomm.RoutCommStruc, addr string) (peer, fn string) {
	if len(filNmI) < 1 {
		return addr, ""
	}
	wr := eztools.FileAppend
	fn, first, data, end := ezcomm.BulkFile(filNmI, addr, comm.Data)
	if data == nil {
		return "", ""
	}
	/*if _, ret := tryWriteFile(fn); ret {
		return "", eztools.ErrAbort.Error()
	}*/
	if first {
		wr = eztools.FileWrite
	}
	if err := wr(fn, data); err != nil {
		Log(ezcomm.StringTran["StrFl2Rcv"], err)
	} else {
		/*if eztools.Debugging && eztools.Verbose > 1 {
			Log("saved first piece",
				first, "to", fn)
		}*/
		if end {
			fn = filepath.Base(fn)
			filRmt.SetText("<- " + fn)
			if len(fn) > MaxRecLen {
				refRcv.Options = append(refRcv.Options, fn[0:MaxRecLen])
			} else {
				refRcv.Options = append(refRcv.Options, fn[0:])
			}
			return addr, fn
		}
	}
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
	filLAf := widget.NewButton(ezcomm.StringTran["StrRcvFil"], func() {
		tabs.Select(tabLAf)
	})
	bots := container.NewVBox(filLAf, sndBut)
	return container.NewBorder(tops, bots, nil, nil, filLcl)
}

func makeControlsRF() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrRmt"]))
	rowTo := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrTo"]))

	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
	refRcv = widget.NewSelect(nil, nil)
	rowRec := container.NewGridWithRows(2, recLbl, refRcv)

	/*filLbl := widget.NewLabel(ezcomm.StringTran["StrDirI"])
	filLbl.Wrapping = fyne.TextWrapWord*/
	tops := container.NewVBox(rowLbl, rowTo, rowUdpSock2, rowTcpSock2, rowRec /*, filLbl*/)
	const dldDir = "Downloads"
	filNmI = appStorage.RootURI().Path()
	dirDld := filepath.Join(filNmI, dldDir)
	Log(dirDld)
	dldUri = storage.NewFileURI(dirDld)
	if exi, err := storage.Exists(dldUri); err != nil {
		eztools.Log("NO", dirDld, "detectable!", err)
	} else {
		if exi {
			if cn, err := storage.CanList(dldUri); err != nil {
				eztools.Log("NO", dirDld, "listable!", err)
			} else {
				if !cn {
					Log(dirDld, "is a file!", filNmI,
						"will be used as download directory!")
				} else {
					filNmI = dirDld
				}
			}
		} else {
			if err = storage.CreateListable(dldUri); err != nil {
				eztools.Log("NO", dirDld, "created!", err)
			} else {
				filNmI = dirDld
			}
		}
	}
	//filRmt = widget.NewButton(ezcomm.StringTran["StrDir"], filButRmt)
	filRmt = widget.NewMultiLineEntry()
	filRmt.Wrapping = fyne.TextWrapWord
	filRmt.SetText(filNmI)
	return container.NewBorder(tops, nil, nil, nil, filRmt)
}

func makeTabFil() *container.TabItem {
	return container.NewTabItem(ezcomm.StringTran["StrFil"],
		container.NewGridWithColumns(2, makeControlsLF(), makeControlsRF()))
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
	protRd.Refresh()
	lstBut.Refresh()
}
