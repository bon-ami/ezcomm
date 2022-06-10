package main

import (
	"time"

	"gitlab.com/bon-ami/ezcomm"
	"gitlab.com/bon-ami/ezcomm/guiFyne"

	"gitee.com/bon-ami/eztools/v4"
)

var (
	Ver, Bld string
)

func main() {
	if len(Ver) < 1 {
		Ver = "dev"
	}
	if len(Bld) < 1 {
		Bld = time.Now().Format("2006-01-02_15:04:05")
	}
	ezcomm.Ver = Ver
	ezcomm.Bld = Bld

	ezcomm.ReadCfg("")

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

	guiFyne.UI()
}
