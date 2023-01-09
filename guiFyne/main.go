package main

import (
	"flag"
	"io"
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
	ezcApp     fyne.App
	ezcWin     fyne.Window
	thm        theme4Fonts
	appStorage fyne.Storage
	// chn is for TCP client and UDP
	chn [2]chan ezcomm.RoutCommStruc
	// svrTCP is for TCP server only
	svrTCP         ezcomm.SvrTCP
	tabs           *container.AppTabs
	tabFil, tabLAf *container.TabItem
)

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
	extPrefAndDoc   = "content://com.android.externalstorage.documents/document/primary"
	extPrefTreDoc   = "content://com.android.externalstorage.documents/tree/primary"
	sdcardPrefAnd   = "/sdcard"
	dldDirNm        = "Downloads"
	invalidFileName = "(invalid)"
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

// translateFilePath changes a path string for platforms
//
//	For Android, a path not beginning with "/" is prefixed with sdcardPrefAnd
func translateFilePath(p string) string {
	switch runtime.GOOS {
	case "android":
		/*change a/b into {extPref}%3Aa%2Fb meaning {extPref}:a/b
		p = extPrefAndDoc + url.QueryEscape(":"+p)
		return storage.ParseURI(p)*/
		if len(p) < 1 {
			return sdcardPrefAnd
		}
		if !strings.HasPrefix(p, "/") {
			return sdcardPrefAnd + "/" + p
		}
	}
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
	}
	return encodeFilePath(dldPath)
}

func decodeFilePath(uri fyne.URI) string {
	switch runtime.GOOS {
	case "android":
		if uri.Scheme() == "file" {
			break
		}
		fn, err := url.PathUnescape(uri.String())
		if err != nil {
			Log("failed to unescape:", uri)
			return ""
		}
		return translateFilePath(strings.TrimPrefix(
			strings.TrimPrefix(fn, extPrefAndDoc+":"),
			extPrefTreDoc+":"))
	}
	return uri.Path()
}

func main() {
	parseParams()
	ezcApp = app.NewWithID(ezcomm.EzcName)
	appStorage = ezcApp.Storage()
	cfgFileName := ezcomm.EzcName + ".xml"
	rdr, err := appStorage.Open(cfgFileName)
	if err != nil {
		eztools.Log("failed to open config file", cfgFileName, ":", err)
	}
	err = ezcomm.ReaderCfg(rdr)
	if err != nil {
		eztools.Log("config/locale failure:", err)
	}
	if eztools.Debugging {
		err = initLog()
		if err != nil {
			eztools.Log("failed to set log:", err)
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

	svrTCP.ActFunc = tcpConnAct
	svrTCP.ConnFunc = TcpSvrConnected
	svrTCP.LogFunc = Log

	//ezcWin.SetFixedSize(true)
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

func cpFile(rd io.ReadCloser, wr io.WriteCloser) (err error) {
	defer rd.Close()
	defer wr.Close()
	_, err = io.Copy(wr, rd)
	return
}
