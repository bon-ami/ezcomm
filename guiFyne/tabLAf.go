package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	lafButSnd *widget.Button
)

func makeTabLAf() *container.TabItem {
	var dldUriSlc []fyne.URI
	if dldUri != nil {
		if ok, err := storage.CanList(dldUri); err != nil {
			Log(err)
		} else {
			if ok {
				dldUriSlc, err = storage.List(dldUri)
				if err != nil {
					Log(err)
				}
			}
		}
	}
	var dldStrSlc []string
	for _, uri := range dldUriSlc {
		dldStrSlc = append(dldStrSlc, uri.Name())
	}
	var lafLst *widget.RadioGroup
	var rowLaf *fyne.Container
	//if dldStrSlc != nil {
	lafLst = widget.NewRadioGroup(dldStrSlc, func(sel string) {
		tabs.Select(tabFil)
	})
	lafButSnd = widget.NewButton(ezcomm.StringTran["StrSnd"], func() {
	})
	rowLaf = container.NewVBox(lafLst, lafButSnd)
	/*} else {
	lafLst = widget.NewRadioGroup
	rowLaf = container.New*/
	//}
	return container.NewTabItem(ezcomm.StringTran["StrDownloads"],
		rowLaf)
}

func tabLAfShown() {
}
