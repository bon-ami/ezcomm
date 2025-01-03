package main

import (
	"flag"
	"io"
	"os"
	"time"

	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	// Ver version
	Ver string
	// Bld build number
	Bld string
)

func main() {
	var (
		paramCfg, paramLog, paramFlw      string
		paramVer                          bool
		paramH, paramV, paramVV, paramVVV bool
	)
	// read current OS language
	/*ezcomm.I18nInit()
	ezcomm.I18nLoad("")*/

	const paramCfgStr = "cfg"
	// try to get cfg param, for language settings
	var (
		argCfgGot bool
		argCfgStr string
	)
	for _, arg1 := range os.Args {
		if argCfgGot {
			argCfgStr = arg1
			break
		}
		if arg1 == "-"+paramCfgStr {
			argCfgGot = true
			continue
		}
	}
	// read config file or default config for default language
	// keep consistent with guiFyne
	ezcomm.Vendor = "fyne"
	ezcomm.AdditionalCfgPath = "Documents"
	if err := ezcomm.ReadCfg(argCfgStr, ""); err != nil {
		if eztools.Verbose > 0 {
			eztools.Log("failed to load", argCfgStr, err)
		}
	}
	// all possible language settings got. parse all params again.

	flag.BoolVar(&paramVer, "version", false, ezcomm.StringTran["StrVer"])
	flag.BoolVar(&paramVer, "ver", false, ezcomm.StringTran["StrVer"])
	flag.BoolVar(&paramH, "h", false, ezcomm.StringTran["StrHlp"])
	flag.BoolVar(&paramH, "help", false, ezcomm.StringTran["StrHlp"])
	flag.BoolVar(&paramV, "v", false, ezcomm.StringTran["StrV"])
	flag.BoolVar(&paramVV, "vv", false, ezcomm.StringTran["StrVV"])
	flag.BoolVar(&paramVVV, "vvv", false, ezcomm.StringTran["StrVVV"])
	flag.StringVar(&paramLog, "log", "", ezcomm.StringTran["StrLogFn"])
	flag.StringVar(&paramCfg, paramCfgStr, "", ezcomm.StringTran["StrCfg"])
	flag.StringVar(&paramFlw, "flow", "", ezcomm.StringTran["StrFlowFnInf"])
	flag.Parse()

	if paramH {
		flag.Usage()
		return
	}

	if len(Ver) < 1 {
		Ver = "dev"
	}
	if len(Bld) < 1 {
		Bld = time.Now().Format("2006-01-02_15:04:05")
	}
	ezcomm.Ver = Ver
	ezcomm.Bld = Bld
	if paramVer {
		eztools.ShowStrln("version " + Ver + " build " + Bld)
		return
	}
	switch {
	case paramV:
		eztools.Verbose = 1
	case paramVV:
		eztools.Verbose = 2
	case paramVVV:
		eztools.Verbose = 3
	}

	if err := ezcomm.ReadCfg(paramCfg, ""); err != nil {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.Log("failed to load config", err)
		}
	}
	if paramLog == "" {
		paramLog = ezcomm.EzcName + ezcomm.LogExt
	}
	if err := ezcomm.SetLog(paramLog, nil); err != nil {
		if eztools.Debugging && eztools.Verbose > 0 {
			eztools.Log("failed to set log", err)
		}
	}

	// db is only for app upgrade
	db, _, err := eztools.MakeDb()

	if err != nil {
		if eztools.Debugging {
			eztools.ShowStrln("config file NOT found for eztools")
			eztools.Log(err)
		}
		// minor error, not affecting exit value
	}
	var upch chan bool
	log := func(sth ...any) {
		if eztools.Debugging {
			eztools.Log(sth)
		}
	}
	defer func() {
		if db != nil {
			if !(<-upch) {
				log("wrong server for update check")
			} else {
				if !(<-upch) {
					log("update check failed")
				} else {
					log("update check done/skipped")
				}
			}
			db.Close()
		}
		/*switch err {
		case nil:
			return
		default:
			os.Exit(1)
		}*/
	}()
	if db != nil {
		upch = make(chan bool, 2)
		defer close(upch)
		go db.AppUpgrade("", ezcomm.EzcName, Ver, nil, upch)
	}
	// db ends

	if len(paramFlw) > 0 {
		ezcomm.FlowReaderNew = func(p string) (io.ReadCloser, error) {
			return os.Open(p)
		}
		ezcomm.FlowWriterNew = func(p string) (io.WriteCloser, error) {
			return os.Create(p)
		}
		if flow, err := ezcomm.ReadFlowFile(paramFlw); err != nil {
			eztools.LogPrint(err)
		} else {
			ezcomm.RunFlow(flow)
		}
	}
}
