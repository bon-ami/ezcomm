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
	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	filLclBut/*, dirRmtBut*/ *widget.Button
	filRmt   *widget.Entry
	incomDir string
	// filPieces maps peer address to former file name to append to
	//filPieces      map[string]string
	refRcv, refSnd    *widget.Select
	filEnable         bool
	outgoFile, dldURI fyne.URI
)

// filLclChk checks whether file is too large for UDP
// Parameters: must not be both non-nil
func filLclChk(rc fyne.URIReadCloser, fil string) (err error) {
	if rc != nil {
		outgoFile = rc.URI()
	}
	if outgoFile == nil && len(fil) < 1 {
		return eztools.ErrInvalidInput
	}
	if len(fil) > 0 {
		uri, err := encodeFileDown(fil)
		if err != nil {
			return err
		}
		if uri == nil {
			return eztools.ErrOutOfBound
		}
		outgoFile = storage.NewFileURI(uri.Path())
	}
	cap := outgoFile.Name()
	filEnable = true
	switch protRd.Selected {
	case ezcomm.StrUDP:
		if rc == nil {
			if len(fil) < 1 {
				filEnable = false
				break
			} else {
				rc, err = storage.Reader(outgoFile)
				if err != nil {
					filEnable = false
					break
				}
			}
		}
		cap, _, err = ezcomm.TryOnlyChunk(cap, rc)
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
	case ezcomm.StrTCP:
	}
	filLclBut.SetText(cap)
	if filEnable {
		chkNEnableSnd(true)
	} else {
		if sndBut.Text == ezcomm.StringTran["StrSnd"] {
			sndBut.Disable()
		}
	}
	return err
}

func filButLcl() {
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err == nil && uri != nil {
			filLclChk(uri, "")
		}
	}, ezcWin)
}

/*func dirButRmt() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dirRmtBut.SetText(err.Error())
			return
		}
		if uri == nil {
			dirRmtBut.SetText(ezcomm.StringTran["StrDir"])
			return
		}
		incomDir = translateFilePath(decodeFilePath(uri))
		dirRmtBut.SetText(incomDir)
	}, ezcWin)
}*/

// isSndFile
// Parameters: wrapperFunc must Close ReadCloser
func isSndFile(wrapperFunc func(string, io.ReadCloser,
	func([]byte) error) error,
	fun func(buf []byte) error) bool {
	if tabFil.Content.Visible() {
		if outgoFile != nil {
			r, err := storage.Reader(outgoFile)
			if err != nil {
				Log(ezcomm.StringTran["StrFl2Snd"], err)
				return true
			}
			if err = wrapperFunc(outgoFile.Name(), r, fun); err != nil {
				Log(ezcomm.StringTran["StrFl2Snd"], err)
			}
		}
		return true
	}
	return false
}

// SntFile checks whether sent data is of a file
// Return values: file name, transfer ended
func SntFile(comm ezcomm.RoutCommStruc) (string, bool) {
	fn, _, _, end := ezcomm.BulkFile("", "", comm.Data)
	return filepath.Base(fn), end
}

// SntFileOk successfully sent of a file
func SntFileOk(fn string, fin bool) {
	if !fin {
		filRmt.SetText(fn + " ...>")
		return
	}
	if len(fn) > MaxRecLen {
		refSnd.Options = append(refSnd.Options, fn[0:MaxRecLen])
	} else {
		refSnd.Options = append(refSnd.Options, fn)
	}
	filRmt.SetText(fn + " ->")
}

// RcvFile saves incoming file piece
//
//	To avoid asking user for every file, when permission needed,
//	all files saved to app directory.
//
// Return values: fn=file name or error string
func RcvFile(comm ezcomm.RoutCommStruc, addr string) (peer, fn string) {
	if len(incomDir) < 1 {
		return addr, ""
	}
	fn, first, data, end := ezcomm.BulkFile(incomDir, addr, comm.Data)
	if data == nil {
		return "", ""
	}
	/*if _, ret := tryWriteFile(ezcomm.FlowWriterNew, fn); ret {
		return "", eztools.ErrAbort.Error()
	}*/
	wr := eztools.FileAppend
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
		fn = filepath.Base(fn)
		if !end {
			filRmt.SetText("<... " + fn)
		} else {
			filRmt.SetText("<- " + fn)
			if len(fn) > MaxRecLen {
				refRcv.Options = append(refRcv.Options,
					fn[0:MaxRecLen])
			} else {
				refRcv.Options = append(refRcv.Options, fn[0:])
			}
			return addr, fn
		}
	}
	return "", ""
}

func makeControlsLF() *fyne.Container {
	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
	refSnd = widget.NewSelect(nil, nil)
	rowRec := container.NewGridWithRows(2, recLbl, refSnd)

	filLbl := widget.NewLabel(ezcomm.StringTran["StrFilO"])
	filLbl.Wrapping = fyne.TextWrapWord
	tops := container.NewVBox(makeControlsLclSocks(), rowRec, filLbl)

	filLclBut = widget.NewButton(ezcomm.StringTran["StrFil"], filButLcl)
	filLAfL = widget.NewButton(ezcomm.StringTran["StrRcvFil"], func() {
		tabs.Select(tabLAf)
	})
	bots := container.NewVBox(filLAfL, sndBut)
	return container.NewBorder(tops, bots, nil, nil,
		container.NewHScroll(filLclBut))
}

var filLAfL, filLAfR *widget.Button

func makeControlsRF() *fyne.Container {
	recLbl := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrRec"]))
	refRcv = widget.NewSelect(nil, nil)
	rowRec := container.NewGridWithRows(2, recLbl, refRcv)

	/*filLbl := widget.NewLabel(ezcomm.StringTran["StrDirI"])
	filLbl.Wrapping = fyne.TextWrapWord*/
	//dirRmtBut = widget.NewButton(ezcomm.StringTran["StrDir"], dirButRmt)
	tops := container.NewVBox(makeControlsRmtSocks(), rowRec /*, dirRmt*/)
	var err error
	incomDir, err = checkDldDir()
	if err != nil {
		Log(dldDirNm, "not created!", err)
		dialog.ShowInformation(ezcomm.StringTran["StrLang"],
			ezcomm.StringTran["StrFnt4LangBuiltin"], ezcWin)
	}
	if len(incomDir) > 0 {
		dirPath := widget.NewMultiLineEntry()
		dirPath.Wrapping = fyne.TextWrapWord
		dirPath.SetText(incomDir)
		tops.Add(dirPath)
	}
	filRmt = widget.NewMultiLineEntry()
	filRmt.Wrapping = fyne.TextWrapWord
	filRmt.Disable()
	filLAfR = widget.NewButton(ezcomm.StringTran["StrRcvFil"], func() {
		tabs.Select(tabLAf)
	})
	return container.NewBorder(tops, filLAfR, nil, nil, filRmt)
}

func makeTabFil() *container.TabItem {
	return container.NewTabItem(ezcomm.StringTran["StrFil"],
		container.NewGridWithColumns(2, makeControlsLF(),
			makeControlsRF()))
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
	filLAfL.Refresh()
	filLAfR.Refresh()
	filLclBut.Refresh()
}
