package main

import (
	"flag"
	"net/url"
	"path/filepath"
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
	svrTcp         ezcomm.SvrTcp
	tabs           *container.AppTabs
	tabFil, tabLAf *container.TabItem
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

const (
	extPrefAndDoc = "content://com.android.externalstorage.documents/document/primary"
	extPrefTreDoc = "content://com.android.externalstorage.documents/tree/primary"
	sdcardPrefAnd = "/sdcard"
	dldDirNm      = "Downloads"
)

var (
	dldDirChk  bool
	dldDirPath string
)

// checkDldDir checks and creates Downloads under app dir
// Return values: eztools.ErrIncomplete=file exists as same name
func checkDldDir() (string, error) {
	if dldDirChk {
		return dldDirPath, nil
	}
	dldDirChk = true
	incomDir = appStorage.RootURI().Path()
	dldDirPath = filepath.Join(incomDir, dldDirNm)
	dldUri = storage.NewFileURI(dldDirPath)
	exi, err := storage.Exists(dldUri)
	if err != nil {
		eztools.Log("NO", dldDirPath, "detectable!", err)
		return "", err
	}
	if exi {
		cn, err := storage.CanList(dldUri)
		if err != nil {
			eztools.Log("NO", dldDirPath, "listable!", err)
			return "", err
		}
		if !cn {
			/*Log(dirDld, "is a file!", incomDir,
			"will be used as download directory!")*/
			return "", eztools.ErrIncomplete
		}
	} else {
		if err = storage.CreateListable(dldUri); err != nil {
			eztools.Log("NO", dldDirPath, "created!", err)
			return "", err
		}
	}
	return dldDirPath, err
}

func translateFilePath(p string) string {
	/*switch runtime.GOOS {
	case "android":
		// change a/b into {extPref}%3Aa%2Fb meaning {extPref}:a/b
		//p = extPrefAnd + url.QueryEscape(":"+p)
		//return storage.ParseURI(p)
		if len(p) < 1 {
			return sdcardPrefAnd
		}
		if strings.HasPrefix(p, "/") {
			return sdcardPrefAnd + p
		}
		return sdcardPrefAnd + "/" + p
	}*/
	return p
}

func encodeFilePath(p string) (fyne.URI, error) {
	//default:
	uri := storage.NewFileURI(translateFilePath(p))
	if uri == nil || len(uri.String()) < 1 {
		return nil, eztools.ErrInvalidInput
	}
	return uri, nil
}

func encodeFileDown(p string) (u fyne.URI, err error) {
	dldPath, err := checkDldDir()
	if err != nil {
		return
	}
	if len(dldPath) < 1 {
		return nil, eztools.ErrAccess
	}
	if len(p) > 0 {
		return encodeFilePath(
			filepath.Join(dldPath,
				filepath.Base(p)))
	} else {
		return encodeFilePath(dldPath)
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
		return strings.TrimPrefix(
			strings.TrimPrefix(fn, extPrefAndDoc+":"),
			extPrefTreDoc+":")
	default:
		return uri.Path()
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
	tabFil = makeTabFil()
	tabCfg := makeTabCfg()
	tabLan := makeTabLan()
	tabLAf = makeTabLAf()
	tabs = container.NewAppTabs(
		tabLan,
		tabMsg,
		tabFil,
		tabLAf,
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
		case tabLAf:
			tabLAfShown()
		}
	}

	svrTcp.ActFunc = tcpConnAct
	svrTcp.ConnFunc = TcpSvrConnected
	svrTcp.LogFunc = Log

	ezcWin.SetFixedSize(true)
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
