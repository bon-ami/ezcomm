package main

import (
	"gitlab.com/bon-ami/ezcomm"
	"gitlab.com/bon-ami/ezcomm/guiFyne"

	"gitee.com/bon-ami/eztools/v4"
)

var (
	// Ver & Bld may or may not be useful
	Ver, Bld string
)

type Guis interface {
	// GuiSetGlbPrm is run at the beginning,
	//   to initialize following for ezcomm.
	//   Ver, Bld, GuiConnected, GuiEnded, GuiLog, GuiRcv, GuiSnt
	GuiSetGlbPrm(Ver, Bld string)
	// GuiRun is run at the end to handle UI in the main thread
	GuiRun()
}

func main() {
	var gui guiFyne.GuiFyne
	gui.GuiSetGlbPrm(Ver, Bld)

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

	gui.GuiRun()
}
