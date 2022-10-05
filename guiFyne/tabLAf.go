package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	lafButSnd *widget.Button
)

func makeTabLAf() *container.TabItem {
	lafLst := widget.NewRadioGroup([]string{}, func(sel string) {
		tabs.Select(tabFil)
	})
	lafButSnd = widget.NewButton(ezcomm.StringTran["StrSnd"], func() {
	})
	return container.NewTabItem(ezcomm.StringTran["StrInfLan"],
		container.NewVBox(lafLst, lafButSnd))
}

func tabLAfShown() {
}
