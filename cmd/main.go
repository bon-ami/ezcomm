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
		paramLog, paramFlw                string
		paramVer                          bool
		paramH, paramV, paramVV, paramVVV bool
	)
	flag.BoolVar(&paramVer, "version", false, ezcomm.StrVer)
	flag.BoolVar(&paramVer, "ver", false, ezcomm.StrVer)
	flag.BoolVar(&paramH, "h", false, ezcomm.StrHlp)
	flag.BoolVar(&paramH, "help", false, ezcomm.StrHlp)
	flag.BoolVar(&paramV, "v", false, ezcomm.StrV)
	flag.BoolVar(&paramVV, "vv", false, ezcomm.StrVV)
	flag.BoolVar(&paramVVV, "vvv", false, ezcomm.StrVVV)
	flag.StringVar(&paramLog, "log", "", ezcomm.StrLogFn)
	flag.StringVar(&paramFlw, "flow", "", ezcomm.StrFlowFnInf)
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

	ezcomm.ReadCfg(paramLog)

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
	defer func() {
		if db != nil {
			if !(<-upch) {
				if eztools.Debugging {
					eztools.LogPrint("wrong server for update check")
				}
			} else {
				if !(<-upch) {
					if eztools.Debugging {
						eztools.LogPrint("update check failed")
					}
				} else {
					if eztools.Debugging {
						eztools.LogPrint("update check done/skipped")
					}
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
		ezcomm.RunFlowFile(paramFlw)
	}
}
