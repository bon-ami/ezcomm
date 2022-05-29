package main

import (
	"encoding/xml"
	"flag"
	"os"

	"gitee.com/bon-ami/eztools/v4"
)

const ezcName = "EZComm"

var (
	// Ver & Bld are stored in toml in Fyne
	Ver, Bld string
)

const (
	StrInt = "interative"
	StrCfg = "config"
	StrLcl = "local"
	StrRmt = "remote"
	StrLst = "listen"
	StrDis = "disconnect peer"
	StrStp = "stop listening"
	StrCon = "connect"
	StrAdr = "address"
	StrPrt = "port"
	StrRec = "content records"
	StrCnt = "content"
	StrSnd = "send"
	StrTo  = "send to"
	StrFrm = "received from"
	DefAdr = "localhost:"
	StrFlw = "flow"
	StrAll = "select all"
	StrUdp = "udp"
	StrTcp = "tcp"
)

type Cfg struct {
	// Root of the XML
	Root xml.Name `xml:"ezcommCfg"`
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	Txt      string `xml:",chardata"`
	Verbose  int
	LogFile  string
	Language string
}

var cfg Cfg

func main() {
	var (
		paramLog, paramFlw string
		// paramVer bool
		paramH, paramV, paramVV, paramVVV bool
	)
	// version info is tracked in toml by Fyne
	/*flag.BoolVar(&paramVer, "version", false, "version info")
	flag.BoolVar(&paramVer, "ver", false, "version info")*/
	flag.BoolVar(&paramH, "h", false, "help info")
	flag.BoolVar(&paramH, "help", false, "help info")
	flag.BoolVar(&paramV, "v", false, "log connection events")
	flag.BoolVar(&paramVV, "vv", false, "also log connetion termination details")
	flag.BoolVar(&paramVVV, "vvv", false, "also log I/O details")
	flag.StringVar(&paramLog, "log", "", "log file name")
	flag.StringVar(&paramFlw, "flow", "", "input file name to control flow/interactions.")
	flag.Parse()
	/*if len(Ver) < 1 {
			Ver = "dev"
		}
		if len(Bld) < 1 {
			Bld = time.Now().Format("2006-01-02_15:04:05")
		}
	if paramVer {
			eztools.ShowStrln("version " + Ver + " build " + Bld)
			return
		}*/
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

	_, err := eztools.XMLReadDefault("", ezcName, &cfg)
	if err == nil {
		if len(cfg.LogFile) > 0 {
			if len(paramLog) < 1 {
				paramLog = cfg.LogFile
			}
		}
		if cfg.Verbose > 0 {
			if eztools.Verbose < cfg.Verbose {
				eztools.Verbose = cfg.Verbose
			}
		}
		if len(cfg.Language) > 0 {

		}
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
