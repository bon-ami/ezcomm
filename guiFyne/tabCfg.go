package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	fontSel         *widget.Select
	fontsNumBuiltin int
)

func writeCfg() {
	cfgFileName := ezcomm.EzcName + ".xml"
	cfgWriter, err := appStorage.Save(cfgFileName)
	if err != nil {
		/* TODO: fyne returns customized errors so I cannot check it now
		if !errors.Is(err, os.ErrNotExist) {
			Log("failed to write to config file", err)
			return
		}*/
		cfgWriter, err = appStorage.Create(cfgFileName)
		if err != nil {
			Log("failed to create config file", err)
			return
		}
		return
	}
	ezcomm.WriterCfg(cfgWriter)
}

func makeTabCfg() *container.TabItem {
	return container.NewTabItem(ezcomm.StringTran["StrCfg"],
		makeControlsCfg())
}

func makeControlsCfg() *fyne.Container {
	rowFloodLbl := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrAntiFld"]))
	floodLblLmt := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLmt"]))
	floodInpLmt := widget.NewEntry()
	floodInpLmt.SetText(strconv.FormatInt(ezcomm.CfgStruc.AntiFlood.Limit, 10))
	floodInpLmt.Validator = validateInt64
	floodInpLmt.OnChanged = func(str string) {
		if validateInt64(str) != nil {
			return
		}
		ezcomm.CfgStruc.AntiFlood.Limit, _ = strconv.ParseInt(str, 10, 64)
		ezcomm.AntiFlood.Limit = ezcomm.CfgStruc.AntiFlood.Limit
		writeCfg()
	}
	floodLblPrd := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrPrd"]))
	floodInpPrd := widget.NewEntry()
	floodInpPrd.SetText(strconv.FormatInt(ezcomm.CfgStruc.AntiFlood.Period, 10))
	floodInpPrd.Validator = validateInt64
	floodInpPrd.OnChanged = func(str string) {
		if validateInt64(str) != nil {
			return
		}
		ezcomm.CfgStruc.AntiFlood.Period, _ = strconv.ParseInt(str, 10, 64)
		ezcomm.AntiFlood.Period = ezcomm.CfgStruc.AntiFlood.Period
		writeCfg()
	}
	rowFloodEnt := container.NewGridWithRows(2,
		floodLblLmt, floodInpLmt, floodLblPrd, floodInpPrd)

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
				Log("open flow file", err)
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
				Log("flow file NOT run", fn)
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
	rowFont := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrFnt"]))
	fontSel = widget.NewSelect(nil, func(font string) {
		suggestion, builtin := chkFontBltIn()
		if builtin && len(suggestion) > 0 {
			dialog.ShowInformation(ezcomm.StringTran["StrLang"],
				ezcomm.StringTran["StrFnt4LangBuiltin"]+" "+suggestion, ezcWin)
		}
	})
	fontSel.PlaceHolder = ezcomm.StringTran["StrFnt"]
	for _, font := range FontsBuiltin {
		fontSel.Options = append(fontSel.Options, font.locale)
	}
	fontsNumBuiltin = len(fontSel.Options)
	for _, font := range ezcomm.ListSystemFonts([]string{".ttf"}) {
		fontSel.Options = append(fontSel.Options, font)
	}

	fontRch := widget.NewRichTextWithText(ezcomm.StringTran["StrFntRch"])

	rowLang := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLang"]))
	langSel := widget.NewSelect(nil, func(str string) {
		prevMsgLang := ezcomm.StringTran["StrReboot4Change"] + "\n"
		//prevMsgFont := ezcomm.StringTran["StrFnt4LangBuiltin"] + "\n"
		//Log("selecting", str)
		//loadStr(langMap[str])
		if ezcomm.CfgStruc.Language == langMap[str] {
			return
		}
		ezcomm.CfgStruc.Language = langMap[str]
		writeCfg()

		lang, err := ezcomm.I18nLoad(ezcomm.CfgStruc.Language)
		if err != nil {
			Log("cannot set language", ezcomm.CfgStruc.Language, err)
			return
		}
		ezcomm.CfgStruc.Language = lang
		ezcomm.MatchFontFromCurrLanguageCfg()
		markFont(useFontFromCfg(true, lang))
		for _, v := range fontRch.Segments {
			Log("richtext seg", v.Textual())
		}
		dialog.ShowInformation(ezcomm.StringTran["StrLang"], prevMsgLang+
			ezcomm.StringTran["StrReboot4Change"], ezcWin)
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
	markFont(useFontFromCfg(false, ezcomm.CfgStruc.Language))

	fontBut := widget.NewButton(ezcomm.StringTran["StrFnt4Lang"], func() {
		lang := langMap[langSel.Selected]
		if len(lang) < 1 {
			return
		}
		saveFontFromIndx(lang)
		dialog.ShowInformation(ezcomm.StringTran["StrLang"], ezcomm.StringTran["StrReboot4Change"], ezcWin)
	})

	abtRow := container.NewCenter(widget.NewLabel(ezcomm.Ver + " - " + ezcomm.Bld))
	return container.NewVBox(rowFloodLbl, rowFloodEnt,
		flowFnTxt, flowFlBut, flowFnStt, rowLang, langSel,
		rowFont, fontSel /*fontRch,*/, fontBut, abtRow)
}

func saveFontFromIndx(lang string) {
	var font string
	indx := fontSel.SelectedIndex()
	if indx < fontsNumBuiltin {
		font = FontsBuiltin[indx].locale
	} else {
		font = ezcomm.MatchSystemFontsFromIndex(indx)
	}
	if len(font) < 1 {
		Log("NO font found!", lang)
		return
	}
	// check whether already in config file
	found := false
	for i := range ezcomm.CfgStruc.Fonts {
		if ezcomm.CfgStruc.Fonts[i].Locale == lang {
			if len(ezcomm.CfgStruc.Fonts[i].Font) > 0 {
				if ezcomm.CfgStruc.Fonts[i].Font != font {
					ezcomm.CfgStruc.Fonts[i].Font = font
				}
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
}

func chkFontBltIn() (suggestion string, builtin bool) {
	indx := fontSel.SelectedIndex()
	if indx >= 0 && indx < fontsNumBuiltin {
		if ezcomm.CfgStruc.Language == FontsBuiltin[indx].locale {
			return "", true
		}
	}
	for _, fontBuiltin := range FontsBuiltin {
		if ezcomm.CfgStruc.Language == fontBuiltin.locale {
			return fontBuiltin.locale, true
		}
	}
	return
}

func useFontFromCfg(setTheme bool, lang string) (fontPath string,
	fontStaticIndx int) {
	fontStaticIndx = eztools.InvalidID
	cfg := ezcomm.CfgStruc.GetFont()
	if len(lang) < 1 && len(cfg) < 1 {
		return
	}
	//Log("setting font=", cfg, ", lang=", lang)
	for i, fontBuiltin := range FontsBuiltin {
		//Log("checking built-in", fontBuiltin.locale)
		if len(cfg) < 1 {
			if lang == fontBuiltin.locale {
				ezcomm.CfgStruc.SetFont(lang)
				thm.SetFontByRes(fontBuiltin.res)
				return "", i
			}
		}
		if cfg == fontBuiltin.locale {
			if setTheme {
				thm.SetFontByRes(fontBuiltin.res)
			}
			return "", i
		}
	}
	if len(cfg) < 1 {
		return
	}
	fontPath = cfg
	if !setTheme {
		return
	}
	err := thm.SetFontByDir(cfg)
	if err != nil {
		Log("failed to set font", cfg, err)
	} else {
		if eztools.Debugging && eztools.Verbose > 2 {
			Log("font set", cfg)
		}
	}
	return
}

func markFont(dir string, indx int) {
	//Log("marking", dir, indx)
	if indx < 0 {
		indx = ezcomm.MatchSystemFontsFromPath(dir)
		if indx != eztools.InvalidID {
			indx += fontsNumBuiltin
		}
	}
	if indx >= 0 {
		fontSel.SetSelectedIndex(indx)
	} else {
		fontSel.ClearSelected()
		//Log("font cleared")
	}
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

func tabCfgShown() {

}
