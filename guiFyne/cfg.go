package guiFyne

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var cfgWriter fyne.URIWriteCloser

func writeCfg() {
	if cfgWriter == nil {
		cfgFileName := ezcomm.EzcName + ".xml"
		var err error
		cfgWriter, err = appStorage.Save(cfgFileName)
		if err != nil {
			/* TODO: fyne returns customized errors so I cannot check it now
			if !errors.Is(err, os.ErrNotExist) {
				eztools.Log("failed to write to config file", err)
				return
			}*/
			cfgWriter, err = appStorage.Create(cfgFileName)
			if err != nil {
				eztools.Log("failed to create to config file", err)
				return
			}
			return
		}
	}
	ezcomm.WriterCfg(cfgWriter)
}

func makeControlsCfg(ezcWin fyne.Window) *fyne.Container {
	// flow part begins
	flowFnStt := widget.NewEntry()
	flowFnStt.PlaceHolder = ezcomm.StringTran["StrStb"]
	flowFnStt.Disable()
	flowFnTxt := widget.NewEntry()
	flowFnTxt.SetText(ezcomm.StringTran["StrFlw"])
	flowFnTxt.Disable()
	var flowFlBut *widget.Button
	flowFlBut = widget.NewButton(ezcomm.StringTran["StrFlw"], func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			flowFnStt.SetText("")
			defer flowFnStt.Refresh()
			if err != nil {
				eztools.LogWtTime("open flow file", err)
				flowFnTxt.SetText(ezcomm.StringTran["StrFlw"])
				flowFnStt.SetText(ezcomm.StringTran["StrFlowOpenErr"])
			}
			if uri == nil {
				flowFnStt.SetText(ezcomm.StringTran["StrFlowOpenNot"])
				return
			}
			flowFlBut.Disable()
			flowFnStt.SetText(ezcomm.StringTran["StrFlowRunning"])
			flowFnStt.Refresh()
			fn := uri.URI().String()
			flowFnTxt.SetText(fn)
			flowFnTxt.Refresh()
			resChn := make(chan bool)
			if !ezcomm.RunFlowReaderBG(uri, resChn) {
				eztools.LogWtTime("flow file NOT run", fn)
				flowFnStt.SetText(ezcomm.StringTran["StrFlowRunNot"])
				flowFnStt.Refresh()
				flowFlBut.Enable()
				return
			}
			go func() {
				var resStr string
				if <-resChn {
					resStr = ezcomm.StringTran["StrOK"]
				} else {
					resStr = ezcomm.StringTran["StrNG"]
				}
				flowFnStt.SetText(ezcomm.StringTran["StrFlowFinAs"] + resStr)
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
		logTxt.SetText(ezcomm.StringTran["StrLog"])
	}
	logTxt.Disable()
	logBut := widget.NewButton(ezcomm.StringTran["StrLog"], func() {
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
		ezcomm.StringTran["StrHh"], ezcomm.StringTran["StrMdm"], ezcomm.StringTran["StrLow"], ezcomm.StringTran["StrNon"],
	}
	verboseSel.SetSelected(verbose2Str())

	fontSel := widget.NewSelect(nil, nil)
	fontSel.PlaceHolder = ezcomm.StringTran["StrFnt"]
	fontSel.Options = ezcomm.ListSystemFonts()
	currIndx := ezcomm.MatchSystemFontsFromPath(ezcomm.CfgStruc.GetFont())
	if currIndx >= 0 {
		fontSel.SetSelectedIndex(currIndx)
	}

	eztools.Log("rich ", ezcomm.StringTran["StrFntRch"])
	fontRch := widget.NewRichTextWithText(ezcomm.StringTran["StrFntRch"])
	langSel := widget.NewSelect(nil, func(str string) {
		//eztools.LogWtTime("selecting", str)
		//loadStr(langMap[str])
		if ezcomm.CfgStruc.Language == langMap[str] {
			return
		}
		ezcomm.CfgStruc.Language = langMap[str]
		writeCfg()

		lang, err := ezcomm.I18nLoad(ezcomm.CfgStruc.Language)
		if err != nil {
			f.Log(true, "cannot set language", ezcomm.CfgStruc.Language, err)
			return
		}
		ezcomm.CfgStruc.Language = lang
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
		for _, v := range fontRch.Segments {
			f.Log(true, "richtext seg", v.Textual())
		}
		dialog.ShowInformation(ezcomm.StringTran["StrLang"], ezcomm.StringTran["StrReboot4Change"], ezcWin)
	})
	langSel.PlaceHolder = ezcomm.StringTran["StrLang"]
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

	fontBut := widget.NewButton(ezcomm.StringTran["StrFnt4Lang"], func() {
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
		writeCfg()
	})

	abtRow := container.NewCenter(widget.NewLabel(ezcomm.Ver + " - " + ezcomm.Bld))
	return container.NewVBox(flowFnTxt, flowFlBut, flowFnStt,
		logTxt, logBut, rowVerbose, verboseSel,
		langSel, fontSel /*fontRch,*/, fontBut, abtRow)
}

func verbose2Str() string {
	switch ezcomm.CfgStruc.Verbose {
	case 0:
		return ezcomm.StringTran["StrNon"]
	case 1:
		return ezcomm.StringTran["StrLow"]
	case 2:
		return ezcomm.StringTran["StrMdm"]
	case 3:
		return ezcomm.StringTran["StrHgh"]
	}
	return ""
}

func verboseFrmStr(str string) int {
	switch str {
	case ezcomm.StringTran["StrHgh"]:
		return 3
	case ezcomm.StringTran["StrMdm"]:
		return 2
	case ezcomm.StringTran["StrLow"]:
		return 1
	case ezcomm.StringTran["StrNon"]:
		return 0
	}
	return 0
}