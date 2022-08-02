package main

import (
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	thm        theme4Fonts
	appStorage fyne.Storage
	// chn is for TCP client and UDP
	chn [2]chan ezcomm.RoutCommStruc
	// svrTcp is for TCP server only
	svrTcp ezcomm.SvrTcp
)

// uiFyne implements Uis
//type uiFyne struct{}

func main() {
	/*for i := range chn {
		chn[i] = make(chan ezcomm.RoutCommStruc, ezcomm.FlowComLen)
	}*/
	ezcApp := app.NewWithID(ezcomm.EzcName)
	appStorage = ezcApp.Storage()
	cfgFileName := ezcomm.EzcName + ".xml"
	rdr, err := appStorage.Open(cfgFileName)
	if err != nil {
		eztools.Log("failed to open config file", err)
	}
	ezcomm.ReaderCfg(rdr, "")

	useFontFromCfg(true, ezcomm.CfgStruc.Language)

	meta := ezcApp.Metadata()
	if len(meta.Version) > 0 && meta.Version != "0.0.0.0" {
		ezcomm.Ver = meta.Version
	}
	if meta.Build > 0 {
		ezcomm.Bld = strconv.Itoa(meta.Build)
	}
	ezcApp.SetIcon(Icon)
	ezcApp.Settings().SetTheme(&thm)
	ezcWin := ezcApp.NewWindow(ezcomm.EzcName)

	tabs := container.NewAppTabs(
		container.NewTabItem(ezcomm.StringTran["StrInt"], makeTabMsg()),
		container.NewTabItem(ezcomm.StringTran["StrInfLog"], makeTabLog()),
		container.NewTabItem(ezcomm.StringTran["StrCfg"], makeTabCfg(ezcWin)),
	)
	ezcWin.SetContent(tabs)

	svrTcp.ActFunc = tcpConnAct
	svrTcp.ConnFunc = TcpSvrConnected
	svrTcp.LogFunc = Log

	ezcWin.Show()
	ezcApp.Run()
	if eztools.Debugging {
		eztools.Log("routines left", runtime.NumGoroutine())
	}
}

func validateInt64(str string) error {
	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		return err
	}
	return nil
}
