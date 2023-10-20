package main

import (
	"flag"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	ezcApp     fyne.App
	ezcWin     fyne.Window
	thm        theme4Fonts
	appStorage fyne.Storage
	// chn is for TCP client and UDP
	tabs                           *container.AppTabs
	tabLan, tabWeb, tabFil, tabLAf *container.TabItem
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
	dldURI = storage.NewFileURI(dldDirPath)
	exi, err := storage.Exists(dldURI)
	if err != nil {
		eztools.Log("NO", dldDirPath, "detectable!", err)
		return "", err
	}
	if exi {
		cn, err := storage.CanList(dldURI)
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
		if err = storage.CreateListable(dldURI); err != nil {
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

func validateInt64(str string) error {
	if _, err := strconv.ParseInt(str, 10, 64); err != nil {
		return err
	}
	return nil
}

// parseSck splits host:port
func parseSck(addr string) (string, string) {
	ip, port, err := net.SplitHostPort(addr)
	//Log("parsing", addr, "to", ip, port)
	if err != nil {
		Log(err)
		return addr, ""
	}
	return ip, port
}

func cpFile(rd io.ReadCloser, wr io.WriteCloser) (err error) {
	defer rd.Close()
	defer wr.Close()
	_, err = io.Copy(wr, rd)
	return
}

var (
	toastLock sync.Mutex
	toastOn   bool
)

// toast shows a toast window
// Parameters: index to ezcomm.StringTran and second line string
func toast(id, inf string) {
	const toNorm = time.Second * 3
	const toFast = time.Second * 1
	var to time.Duration
	if toastOn {
		to = toFast
	} else {
		to = toNorm
	}
	toastLock.Lock()
	toastOn = true
	go func(inf string, to time.Duration) {
		if drv, ok := ezcApp.Driver().(desktop.Driver); ok {
			w := drv.CreateSplashWindow()
			w.SetContent(widget.NewLabel(
				ezcomm.StringTran[id] + "\n" + inf))
			w.Show()
			go func(inf string, to time.Duration) {
				time.Sleep(to)
				w.Close()
				toastOn = false
				toastLock.Unlock()
			}(inf, to)
		}
	}(inf, to)
}

func cp2Clip(str string) {
	drv := ezcApp.Driver()
	clipboard := drv.AllWindows()[0].Clipboard()
	clipboard.SetContent(str)
	go toast("StrCopied", str)
}

func main() {
	ezcApp = app.NewWithID(ezcomm.EzcName)
	run(nil)
}

// run the main loop of UI
func run(chnHTTP chan bool) {
	parseParams()
	appStorage = ezcApp.Storage()
	cfgFileName := ezcomm.EzcName + ".xml"
	rdr, err := appStorage.Open(cfgFileName)
	if err != nil {
		if eztools.Verbose > 1 {
			eztools.Log("failed to open config file", cfgFileName, ":", err)
		}
	}
	// to avoid no fonts for current env, use en regardless of system language
	err = ezcomm.ReaderCfg(rdr, "en")
	if err != nil {
		if eztools.Verbose > 1 {
			eztools.Log("config/locale failure:", err)
		}
	}
	if eztools.Debugging {
		err = initLog()
		if err != nil {
			if eztools.Verbose > 1 {
				eztools.Log("failed to set log:", err)
			}
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

	makeControlsSocks()
	tabLog := makeTabLog()
	tabMsg := makeTabMsg()
	tabFil = makeTabFil()
	tabCfg := makeTabCfg()
	if chnHTTP == nil {
		chnHTTP = make(chan bool, 1)
		defer close(chnHTTP)
	}
	chnLan = make(chan bool, 1)
	defer close(chnLan)
	tabLan = makeTabLan(chnHTTP)
	tabWeb = makeTabWeb()
	tabLAf = makeTabLAf()
	tabs = container.NewAppTabs(
		tabLan,
		tabWeb,
		tabMsg,
		tabFil,
		tabLAf,
		tabLog,
		tabCfg,
	)
	var currTabLan bool
	tabs.OnSelected = func(tb *container.TabItem) {
		if currTabLan {
			switch tb {
			case tabLan, tabWeb:
				break
			default:
				currTabLan = false
				tabLanShown(currTabLan)
			}
		}
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
			if !currTabLan { // web->lan
				currTabLan = true
				tabLanShown(currTabLan)
			}
		case tabWeb:
			tabWebShown()
		case tabLAf:
			tabLAfShown()
		}
	}
	ezcWin.SetContent(tabs)

	//ezcWin.SetFixedSize(true)
	ezcWin.Show()
	currTabLan = true
	tabLanShown(true)
	// 9 routines here
	ezcApp.Run()
	tabLanShown(false)
	if eztools.Debugging && runtime.NumGoroutine() > 7 {
		buf := make([]byte, 16384)
		stackSize := runtime.Stack(buf, true)
		eztools.Log(buf[:stackSize])
	}
	if logger != nil {
		logger.Close()
	}
}
