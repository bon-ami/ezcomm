package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	fyneRowLog *Entry // widget.Entry
)

func makeTabLog() *container.TabItem {
	logTxt := widget.NewEntry()
	if len(ezcomm.CfgStruc.LogFile) > 0 {
		logTxt.SetText(ezcomm.CfgStruc.LogFile)
	} else {
		logTxt.SetText(ezcomm.StringTran["StrLog"])
	}
	logTxt.Disable()
	logBut := widget.NewButton(ezcomm.StringTran["StrLog"], func() {
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			/*fyneCfgLogTxt.SetText("")
			defer fyneCfgLogTxt.Refresh()*/
			if err != nil {
				Log("open log file", err)
				return
			}
			if uri == nil {
				return
			}
			fn := uri.URI().Path()
			if eztools.InitLogger(uri) == nil {
				ezcomm.CfgStruc.LogFile = fn
				writeCfg()
				logTxt.SetText(fn)
				logTxt.Refresh()
			}
		}, ezcWin)
	})
	rowVerbose := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrVbs"]))
	verboseSel := widget.NewSelect(nil, func(lvl string) {
		newLvl := verboseFrmStr(lvl)
		if newLvl == ezcomm.CfgStruc.Verbose {
			return
		}
		eztools.Verbose = newLvl
		ezcomm.CfgStruc.Verbose = newLvl
		writeCfg()
	})
	verboseSel.Options = []string{
		ezcomm.StringTran["StrHgh"], ezcomm.StringTran["StrMdm"], ezcomm.StringTran["StrLow"], ezcomm.StringTran["StrNon"],
	}
	verboseSel.SetSelected(verbose2Str())
	fyneRowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	fyneRowLog.Wrapping = fyne.TextWrapWord
	fyneRowLog.Disable()
	top := container.NewVBox(logTxt, logBut, rowVerbose, verboseSel)
	return container.NewTabItem(ezcomm.StringTran["StrInfLog"],
		container.NewBorder(top, nil, nil, nil, fyneRowLog))
}

func tabLogShown() {

}
