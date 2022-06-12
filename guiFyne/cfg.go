package guiFyne

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

func guiFyneMakeControlsCfg(ezcWin fyne.Window) *fyne.Container {
	// flow part begins
	flowFnStt := widget.NewEntry()
	flowFnStt.PlaceHolder = ezcomm.StrStb
	flowFnStt.Disable()
	flowFnTxt := widget.NewEntry()
	flowFnTxt.SetText(ezcomm.StrFlw)
	flowFnTxt.Disable()
	var flowFlBut *widget.Button
	flowFlBut = widget.NewButton(ezcomm.StrFlw, func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			flowFnStt.SetText("")
			defer flowFnStt.Refresh()
			if err != nil {
				eztools.LogWtTime("open flow file", err)
				flowFnTxt.SetText(ezcomm.StrFlw)
				flowFnStt.SetText(ezcomm.StrFlowOpenErr)
			}
			if uri == nil {
				flowFnStt.SetText(ezcomm.StrFlowOpenNot)
				return
			}
			flowFlBut.Disable()
			flowFnStt.SetText(ezcomm.StrFlowRunning)
			flowFnStt.Refresh()
			fn := uri.URI().String()
			flowFnTxt.SetText(fn)
			flowFnTxt.Refresh()
			resChn := make(chan bool)
			if !ezcomm.RunFlowReaderBG(uri, resChn) {
				eztools.LogWtTime("flow file NOT run", fn)
				flowFnStt.SetText(ezcomm.StrFlowRunNot)
				flowFnStt.Refresh()
				flowFlBut.Enable()
				return
			}
			go func() {
				var resStr string
				if <-resChn {
					resStr = ezcomm.StrOK
				} else {
					resStr = ezcomm.StrNG
				}
				flowFnStt.SetText(ezcomm.StrFlowFinAs + resStr)
				flowFnStt.Refresh()
				flowFlBut.Enable()
			}()
		}, ezcWin)
	})
	// flow part ends

	langMap := make(map[string]string)
	logTxt := widget.NewEntry()
	if len(ezcomm.CfgStruc.LogFile) > 0 {
		logTxt.SetText(ezcomm.CfgStruc.LogFile)
	} else {
		logTxt.SetText(ezcomm.StrLog)
	}
	logTxt.Disable()
	logBut := widget.NewButton(ezcomm.StrLog, func() {
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			/*fyneCfgLogTxt.SetText("")
			defer fyneCfgLogTxt.Refresh()*/
			if err != nil {
				eztools.LogWtTime("open log file", err)
				return
			}
			if uri == nil {
				return
			}
			fn := uri.URI().Path()
			if eztools.InitLogger(uri) == nil {
				ezcomm.CfgStruc.LogFile = fn
				ezcomm.WriteCfg()
				logTxt.SetText(fn)
				logTxt.Refresh()
			}
		}, ezcWin)
	})
	rowVerbose := container.NewCenter(widget.NewLabel(ezcomm.StrVbs))
	verboseSel := widget.NewSelect(nil, func(lvl string) {
		newLvl := verboseFrmStr(lvl)
		if newLvl == ezcomm.CfgStruc.Verbose {
			return
		}
		eztools.Verbose = newLvl
		ezcomm.CfgStruc.Verbose = newLvl
		ezcomm.WriteCfg()
	})
	verboseSel.Options = []string{
		ezcomm.StrHgh, ezcomm.StrMdm, ezcomm.StrLow, ezcomm.StrNon,
	}
	verboseSel.SetSelected(verbose2Str())

	fontSel := widget.NewSelect(nil, nil)
	fontSel.PlaceHolder = ezcomm.StrFnt
	fontSel.Options = ezcomm.ListSystemFonts()
	currIndx := ezcomm.MatchSystemFontsFromPath(ezcomm.CfgStruc.GetFont())
	if currIndx >= 0 {
		fontSel.SetSelectedIndex(currIndx)
	}

	langSel := widget.NewSelect(nil, func(str string) {
		//eztools.LogWtTime("selecting", str)
		//loadStr(langMap[str])
		if ezcomm.CfgStruc.Language == langMap[str] {
			return
		}
		ezcomm.CfgStruc.Language = langMap[str]
		ezcomm.WriteCfg()

		ezcomm.I18nLoad(ezcomm.CfgStruc.Language)
		ezcomm.MatchFontFromLanguage()
		thm.SetFont(ezcomm.CfgStruc.GetFont())
		indx := ezcomm.MatchSystemFontsFromPath(ezcomm.CfgStruc.GetFont())
		if indx != eztools.InvalidID {
			fontSel.SetSelectedIndex(indx)
			//eztools.Log("font", indx)
		} else {
			fontSel.ClearSelected()
			//eztools.Log("font cleared")
		}
		dialog.ShowInformation(ezcomm.StrLang, ezcomm.StrReboot4Change, ezcWin)
	})
	langSel.PlaceHolder = ezcomm.StrLang
	eztools.ListLanguages(func(name, id string) {
		full := id + "_" + name
		langMap[full] = id
		langSel.Options = append(langSel.Options, full)
		if ezcomm.CfgStruc.Language == id {
			langSel.SetSelected(full)
		}
	})
	indx := ezcomm.MatchSystemFontsFromPath(ezcomm.CfgStruc.GetFont())
	if indx != eztools.InvalidID {
		fontSel.SetSelectedIndex(indx)
	}

	fontBut := widget.NewButton(ezcomm.StrFnt4Lang, func() {
		lang := langMap[langSel.Selected]
		if len(lang) < 1 {
			return
		}
		font := ezcomm.MatchSystemFontsFromIndex(fontSel.SelectedIndex())
		if len(font) < 1 {
			return
		}
		// check whether already in config file
		found := false
		for i := range ezcomm.CfgStruc.Fonts {
			if ezcomm.CfgStruc.Fonts[i].Locale == lang {
				if len(font) > 0 && ezcomm.CfgStruc.Fonts[i].Font != font {
					ezcomm.CfgStruc.Fonts[i].Font = font
					found = true
					break
				}
			}
		}
		if !found {
			ezcomm.CfgStruc.Fonts = append(ezcomm.CfgStruc.Fonts, ezcomm.EzcommFonts{
				Locale: lang,
				Font:   font})

		}
		ezcomm.WriteCfg()
	})

	abtRow := container.NewCenter(widget.NewLabel(ezcomm.Ver + "." + ezcomm.Bld))
	return container.NewVBox(flowFnTxt, flowFlBut, flowFnStt,
		logTxt, logBut, rowVerbose, verboseSel,
		langSel, fontSel, fontBut, abtRow)
}

func verbose2Str() string {
	switch ezcomm.CfgStruc.Verbose {
	case 0:
		return ezcomm.StrNon
	case 1:
		return ezcomm.StrLow
	case 2:
		return ezcomm.StrMdm
	case 3:
		return ezcomm.StrHgh
	}
	return ""
}

func verboseFrmStr(str string) int {
	switch str {
	case ezcomm.StrHgh:
		return 3
	case ezcomm.StrMdm:
		return 2
	case ezcomm.StrLow:
		return 1
	case ezcomm.StrNon:
		return 0
	}
	return 0
}
