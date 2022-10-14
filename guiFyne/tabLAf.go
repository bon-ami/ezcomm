package main

import (
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
	defer showNG(err)
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
	Log(dld)
	if err != nil {
		return
	}
	fdld := filepath.Join(dld, lafLst.Selected)
	return encodeFilePath(fdld)
}

func showNG(err error) {
	if err != nil {
		Log(err)
		dialog.ShowInformation(ezcomm.StringTran["StrNG"],
			err.Error(), ezcWin)
	}
}

func delDld() {
	u, err := getDld()
	defer showNG(err)
	if err != nil {
		return
	}
	err = storage.Delete(u)
	tabLAfShown()
}

func expDld(uriDst fyne.ListableURI) {
	fold := filepath.Join(decodeFilePath(uriDst), lafLst.Selected)
	Log(fold)
	fnew, abt := tryWriteFile(writerNew, fold)
	if abt {
		return
	}
	if len(fnew) < 1 {
		fnew = fold
	}
	fsrc, err := getDld()
	defer showNG(err)
	if err != nil {
		return
	}
	fdst, err := encodeFilePath(fnew)
	if err != nil {
		return
	}
	Log(fsrc, fdst)
	err = storage.Copy(fsrc, fdst)
	Log(err)
}

func makeTabLAf() *container.TabItem {
	//if dldStrSlc != nil {
	lafLst = widget.NewRadioGroup(dldStrSlc, func(string) {
		resetLafButs()
	})
	lafButSnd = widget.NewButton(ezcomm.StringTran["StrSnd"], func() {
		if err := filLclChk(nil, lafLst.Selected); err != nil {
			showNG(err)
			return
		}
		tabs.Select(tabFil)
	})
	lafButExp = widget.NewButton(ezcomm.StringTran["StrExp"], func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				expDld(uri)
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
	var dldUriSlc []fyne.URI
	_, err := checkDldDir()
	if err != nil || dldUri == nil {
		Log("NO downloads dir!", dldUri, err)
		return
	}
	if ok, err := storage.CanList(dldUri); err != nil {
		Log("downloads NOT listable!", err)
		return
	} else {
		if ok {
			dldUriSlc, err = storage.List(dldUri)
			if err != nil {
				Log("downloads NOT listed!", err)
				return
			}
		}
	}
	dldStrSlc = nil
	var prev bool
	for _, uri := range dldUriSlc {
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
		Log(lafLst.Selected)
		if len(lafLst.Selected) > 0 {
			lafButExp.Enable()
			lafButDel.Enable()
			lafButSnd.Enable()
		}
		lafButClr.Enable()
	}
}
