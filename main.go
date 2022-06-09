package main

import (
	"flag"
	"time"

	"gitee.com/bon-ami/eztools/v4"
)

const ezcName = "EZComm"

var (
	Ver, Bld string
)

const (
	DefAdr = "localhost:"
	StrUdp = "udp"
	StrTcp = "tcp"
)

func main() {
	var (
		paramLog, paramFlw                string
		paramVer                          bool
		paramH, paramV, paramVV, paramVVV bool
	)
	flag.BoolVar(&paramVer, "version", false, "version info")
	flag.BoolVar(&paramVer, "ver", false, "version info")
	flag.BoolVar(&paramH, "h", false, "help info")
	flag.BoolVar(&paramH, "help", false, "help info")
	flag.BoolVar(&paramV, "v", false, "log connection events")
	flag.BoolVar(&paramVV, "vv", false, "also log connetion termination details")
	flag.BoolVar(&paramVVV, "vvv", false, "also log I/O details")
	flag.StringVar(&paramLog, "log", "", "log file name")
	flag.StringVar(&paramFlw, "flow", "", "input file name to control flow/interactions.")
	flag.Parse()
	if len(Ver) < 1 {
		Ver = "dev"
	}
	if len(Bld) < 1 {
		Bld = time.Now().Format("2006-01-02_15:04:05")
	}
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

	readCfg(paramLog)

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
		go db.AppUpgrade("", ezcName, Ver, nil, upch)
	}
	// db ends

	if len(paramFlw) < 1 || !runFlowFile(paramFlw) {
		guiFyne()
	}

}
