package main

import (
	"fyne.io/fyne/v2/container"
)

func makeTabWeb() *container.TabItem {
	return container.NewTabItem("HTTP", //ezcomm.StringTran["StrHTTP"],
		container.NewGridWithColumns(2, makeControlsLF(), makeControlsRF()))
}

func tabWebShown() {
	protRd.Refresh()
	lstBut.Refresh()
}
