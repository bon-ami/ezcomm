package main

import (
	"io"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	ezcomm.FlowReaderNew = func(p string) (io.ReadCloser, error) {
		uri := storage.NewFileURI(p)
		return storage.Reader(uri)
	}
	ezcomm.FlowWriterNew = func(p string) (io.WriteCloser, error) {
		uri, err := encodeFilePath(p)
		if err != nil {
			return nil, err
		}
		if b, err := storage.CanWrite(uri); err != nil {
			return nil, err
		} else {
			if !b {
				return nil, eztools.ErrAccess
			}
		}
		return storage.Writer(uri)
	}
	return container.NewTabItem(ezcomm.StringTran["StrCfg"],
		makeControlsCfg())
}

// tryWriteFile prompts user to create the file,
//    if no writer able to be created.
// Return values: error string and whether to abort
func tryWriteFile(fn string) (res string, abt bool) {
	wr, err := ezcomm.FlowWriterNew(fn)
	if err == nil {
		wr.Close()
		return
	}
	ch := make(chan bool, 1)
	dialog.ShowConfirm(
		ezcomm.StringTran["StrFileAwareness"],
		ezcomm.StringTran["StrFileSave"]+
			"\n"+fn,
		func(ret bool) {
			ch <- ret
		},
		ezcWin)
	if !<-ch {
		return "", true
	}
	dialog.ShowFileSave(func(wr fyne.URIWriteCloser, err error) {
		if err == nil && wr != nil {
			wr.Close()
		}
		if err == nil {
			ch <- true
		} else {
			res = err.Error()
			ch <- false
		}
	}, ezcWin)
	if !<-ch {
		return res, true
	}
	return
}

// chkFlowStruc go over all steps and check file existence for output
// Android requires user to allow file creation
// Return value: whether to abort flow
func chkFlowStruc(flow ezcomm.FlowStruc) (string, bool) {
	/*switch runtime.GOOS {
	case "android":
	default:
		return false
	}*/
	var loopSteps func(ezcomm.FlowConnStruc, []ezcomm.FlowStepStruc) (string, bool)
	loopSteps = func(conn ezcomm.FlowConnStruc, stp []ezcomm.FlowStepStruc) (string, bool) {
		for _, step := range stp {
			if step.Act == ezcomm.FlowActRcv &&
				len(step.Data) > 0 {
				fn, fil := step.ParseData(flow, conn)
				if fil == 1 {
					if res, ret := tryWriteFile(fn); ret {
						return res, ret
					}
				}
			}
			if res, ng := loopSteps(conn, step.Steps); ng {
				return res, true
			}
		}
		return "", false
	}
	for _, conn := range flow.Conns {
		if res, ng := loopSteps(conn, conn.Steps); ng {
			return res, true
		}
	}
	return "", false
}

func runFlow(uri fyne.URIReadCloser) {
	var (
		err    error
		resStr string
	)
	resTxt := ezcomm.StringTran["StrFlowRunNot"]
	defer func() {
		if len(resStr) > 0 {
			resTxt = ezcomm.StringTran["StrFlowFinAs"] + resStr
		}
		flowFnStt.SetText(resTxt)
		flowFnStt.Refresh()
		flowFlBut.Enable()
	}()
	flow, err := ezcomm.ReadFlowReader(uri)
	if err != nil {
		Log("flow file NOT read", err)
		return
	}
	if res, ng := chkFlowStruc(flow); ng {
		Log("flow output creation skipped")
		if len(res) > 0 {
			resTxt = res
		}
		return
	}
	if ezcomm.RunFlow(flow) {
		resStr = ezcomm.StringTran["StrOK"]
	} else {
		resStr = ezcomm.StringTran["StrNG"]
	}
}

var (
	flowFlBut, fontBut, currLngBut *widget.Button
	flowFnStt                      *widget.Entry
	flowResChn                     chan bool
	currLang                       string
	langResMap                     map[string]int
)

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
	flowFnStt = widget.NewEntry()
	flowFnStt.PlaceHolder = ezcomm.StringTran["StrStb"]
	flowFnStt.Disable()
	flowFnTxt := widget.NewEntry()
	flowFnTxt.SetText(ezcomm.StringTran["StrFlw"])
	flowFnTxt.Disable()
	flowResChn = make(chan bool, 1)
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
			flowFnTxt.SetText(decodeFilePath(uri.URI()))
			flowFnTxt.Refresh()
			go runFlow(uri)
		}, ezcWin)
	})
	// flow part ends

	rowFont := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrFnt"]))
	fontSel = widget.NewSelect(nil, func(font string) {
		suggestion, builtin := chkFontBltIn()
		if builtin && len(suggestion) > 0 {
			dialog.ShowInformation(ezcomm.StringTran["StrLang"],
				ezcomm.StringTran["StrFnt4LangBuiltin"]+" "+suggestion, ezcWin)
		}
	})
	fontSel.PlaceHolder = ezcomm.StringTran["StrFnt"]
	langResMap = make(map[string]int)
	for i, res := range LangsBuiltin {
		langResMap[res.locale] = i
		if res.fnt == nil {
			continue
		}
		fontSel.Options = append(fontSel.Options, res.locale)
	}
	fontsNumBuiltin = len(fontSel.Options)
	for _, font := range ezcomm.ListSystemFonts([]string{".ttf"}) {
		fontSel.Options = append(fontSel.Options, font)
	}

	rowLang := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrLang"]))
	/*langMap := make(map[string]string)
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
		dialog.ShowInformation(ezcomm.StringTran["StrLang"], prevMsgLang+
			ezcomm.StringTran["StrReboot4Change"], ezcWin)
	})
	langSel.PlaceHolder = ezcomm.StringTran["StrLang"]*/
	langImgs := make([]fyne.CanvasObject, 0)
	langButs := make([]fyne.CanvasObject, 0)
	langId2But := make(map[string]*widget.Button)
	eztools.ListLanguages(func(name, id string) {
		icon := LangsBuiltin[langResMap[id]].name
		langSelFun := func() {
			prevMsgLang := ezcomm.StringTran["StrReboot4Change"] + "\n"
			if ezcomm.CfgStruc.Language == id {
				return
			}
			ezcomm.CfgStruc.Language = id
			writeCfg()

			lang, err := ezcomm.I18nLoad(ezcomm.CfgStruc.Language)
			if err != nil {
				Log("cannot set language", ezcomm.CfgStruc.Language, err)
				return
			}
			if currLngBut != nil {
				currLngBut.Enable()
			}
			currLngBut = langId2But[id]
			currLngBut.Disable()
			ezcomm.CfgStruc.Language = lang
			ezcomm.MatchFontFromCurrLanguageCfg()
			markFont(useFontFromCfg(true, lang))
			dialog.ShowInformation(ezcomm.StringTran["StrLang"], prevMsgLang+
				ezcomm.StringTran["StrReboot4Change"], ezcWin)
		}
		langBut1 := widget.NewButton(id, langSelFun)
		imgContainer := container.NewGridWrap(fyne.NewSize(LangsBuiltin[langResMap[id]].width,
			LangsBuiltin[langResMap[id]].height), canvas.NewImageFromResource(icon))
		langId2But[id] = langBut1
		if ezcomm.CfgStruc.Language == id {
			//langSel.SetSelected(full)
			currLang = id
			currLngBut = langBut1
			langBut1.Disable()
		}
		langImgs = append(langImgs, imgContainer)
		langButs = append(langButs, langBut1)
		//full := id + "_" + name
		//langMap[full] = id
		//langSel.Options = append(langSel.Options, full)
	})
	markFont(useFontFromCfg(false, ezcomm.CfgStruc.Language))
	for _, but := range langButs {
		langImgs = append(langImgs, but)
	}
	langSel := container.NewGridWithColumns(len(langButs), langImgs...)

	fontBut = widget.NewButton(ezcomm.StringTran["StrFnt4Lang"], func() {
		//lang := langMap[langSel.Selected]
		if len(currLang) < 1 {
			return
		}
		saveFontFromIndx(currLang)
		dialog.ShowInformation(ezcomm.StringTran["StrLang"], ezcomm.StringTran["StrReboot4Change"], ezcWin)
	})

	abtRow := container.NewCenter(widget.NewLabel(ezcomm.Ver + " - " + ezcomm.Bld))
	return container.NewVBox(rowFloodLbl, rowFloodEnt,
		flowFnTxt, flowFlBut, flowFnStt, rowLang, langSel,
		rowFont, fontSel /*fontRch,*/, fontBut, abtRow)
}

// saveFontFromIndx
// Parameter: local
func saveFontFromIndx(lang string) {
	var font string
	indx := fontSel.SelectedIndex()
	if indx < fontsNumBuiltin {
		font = LangsBuiltin[indx].locale
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
	i, ok := langResMap[ezcomm.CfgStruc.Language]
	if !ok {
		return
	}
	if LangsBuiltin[i].fnt == nil || fontSel.Selected == LangsBuiltin[i].locale {
		return "", true
	}
	return LangsBuiltin[i].locale, true
}

func useFontFromCfg(setTheme bool, lang string) (fontPath string,
	fontStaticIndx int) {
	fontStaticIndx = eztools.InvalidID
	cfg := ezcomm.CfgStruc.GetFont()
	if len(lang) < 1 && len(cfg) < 1 {
		return
	}
	//Log("setting font=", cfg, ", lang=", lang)
	i := 0
	for _, res := range LangsBuiltin {
		//Log("checking built-in", fontBuiltin.locale)
		if res.fnt == nil {
			continue
		}
		if len(cfg) < 1 {
			if lang == res.locale {
				ezcomm.CfgStruc.SetFont(lang)
				thm.SetFontByRes(res.fnt)
				return "", i
			}
		}
		if cfg == res.locale {
			if setTheme {
				thm.SetFontByRes(res.fnt)
			}
			return "", i
		}
		// i != index of LangsBuiltin, but from all res.fnt != nil
		i += 1
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
	flowFlBut.Refresh()
	fontBut.Refresh()
}
