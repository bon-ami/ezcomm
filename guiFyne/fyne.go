package guiFyne

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
	"gitlab.com/bon-ami/ezcomm/res"
)

var (
	thm        theme4Fonts
	appStorage fyne.Storage
)

type GuiFyne struct{}

func (GuiFyne) GuiSetGlbPrm(Ver, Bld string) {
	// Ver & Bld will be overwritten by GuiRun()
	if len(Ver) < 1 {
		Ver = "dev"
	}
	if len(Bld) < 1 {
		Bld = time.Now().Format("2006-01-02_15:04:05")
	}
	ezcomm.Ver = Ver
	ezcomm.Bld = Bld

	ezcomm.LogPrintFunc = eztools.Log
	ezcomm.GuiConnected = Connected
	ezcomm.GuiEnded = Ended
	ezcomm.GuiLog = Log
	ezcomm.GuiRcv = Rcv
	ezcomm.GuiSnt = Snt
}

func (GuiFyne) GuiRun() {
	//eztools.Log("setting font=", cfgStruc.font)
	fontStr := ezcomm.CfgStruc.GetFont()
	if len(fontStr) > 0 {
		err := thm.SetFont(fontStr)
		if err != nil {
			eztools.Log("failed to set font", fontStr, err)
		} else {
			if eztools.Debugging && eztools.Verbose > 2 {
				eztools.Log("font set", fontStr)
			}
		}
	}
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

	meta := ezcApp.Metadata()
	if len(meta.Version) > 0 && meta.Version != "0.0.0.0" {
		ezcomm.Ver = meta.Version
	}
	if meta.Build > 0 {
		ezcomm.Bld = strconv.Itoa(meta.Build)
	}
	/*icon, err := fyne.LoadResource("Icon.png")
	if err == nil {*/
	ezcApp.SetIcon(res.ResourceIconPng)
	//}
	ezcApp.Settings().SetTheme(&thm)
	ezcWin := ezcApp.NewWindow(ezcomm.EzcName)

	contLcl := makeControlsLcl()
	contRmt := makeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)

	tabs := container.NewAppTabs(
		container.NewTabItem(ezcomm.StringTran["StrInt"], cont),
		container.NewTabItem(ezcomm.StringTran["StrInfLog"], guiFyneMakeControlsInfLog()),
		container.NewTabItem(ezcomm.StringTran["StrCfg"], guiFyneMakeControlsCfg(ezcWin)),
	)
	ezcWin.SetContent(tabs)

	/*selectCtrls = []*widget.SelectEntry{ sockLcl[0], sockLcl[1], //sockRmt[0], sockRmt[1],
		//sockRcv[0], sockRcv[0], //recRmt,
	}*/

	ezcWin.Show()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("to show UI")
	}
	ezcApp.Run()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("UI done")
	}
	if cfgWriter != nil {
		cfgWriter.Close()
	}
}
