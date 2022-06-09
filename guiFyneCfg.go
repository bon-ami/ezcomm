package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
)

func guiFyneMakeControlsCfg(ezcWin fyne.Window) *fyne.Container {
	// flow part begins
	flowFnStt := widget.NewEntry()
	flowFnStt.PlaceHolder = StrStb
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
	// flow part ends

	langMap := make(map[string]string)
	logTxt := widget.NewEntry()
	if len(cfgStruc.LogFile) > 0 {
		logTxt.SetText(cfgStruc.LogFile)
	} else {
		logTxt.SetText(StrLog)
	}
	logTxt.Disable()
	logBut := widget.NewButton(StrLog, func() {
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
				cfgStruc.LogFile = fn
				writeCfg()
				logTxt.SetText(fn)
				logTxt.Refresh()
			}
		}, ezcWin)
	})
	rowVerbose := container.NewCenter(widget.NewLabel(StrVbs))
	verboseSel := widget.NewSelect(nil, func(lvl string) {
		newLvl := verboseFrmStr(lvl)
		if newLvl == cfgStruc.Verbose {
			return
		}
		eztools.Verbose = newLvl
		cfgStruc.Verbose = newLvl
		writeCfg()
	})
	verboseSel.Options = []string{
		StrHgh, StrMdm, StrLow, StrNon,
	}
	verboseSel.SetSelected(verbose2Str())

	fontSel := widget.NewSelect(nil, nil)
	fontSel.PlaceHolder = StrFnt
	fontSel.Options = listSystemFonts()
	currIndx := matchSystemFontsFromPath(cfgStruc.font)
	if currIndx >= 0 {
		fontSel.SetSelectedIndex(currIndx)
	}

	langSel := widget.NewSelect(nil, func(str string) {
		//eztools.LogWtTime("selecting", str)
		//loadStr(langMap[str])
		if cfgStruc.Language == langMap[str] {
			return
		}
		cfgStruc.Language = langMap[str]
		writeCfg()

		i18nLoad(cfgStruc.Language)
		matchFontFromLanguage()
		thm.SetFont(cfgStruc.font)
		indx := matchSystemFontsFromPath(cfgStruc.font)
		if indx != eztools.InvalidID {
			fontSel.SetSelectedIndex(indx)
			//eztools.Log("font", indx)
		} else {
			fontSel.ClearSelected()
			//eztools.Log("font cleared")
		}
		dialog.ShowInformation(StrLang, StrReboot4Change, ezcWin)
	})
	langSel.PlaceHolder = StrLang
	eztools.ListLanguages(func(name, id string) {
		full := id + "_" + name
		langMap[full] = id
		langSel.Options = append(langSel.Options, full)
		if cfgStruc.Language == id {
			langSel.SetSelected(full)
		}
	})
	indx := matchSystemFontsFromPath(cfgStruc.font)
	if indx != eztools.InvalidID {
		fontSel.SetSelectedIndex(indx)
	}

	fontBut := widget.NewButton(StrFnt4Lang, func() {
		lang := langMap[langSel.Selected]
		if len(lang) < 1 {
			return
		}
		font := matchSystemFontsFromIndex(fontSel.SelectedIndex())
		if len(font) < 1 {
			return
		}
		// check whether already in config file
		found := false
		for i := range cfgStruc.Fonts {
			if cfgStruc.Fonts[i].Locale == lang {
				if len(font) > 0 && cfgStruc.Fonts[i].Font != font {
					cfgStruc.Fonts[i].Font = font
					found = true
					break
				}
			}
		}
		if !found {
			cfgStruc.Fonts = append(cfgStruc.Fonts, ezcommFonts{
				Locale: lang,
				Font:   font})

		}
		writeCfg()
	})

	abtRow := container.NewCenter(widget.NewLabel(Ver + "." + Bld))
	return container.NewVBox(flowFnTxt, flowFlBut, flowFnStt,
		logTxt, logBut, rowVerbose, verboseSel,
		langSel, fontSel, fontBut, abtRow)
}

func verbose2Str() string {
	switch cfgStruc.Verbose {
	case 0:
		return StrNon
	case 1:
		return StrLow
	case 2:
		return StrMdm
	case 3:
		return StrHgh
	}
	return ""
}

func verboseFrmStr(str string) int {
	switch str {
	case StrHgh:
		return 3
	case StrMdm:
		return 2
	case StrLow:
		return 1
	case StrNon:
		return 0
	}
	return 0
}
