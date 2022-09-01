package main

import (
	"flag"
	"time"

	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	Ver, Bld string
)

func main() {
	var (
		paramCfg, paramLog, paramFlw      string
		paramVer                          bool
		paramH, paramV, paramVV, paramVVV bool
	)
	// read current OS language
	ezcomm.I18nInit()
	_, err := ezcomm.I18nLoad("")
	flag.BoolVar(&paramVer, "version", false, ezcomm.StringTran["StrVer"])
	flag.BoolVar(&paramVer, "ver", false, ezcomm.StringTran["StrVer"])
	flag.BoolVar(&paramH, "h", false, ezcomm.StringTran["StrHlp"])
	flag.BoolVar(&paramH, "help", false, ezcomm.StringTran["StrHlp"])
	flag.BoolVar(&paramV, "v", false, ezcomm.StringTran["StrV"])
	flag.BoolVar(&paramVV, "vv", false, ezcomm.StringTran["StrVV"])
	flag.BoolVar(&paramVVV, "vvv", false, ezcomm.StringTran["StrVVV"])
	flag.StringVar(&paramLog, "log", "", ezcomm.StringTran["StrLogFn"])
	flag.StringVar(&paramCfg, "cfg", "", ezcomm.StringTran["StrCfg"])
	flag.StringVar(&paramFlw, "flow", "", ezcomm.StringTran["StrFlowFnInf"])
	flag.Parse()
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

	paramLog, err = ezcomm.ReadCfg(paramCfg, paramLog)
	if err == nil {
		ezcomm.SetLog(paramLog, nil)
	}

	if paramH {
		flag.Usage()
		return
	}

	// db is only for app upgrade
	db, _, err := eztools.MakeDbs()

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
		go db.AppUpgrade("", ezcomm.EzcName, Ver, nil, upch)
	}
	// db ends

	if len(paramFlw) > 0 {
		if flow, err := ezcomm.ReadFlowFile(paramFlw); err != nil {
			eztools.LogPrint(err)
		} else {
			ezcomm.RunFlow(flow)
		}
	}
}
