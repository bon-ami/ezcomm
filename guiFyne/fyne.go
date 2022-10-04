package main

import (
	"flag"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	ezcWin     fyne.Window
	thm        theme4Fonts
	appStorage fyne.Storage
	// chn is for TCP client and UDP
	chn [2]chan ezcomm.RoutCommStruc
	// svrTcp is for TCP server only
	svrTcp ezcomm.SvrTcp
)

// uiFyne implements Uis
//type uiFyne struct{}

func parseParams() {
	var paramV, paramVV, paramVVV bool
	flag.BoolVar(&paramV, "v", false, ezcomm.StringTran["StrV"])
	flag.BoolVar(&paramVV, "vv", false, ezcomm.StringTran["StrVV"])
	flag.BoolVar(&paramVVV, "vvv", false, ezcomm.StringTran["StrVVV"])
	flag.Parse()
	switch {
	case paramV:
		eztools.Verbose = 1
	case paramVV:
		eztools.Verbose = 2
	case paramVVV:
		eztools.Verbose = 3
	}
}

const extPrefAnd = "content://com.android.externalstorage.documents/document/primary"

func encodeFilePath(p string) (fyne.URI, error) {
	switch runtime.GOOS {
	case "android":
		// change a/b into {extPref}%3Aa%2Fb meaning {extPref}:a/b
		p = extPrefAnd + url.QueryEscape(":"+p)
		return storage.ParseURI(p)
	default:
		uri := storage.NewFileURI(p)
		if uri == nil || len(uri.String()) < 1 {
			return nil, eztools.ErrInvalidInput
		}
		return uri, nil
	}
}

func decodeFilePath(uri fyne.URI) string {
	switch runtime.GOOS {
	case "android":
		fn, err := url.PathUnescape(uri.String())
		if err != nil {
			Log("failed to unescape", uri)
			return ""
		}
		return strings.TrimPrefix(fn, extPrefAnd+":")
	default:
		return uri.String()
	}
}

func main() {
	parseParams()
	ezcApp := app.NewWithID(ezcomm.EzcName)
	appStorage = ezcApp.Storage()
	cfgFileName := ezcomm.EzcName + ".xml"
	rdr, err := appStorage.Open(cfgFileName)
	if err != nil {
		eztools.Log("failed to open config file", err)
	}
	paramLog, err := ezcomm.ReaderCfg(rdr, "")
	if err != nil {
		err = initLog(paramLog)
		if err != nil {
			eztools.Log("failed to set log file", err)
		}
	}

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
	ezcWin = ezcApp.NewWindow(ezcomm.EzcName)

	tabLog := makeTabLog()
	tabMsg := makeTabMsg()
	tabFil := makeTabFil()
	tabCfg := makeTabCfg()
	tabLan := makeTabLan()
	tabs := container.NewAppTabs(
		tabLan,
		tabMsg,
		tabFil,
		tabLog,
		tabCfg,
	)
	ezcWin.SetContent(tabs)
	tabs.OnSelected = func(tb *container.TabItem) {
		tabLanShown(false)
		switch tb {
		case tabMsg:
			tabMsgShown()
		case tabFil:
			tabFilShown()
		case tabLog:
			tabLogShown()
		case tabCfg:
			tabCfgShown()
		case tabLan:
			tabLanShown(true)
		}
	}

	svrTcp.ActFunc = tcpConnAct
	svrTcp.ConnFunc = TcpSvrConnected
	svrTcp.LogFunc = Log

	ezcWin.Show()
	tabLanShown(true)
	// 9 routines here
	ezcApp.Run()
	if eztools.Debugging {
		eztools.Log("routines (5 is normal) left", runtime.NumGoroutine())
	}
	if logger != nil {
		logger.Close()
	}
}

func validateInt64(str string) error {
	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		return err
	}
	return nil
}
