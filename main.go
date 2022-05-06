package main

import (
	"flag"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

func main() {
	var (
		paramFlw         string
		paramH, paramVer bool
	)
	flag.BoolVar(&paramVer, "version", false, "version info")
	flag.BoolVar(&paramVer, "ver", false, "version info")
	flag.BoolVar(&paramH, "h", false, "help info")
	flag.BoolVar(&paramH, "help", false, "help info")
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
	db, _, err := eztools.MakeDbs()

	if err != nil {
		eztools.LogFatal(err)
	}
	defer func() {
		db.Close()
		switch err {
		case nil:
			return
		default:
			os.Exit(1)
		}
	}()

	upch := make(chan bool, 2)
	go db.AppUpgrade("", ezcName, Ver, nil, upch)

	if len(paramFlw) < 1 || !runFlowFile(paramFlw) {
		gui()
	}

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

const (
	STR_LCL = "local"
	STR_RMT = "remote"
	STR_LST = "listen"
	STR_ADR = "address"
	STR_PRT = "port"
	STR_REC = "content records"
	STR_CNT = "content"
	STR_SND = "send"
	STR_TO  = "send to"
	STR_FRM = "received from"
	DEF_ADR = "localhost:"
)

func gui() {
	ezcApp := app.New()
	/*icon, err := LoadResourceFromPath("icon.ico")
	if err == nil {
		ezcApp.SetIcon(icon)
	}*/
	ezcWin := ezcApp.NewWindow(ezcName)

	contLcl := guiMakeControlsLcl()
	contRmt := guiMakeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)
	//lay.Layout(prot)

	ezcWin.SetContent(cont) //prot)
	//ezcWin.SetContent(widget.NewLabel("Hello"))

	ezcWin.Show()
	ezcApp.Run()
}

func guiMakeControlsLcl() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(STR_LCL))

	addrLbl := widget.NewLabel(STR_ADR)
	portLbl := widget.NewLabel(STR_PRT)
	addrInp := widget.NewSelectEntry(nil)
	addrInp.PlaceHolder = DEF_ADR
	portInp := widget.NewSelectEntry(nil)
	rowSock := container.NewGridWithRows(2,
		addrLbl, addrInp, portLbl, portInp)

	prot := widget.NewRadioGroup([]string{"udp", "tcp"}, nil)
	prot.Horizontal = true
	prot.SetSelected("udp")
	lstBut := widget.NewButton(STR_LST, nil)
	rowProt := container.NewHBox(prot, lstBut)

	recLbl := container.NewCenter(widget.NewLabel(STR_REC))
	recInp := widget.NewSelectEntry(nil)
	cntLbl := container.NewCenter(widget.NewLabel(STR_CNT))
	rowRec := container.NewGridWithRows(3, recLbl, recInp, cntLbl)

	cntTxt := widget.NewMultiLineEntry()

	sndBut := widget.NewButton(STR_SND, nil)

	tops := container.NewVBox(rowLbl, rowSock, rowProt, rowRec)
	return container.NewBorder(tops, sndBut, nil, nil, cntTxt)
}

func guiMakeControlsRmt() *fyne.Container {
	rowLbl := container.NewCenter(widget.NewLabel(STR_RMT))
	rowTo := container.NewCenter(widget.NewLabel(STR_TO))

	addrInp := widget.NewSelectEntry(nil)
	portInp := widget.NewSelectEntry(nil)
	rowSock := container.NewGridWithColumns(2, addrInp, portInp)

	rowFrm := container.NewCenter(widget.NewLabel(STR_FRM))

	addr2Inp := widget.NewSelectEntry(nil)
	port2Inp := widget.NewSelectEntry(nil)
	row2Sock := container.NewGridWithColumns(2, addr2Inp, port2Inp)

	recLbl := container.NewCenter(widget.NewLabel(STR_REC))
	recInp := widget.NewSelectEntry(nil)
	cntLbl := container.NewCenter(widget.NewLabel(STR_CNT))
	rowRec := container.NewGridWithRows(3, recLbl, recInp, cntLbl)

	cntTxt := widget.NewMultiLineEntry()

	rowLog := widget.NewMultiLineEntry() //.NewList.NewTextGrid()
	rowLog.Disable()

	tops := container.NewVBox(rowLbl, rowTo, rowSock, rowFrm, row2Sock, rowRec)
	return container.NewBorder(tops, rowLog, nil, nil, cntTxt)
}
