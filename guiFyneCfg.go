package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
)

var (
	fyneCfgLogTxt                 *widget.Entry
	fyneCfgLogBut, fyneCfgVerbose *widget.Button
	fyneCfgLangSel                *widget.Select
)

func guiFyneMakeControlsCfg(ezcWin fyne.Window) *fyne.Container {
	flowFnStt := widget.NewEntry()
	flowFnStt.Disable()
	flowFnTxt := widget.NewEntry()
	flowFnTxt.SetText(StrFlw)
	flowFnTxt.Disable()
	var flowFlBut *widget.Button
	flowFlBut = widget.NewButton(StrFlw, func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			flowFnStt.SetText("")
			defer flowFnStt.Refresh()
			if err != nil {
				eztools.LogWtTime("open flow file", err)
				flowFnTxt.SetText(StrFlw)
				flowFnStt.SetText("flow file opened ERR")
			}
			if uri == nil {
				flowFnStt.SetText("flow file NOT opened")
				return
			}
			flowFlBut.Disable()
			flowFnStt.SetText("flow file running...")
			flowFnStt.Refresh()
			fn := uri.URI().String()
			flowFnTxt.SetText(fn)
			flowFnTxt.Refresh()
			resChn := make(chan bool)
			if !runFlowReaderBG(uri, resChn) {
				eztools.LogWtTime("flow file NOT run", fn)
				flowFnStt.SetText("flow NOT run")
				flowFnStt.Refresh()
				flowFlBut.Enable()
				return
			}
			go func() {
				var resStr string
				if <-resChn {
					resStr = "OK"
				} else {
					resStr = "NG"
				}
				flowFnStt.SetText("flow file finished as " + resStr)
				flowFnStt.Refresh()
				flowFlBut.Enable()
			}()
		}, ezcWin)
	})
	return container.NewVBox(flowFnTxt, flowFlBut, flowFnStt)
}
