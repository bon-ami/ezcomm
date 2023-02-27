package main

import (
	"errors"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	lafButSnd, lafButDel, lafButClr, lafButExp *widget.Button
	lafLst                                     *widget.RadioGroup
	dldStrSlc                                  []string
)

func clrDlds() {
	uri, err := encodeFileDown("")
	defer func() {
		showNG("", err)
	}()
	if err != nil {
		return
	}
	lst, err := storage.List(uri)
	if err != nil {
		return
	}
	for _, u1 := range lst {
		err = storage.Delete(u1)
		// only last result will be prompted
		if err != nil {
			Log(err)
		}
	}
	tabLAfShown()
}

func getDld() (u fyne.URI, err error) {
	dld, err := checkDldDir()
	if err != nil {
		return
	}
	fdld := filepath.Join(dld, lafLst.Selected)
	return encodeFilePath(fdld)
}

func showNG(msg string, err error) {
	if err != nil {
		if len(msg) > 0 {
			Log(msg, err)
		} else {
			Log(err)
		}
		dialog.ShowInformation(ezcomm.StringTran["StrNG"],
			msg+err.Error(), ezcWin)
	}
}

func delDld() {
	u, err := getDld()
	defer func() {
		showNG("", err)
	}()
	if err != nil {
		return
	}
	err = storage.Delete(u)
	tabLAfShown()
}

// expDld exports a file from Downloads
//
//	it must run in non-UI thread, otherwise tryWriteFile() will block
func expDld(sel string, uriDst fyne.ListableURI) {
	fold := filepath.Join(decodeFilePath(uriDst), sel)
	wr, fnew, abt := tryWriteFile(writerNew, fold)
	if wr != nil {
		defer wr.Close()
	}
	var (
		err    error
		errMsg string
	)
	defer func() {
		showNG(errMsg, err)
	}()
	if abt {
		if len(fnew) > 0 {
			err = errors.New(fnew)
		}
		return
	}
	if wr == nil {
		if len(fnew) < 1 {
			fnew = fold
		}
		fdst, err := encodeFilePath(fnew)
		if err != nil {
			errMsg = "fnew NOT encoded: "
			return
		}
		wr, err = storage.Writer(fdst)
		if err != nil {
			errMsg = "fnew NOT a writer: "
			return
		}
	}
	fsrc, err := getDld()
	if err != nil {
		errMsg = "download NOT got: "
		return
	}
	rd, err := storage.Reader(fsrc)
	if err != nil {
		errMsg = fsrc.String() + "NOT read: "
		return
	}
	err = cpFile(rd, wr)
	// storage.Copy() does not work for file:// on Android
	// repository.GenericCopy() does not work on Android if no permission granted in Settings
}

func makeTabLAf() *container.TabItem {
	//if dldStrSlc != nil {
	lafLst = widget.NewRadioGroup(dldStrSlc, func(string) {
		resetLafButs()
	})
	lafButSnd = widget.NewButton(ezcomm.StringTran["StrSnd"], func() {
		if err := filLclChk(nil, lafLst.Selected); err != nil {
			showNG("", err)
			return
		}
		tabs.Select(tabFil)
	})
	lafButExp = widget.NewButton(ezcomm.StringTran["StrExp"], func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				go expDld(lafLst.Selected, uri)
			}
		}, ezcWin)
	})
	lafButDel = widget.NewButton(ezcomm.StringTran["StrDel"], func() {
		dialog.ShowConfirm(ezcomm.StringTran["StrDel"],
			lafLst.Selected, func(c bool) {
				if !c {
					return
				}
				delDld()
			}, ezcWin)
	})
	lafButClr = widget.NewButton(ezcomm.StringTran["StrClr"], func() {
		dialog.ShowConfirm(ezcomm.StringTran["StrAlert"],
			ezcomm.StringTran["StrRmAll"], func(c bool) {
				if !c {
					return
				}
				clrDlds()
			}, ezcWin)
	})
	rowButs := container.NewHBox(lafButSnd, lafButExp, lafButDel, lafButClr)

	rowLaf := container.NewBorder(nil, rowButs, nil, nil, container.NewVScroll(lafLst))
	return container.NewTabItem(ezcomm.StringTran["StrDownloads"],
		rowLaf)
}

func tabLAfShown() {
	var dldURISlc []fyne.URI
	_, err := checkDldDir()
	if err != nil || dldURI == nil {
		Log("NO downloads dir!", dldURI, err)
		return
	}
	if ok, err := storage.CanList(dldURI); err == nil {
		if ok {
			dldURISlc, err = storage.List(dldURI)
			if err != nil {
				Log("downloads NOT listed!", dldURI, err)
				return
			}
		}
	} else {
		Log("downloads NOT listable!", dldURI, err)
		return
	}
	dldStrSlc = nil
	var prev bool
	for _, uri := range dldURISlc {
		u := uri.Name()
		dldStrSlc = append(dldStrSlc, u)
		if u == lafLst.Selected {
			prev = true
		}
	}
	if !prev {
		lafLst.SetSelected("")
	}
	lafLst.Options = dldStrSlc
	resetLafButs()
	lafLst.Refresh()
	lafButSnd.Refresh()
	lafButExp.Refresh()
	lafButDel.Refresh()
	lafButClr.Refresh()
}

func resetLafButs() {
	lafButSnd.Disable()
	lafButExp.Disable()
	lafButDel.Disable()
	lafButClr.Disable()
	if lafLst.Options != nil {
		if len(lafLst.Selected) > 0 {
			lafButExp.Enable()
			lafButDel.Enable()
			lafButSnd.Enable()
		}
		lafButClr.Enable()
	}
}
