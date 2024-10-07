package main

import (
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	fontSel         *widget.Select
	flowFnTxt       *widget.Entry
	fontsNumBuiltin int
	tabCfgInit      bool
)

func writeCfg() {
	cfgFileName := ezcomm.EzcName + ".xml"
	cfgWriter, err := appStorage.Save(cfgFileName)
	if err != nil {
		if !errors.Is(err, storage.ErrNotExists) {
			Log("failed to write to config file", err)
			return
		}
		cfgWriter, err = appStorage.Create(cfgFileName)
		if err != nil {
			Log("failed to create config file", err)
			return
		}
	}
	if err = ezcomm.WriterCfg(cfgWriter); err != nil {
		Log("failed to write config", err)
	}
}

func writerNew(p string) (io.WriteCloser, error) {
	uri, err := encodeFilePath(p)
	if err != nil {
		return nil, err
	}
	if b, err := storage.CanWrite(uri); err == nil {
		if !b {
			return nil, eztools.ErrAccess
		}
	} else {
		return nil, err
	}
	return storage.Writer(uri)
}

func makeTabCfg() *container.TabItem {
	ezcomm.FlowReaderNew = func(p string) (io.ReadCloser, error) {
		uri := storage.NewFileURI(p)
		return storage.Reader(uri)
	}
	ezcomm.FlowWriterNew = func(p string) (io.WriteCloser, error) {
		dldPath, err := checkDldDir()
		if err != nil {
			return nil, err
		}
		if len(dldPath) < 1 || len(p) < 1 {
			return nil, eztools.ErrAccess
		}
		return writerNew(filepath.Join(dldPath, filepath.Base(p)))
	}
	return container.NewTabItem(ezcomm.StringTran["StrCfg"],
		makeControlsCfg())
}

// treWriteFile overwrites the file with the name under Downloads of app,
//
//	and prompts user to create one,
//	if no writer able to be created.
//	The existing file will be truncated.
//	DO NOT call this in UI thread!
//
// Return values:
//
//	wc=Close() to be called by caller
//	res=error string, or selected file path by user
//	abt=whether to abort
func tryWriteFile(fun func(string) (io.WriteCloser, error),
	fn string) (wc io.WriteCloser, res string, abt bool) {
	wr, err := fun(fn)
	if err == nil {
		return wr, fn, false
	}
	ch := make(chan bool, 1)
	defer close(ch)
	dialog.ShowConfirm(
		ezcomm.StringTran["StrFileAwareness"],
		ezcomm.StringTran["StrFileSave"]+
			"\n"+fn,
		func(ret bool) {
			ch <- ret
		},
		ezcWin)
	if !<-ch {
		return nil, "", true
	}
	dialog.ShowFileSave(func(wr fyne.URIWriteCloser, err error) {
		if err == nil && wr != nil {
			if !strings.HasPrefix(wr.URI().Name(),
				invalidFileName) {
				res = decodeFilePath(wr.URI())
				wc = wr
				ch <- true
				return
			}
		}
		if wr != nil {
			wr.Close()
		}
		if err != nil {
			res = err.Error()
		}
		ch <- false
	}, ezcWin)
	if !<-ch {
		return wc, res, true
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
	var loopSteps func(ezcomm.FlowConnStruc,
		[]ezcomm.FlowStepStruc) (string, bool)
	loopSteps = func(conn ezcomm.FlowConnStruc,
		stp []ezcomm.FlowStepStruc) (string, bool) {
		for _, step := range stp {
			if step.Act == ezcomm.FlowActRcv &&
				len(step.Data) > 0 {
				fn, fil := step.ParseData(flow, conn)
				if fil == 1 {
					wr, res, ret := tryWriteFile(
						ezcomm.FlowWriterNew, fn)
					// since incoming files are saved under app's dir, we do not need wr
					if wr != nil {
						wr.Close()
					}
					if ret {
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
	currLang                       string
	langResMap                     map[string]int
	langButs                       []fyne.CanvasObject
)

func makeControlsCfg() *fyne.Container {
	rowFloodLbl := container.NewCenter(
		widget.NewLabel(ezcomm.StringTran["StrAntiFld"]))
	floodLblLmt := container.NewCenter(
		widget.NewLabel(ezcomm.StringTran["StrLmt"]))
	floodInpLmt := widget.NewEntry()
	floodInpLmt.SetText(strconv.FormatInt(
		ezcomm.CfgStruc.AntiFlood.Limit, 10))
	floodInpLmt.Validator = validateInt64
	floodInpLmt.OnChanged = func(str string) {
		if validateInt64(str) != nil {
			return
		}
		ezcomm.CfgStruc.AntiFlood.Limit, _ =
			strconv.ParseInt(str, 10, 64)
		ezcomm.AntiFlood.Limit = ezcomm.CfgStruc.AntiFlood.Limit
		writeCfg()
	}
	floodLblPrd := container.NewCenter(
		widget.NewLabel(ezcomm.StringTran["StrPrd"]))
	floodInpPrd := widget.NewEntry()
	floodInpPrd.SetText(strconv.FormatInt(
		ezcomm.CfgStruc.AntiFlood.Period, 10))
	floodInpPrd.Validator = validateInt64
	floodInpPrd.OnChanged = func(str string) {
		if validateInt64(str) != nil {
			return
		}
		ezcomm.CfgStruc.AntiFlood.Period, _ =
			strconv.ParseInt(str, 10, 64)
		ezcomm.AntiFlood.Period = ezcomm.CfgStruc.AntiFlood.Period
		writeCfg()
	}
	rowFloodEnt := container.NewGridWithRows(2,
		floodLblLmt, floodInpLmt, floodLblPrd, floodInpPrd)

	// flow part begins
	flowFnStt = widget.NewEntry()
	flowFnStt.PlaceHolder = ezcomm.StringTran["StrStb"]
	flowFnStt.Disable()
	flowFnTxt = widget.NewEntry()
	flowFnTxt.SetText(ezcomm.StringTran["StrFlw"])
	flowFnTxt.Disable()
	flowFlBut = widget.NewButton(ezcomm.StringTran["StrFlw"], func() {
		dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
			flowFnStt.SetText("")
			defer flowFnStt.Refresh()
			if err != nil {
				Log("open flow file", err)
				flowFnTxt.SetText(ezcomm.StringTran["StrFlw"])
				flowFnStt.SetText(
					ezcomm.StringTran["StrFlowOpenErr"])
			}
			if uri == nil {
				flowFnStt.SetText(
					ezcomm.StringTran["StrFlowOpenNot"])
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
				ezcomm.StringTran["StrFnt4LangBuiltin"]+
					" "+suggestion, ezcWin)
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
	fontSel.Options = append(fontSel.Options, ezcomm.ListSystemFonts([]string{".otf"})...)

	rowLang := container.NewCenter(widget.NewLabel(
		ezcomm.StringTran["StrLang"]))
	langImgs := make([]fyne.CanvasObject, 0)
	langButs = make([]fyne.CanvasObject, 0)
	langID2But := make(map[string]*widget.Button)
	eztools.ListLanguages(func(name, id string) {
		icon := LangsBuiltin[langResMap[id]].name
		langSelFun := func() {
			prevMsgRbt := ezcomm.StringTran["StrReboot4Change"]
			prevMsgOK := ezcomm.StringTran["StrOK"]
			prevMsgLang := ezcomm.StringTran["StrLang"]
			if ezcomm.CfgStruc.Language == id {
				return
			}
			ezcomm.CfgStruc.Language = id
			writeCfg()

			lang, err := ezcomm.I18nLoad(ezcomm.CfgStruc.Language)
			if err != nil {
				Log("cannot set language",
					ezcomm.CfgStruc.Language, err)
				return
			}
			if currLngBut != nil {
				currLngBut.Enable()
			}
			currLngBut = langID2But[id]
			currLngBut.Disable()
			ezcomm.CfgStruc.Language = lang
			ezcomm.MatchFontFromCurrLanguageCfg()
			markFont(useFontFromCfg(true, lang))
			title := prevMsgLang + " " +
				ezcomm.StringTran["StrLang"]
			if LangsBuiltin[langResMap[id]].rbt == nil {
				// ASCII. no icon
				dialog.ShowInformation(title,
					prevMsgRbt+"\n"+ezcomm.StringTran["StrReboot4Change"],
					ezcWin)
				return
			}
			// non-ASCII. show icon
			msg := widget.NewLabel(prevMsgRbt)
			icn := container.NewGridWrap(
				fyne.NewSize(
					LangsBuiltin[langResMap[id]].
						rbtWidth,
					LangsBuiltin[langResMap[id]].
						rbtHeight),
				canvas.NewImageFromResource(
					LangsBuiltin[langResMap[id]].
						rbt))
			formItems := container.NewVBox(msg,
				container.NewCenter(icn))
			dialog.ShowCustom(title,
				prevMsgOK+" "+ezcomm.StringTran["StrOK"],
				formItems, ezcWin)
		}
		langBut1 := widget.NewButton(id, langSelFun)
		imgContainer := container.NewGridWrap(
			fyne.NewSize(LangsBuiltin[langResMap[id]].nameWidth,
				LangsBuiltin[langResMap[id]].nameHeight),
			canvas.NewImageFromResource(icon))
		langID2But[id] = langBut1
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
	langImgs = append(langImgs, langButs...)
	langSel := container.NewGridWithColumns(len(langButs), langImgs...)

	fontBut = widget.NewButton(ezcomm.StringTran["StrFnt4Lang"], func() {
		//lang := langMap[langSel.Selected]
		if len(currLang) < 1 {
			return
		}
		saveFontFromIndx(currLang)
		dialog.ShowInformation(ezcomm.StringTran["StrLang"],
			ezcomm.StringTran["StrReboot4Change"], ezcWin)
	})

	abtRow := container.NewCenter(widget.NewLabel(
		ezcomm.Ver + " - " + ezcomm.Bld))
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
		ezcomm.CfgStruc.Fonts = append(ezcomm.CfgStruc.Fonts,
			ezcomm.Fonts{Locale: lang, Font: font})

	}
	writeCfg()
}

func chkFontBltIn() (suggestion string, builtin bool) {
	i, ok := langResMap[ezcomm.CfgStruc.Language]
	if !ok {
		return
	}
	if LangsBuiltin[i].fnt == nil ||
		fontSel.Selected == LangsBuiltin[i].locale {
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
		i++
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
	if tabCfgInit {
		return
	}
	tabCfgInit = true
	flowFlBut.Refresh()
	flowFnStt.Refresh()
	flowFnTxt.Refresh()
	fontBut.Refresh()
	fontSel.Refresh()
	for _, but := range langButs {
		but.Refresh()
	}
}
