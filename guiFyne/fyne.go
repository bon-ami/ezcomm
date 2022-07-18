package main

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	thm        theme4Fonts
	appStorage fyne.Storage
	f          GuiFyne
)

// GuiFyne implements Guis
type GuiFyne struct{}

func (g GuiFyne) Run(ver, bld string) {
	f = g
	// Ver & Bld will be overwritten by GuiRun()
	if len(ver) < 1 {
		ver = "dev"
	}
	if len(bld) < 1 {
		bld = time.Now().Format("2006-01-02_15:04:05")
	}
	ezcomm.Ver = ver
	ezcomm.Bld = bld

	ezcApp := app.NewWithID(ezcomm.EzcName)
	/*uri, err := storage.Child(ezcApp.Storage().RootURI(), ezcomm.EzcName+".xml")
	var cfg string
	if err == nil {
		cfg = uri.String()
	}
	v, err := storage.CanWrite(uri)
	if v == false || err != nil {
		log.Println("Error E1: ", err)
		return
	}
	storer, err := storage.Writer(uri)
	defer storer.Close()
	_, err = storer.Write([]byte("o"))
	if err != nil {
		log.Println("write fail", err)
	} else {
		log.Println("write ok", uri)
	}*/
	appStorage = ezcApp.Storage()
	cfgFileName := ezcomm.EzcName + ".xml"
	rdr, err := appStorage.Open(cfgFileName)
	if err != nil {
		eztools.Log("failed to open config file", err)
	}
	ezcomm.ReaderCfg(rdr, "")

	useFontFromCfg(true, "")

	meta := ezcApp.Metadata()
	if len(meta.Version) > 0 && meta.Version != "0.0.0.0" {
		ezcomm.Ver = meta.Version
	}
	if meta.Build > 0 {
		ezcomm.Bld = strconv.Itoa(meta.Build)
	}
	/*icon, err := fyne.LoadResource("Icon.png")
	if err == nil {*/
	ezcApp.SetIcon(Icon)
	//}
	ezcApp.Settings().SetTheme(&thm)
	ezcWin := ezcApp.NewWindow(ezcomm.EzcName)

	contLcl := makeControlsLcl()
	contRmt := makeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)

	tabs := container.NewAppTabs(
		container.NewTabItem(ezcomm.StringTran["StrInt"], cont),
		container.NewTabItem(ezcomm.StringTran["StrInfLog"], makeControlsInfLog()),
		container.NewTabItem(ezcomm.StringTran["StrCfg"], makeControlsCfg(ezcWin)),
	)
	ezcWin.SetContent(tabs)

	/*selectCtrls = []*widget.SelectEntry{ sockLcl[0], sockLcl[1], //sockRmt[0], sockRmt[1],
		//sockRcv[0], sockRcv[0], //recRmt,
	}*/

	ezcWin.Show()
	/*if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("to show UI")
	}*/
	ezcApp.Run()
	/*if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("UI done")
	}*/
}

func validateInt64(str string) error {
	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		return err
	}
	return nil
}
