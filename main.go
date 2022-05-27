package main

import (
	"flag"
	"os"
	"time"

	"gitee.com/bon-ami/eztools/v4"
)

const ezcName = "EZComm"

var (
	Ver, Bld string
)

/*const (
	EZCOMM_TYPE_UDP = iota
	EZCOMM_TYPE_TCP
)

type EzComm struct {
	tp   int
	addr string
	port int
}
*/

const (
	STR_INT = "interative"
	STR_CFG = "config"
	STR_LCL = "local"
	STR_RMT = "remote"
	STR_LST = "listen"
	STR_DIS = "disconnect peer"
	STR_STP = "stop listening"
	STR_CON = "connect"
	STR_ADR = "address"
	STR_PRT = "port"
	STR_REC = "content records"
	STR_CNT = "content"
	STR_SND = "send"
	STR_TO  = "send to"
	STR_FRM = "received from"
	DEF_ADR = "localhost:"
	STR_FLW = "flow"
)

func main() {
	var (
		paramLog, paramFlw                          string
		paramH, paramVer, paramV, paramVV, paramVVV bool
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
	if paramH {
		flag.Usage()
		return
	}
	eztools.Debugging = paramV || paramVV || paramVVV
	switch {
	case paramV:
		eztools.Verbose = 1
	case paramVV:
		eztools.Verbose = 2
	case paramVVV:
		eztools.Verbose = 3
	}
	if eztools.Debugging {
		if len(paramLog) < 1 {
			paramLog = ezcName + ".log"
		}
	}
	if len(paramLog) > 0 {
		logger, err := os.OpenFile(paramLog,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			if err = eztools.InitLogger(logger); err != nil {
				eztools.LogPrint(err)
			}
		} else {
			eztools.LogPrint("Failed to open log file "+paramLog, err)
		}
	}
	db, _, err := eztools.MakeDbs()

	if err != nil {
		if eztools.Debugging {
			eztools.ShowStrln("config file NOT found for eztools")
			eztools.Log(err)
		}
		err = nil // minor error, not affecting exit value
	}
	defer func() {
		if db != nil {
			db.Close()
		}
		switch err {
		case nil:
			return
		default:
			os.Exit(1)
		}
	}()

	var upch chan bool
	if db != nil {
		upch = make(chan bool, 2)
		go db.AppUpgrade("", ezcName, Ver, nil, upch)
	}

	if len(paramFlw) < 1 || !runFlowFile(paramFlw) {
		guiFyne()
	}

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
	}
}
