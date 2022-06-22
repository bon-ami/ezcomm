package guiFyne

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var (
	fyneRowLog *Entry // widget.Entry
)

func guiFyneMakeControlsInfLog() *fyne.Container {
	fyneRowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	fyneRowLog.Disable()
	/*eztools.SetLogFunc(func(p ...any) {
		GuiLog(false, p)
	})*/
	return container.NewBorder(nil, nil, nil, nil, fyneRowLog)
}
