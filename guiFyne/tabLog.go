package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v5"
	"gitlab.com/bon-ami/ezcomm"
)

var (
	fyneRowLog *Entry
	logger     fyne.URIWriteCloser
	logBut     *widget.Button
)

func initLog() error {
	nm := ezcomm.EzcName + ezcomm.LogExt
	// backup if in existence
	rd, err := appStorage.Open(nm)
	var errMv error
	if err == nil {
		fn := ezcomm.EzcName + ezcomm.CurrTime() + ezcomm.LogExt
		wr, err := appStorage.Create(fn)
		errMv = err
		if err == nil {
			cpFile(rd, wr)
			/*if eztools.Debugging && eztools.Verbose > 1 {
				eztools.LogPrint(
					"copied", nm, "to", fn, "err", err)
			}*/
		} else {
			/*if eztools.Debugging && eztools.Verbose > 1 {
				eztools.LogPrint("err creating", fn, err)
			}*/
			rd.Close()
		}
		// rd and wr are closed
		/*} else {
		if eztools.Debugging && eztools.Verbose > 1 {
			eztools.LogPrint("no", nm)
		}*/
	}
	defer func() {
		if errMv != nil {
			eztools.Log(nm, "NOT backuped:", errMv)
		}
	}()
	var wr fyne.URIWriteCloser
	if err != nil {
		/*if eztools.Debugging && eztools.Verbose > 1 {
			eztools.LogPrint("creating", nm)
		}*/
		wr, err = appStorage.Create(nm)
		if err != nil {
			return err
		}
	} else {
		/*if eztools.Debugging && eztools.Verbose > 1 {
			eztools.LogPrint("saving", nm)
		}*/
		wr, err = appStorage.Save(nm)
		if err != nil {
			return err
		}
	}
	logger = wr
	return ezcomm.SetLog("", wr)
}

var logLock sync.Mutex

func Log(inf ...any) {
	if fyneRowLog != nil {
		str := fmt.Sprintf("%s%v\n",
			time.Now().Format("01-02 15:04:05"), inf)
		logLock.Lock()
		fyneRowLog.SetText(fyneRowLog.Text + //strconv.Itoa(rowLog.CursorRow) + ":" +
			str)
		fyneRowLog.CursorRow++
		logLock.Unlock()
	}
	logLock.Lock()
	eztools.Log(inf...)
	logLock.Unlock()
}

func expLogs(wr fyne.URIWriteCloser) (string, error) {
	var files []fyne.URIReadCloser
	for _, f1 := range appStorage.List() {
		if strings.HasPrefix(f1, ezcomm.EzcName) &&
			strings.HasSuffix(f1, ezcomm.LogExt) {
			err := appStorage.Remove(f1)
			Log("removing", f1, err)
			rd, err := appStorage.Open(f1)
			if err != nil {
				Log("opening", f1, err)
				continue
			}
			files = append(files, rd)
		}
	}
	numFiles := len(files)
	if numFiles < 1 {
		return "", nil
	}

	i := -1
	errs := eztools.ZipWtReaderWriter(func() (string, io.ReadCloser) {
		i++
		if i >= numFiles {
			return "", nil
		}
		return files[i].URI().Name(), files[i]
	}, wr)
	var msg string
	for i, err := range errs {
		if err != nil {
			if i <= numFiles-1 {
				msg = files[i].URI().Name() + ": "
			} else {
				msg = wr.URI().Name() + ": "
			}
			return msg, err
		}
	}
	if len(errs) != numFiles+1 {
		return strconv.Itoa(len(errs)), eztools.ErrOutOfBound
	}
	var err error
	for _, f1 := range files {
		fn := f1.URI().Name()
		f1.Close()
		if fn == ezcomm.EzcName+ezcomm.LogExt {
			// keep current
			continue
		}
		if err1 := appStorage.Remove(fn); err1 != nil {
			msg += fn + ": "
			err = err1
		}
	}
	return msg, err
}

func makeTabLog() *container.TabItem {
	logBut = widget.NewButton(ezcomm.StringTran["StrLogExp"], func() {
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			if err != nil {
				Log("open log file:", err)
				dialog.ShowInformation(
					ezcomm.StringTran["StrAlert"],
					ezcomm.StringTran["StrNoPerm"], ezcWin)
				return
			}
			if uri == nil {
				return
			}
			msg, err := expLogs(uri)
			if err != nil {
				showNG(msg, err)
			} else {
				Log("logs exported to", uri.URI().String())
			}
		}, ezcWin)
	})
	rowVerbose := container.NewCenter(widget.NewLabel(ezcomm.StringTran["StrVbs"]))
	verboseSel := widget.NewSelect(nil, func(lvl string) {
		newLvl := verboseFrmStr(lvl)
		if newLvl == ezcomm.CfgStruc.Verbose {
			return
		}
		eztools.Verbose = newLvl
		ezcomm.CfgStruc.Verbose = newLvl
		writeCfg()
	})
	verboseSel.Options = []string{
		ezcomm.StringTran["StrHgh"], ezcomm.StringTran["StrMdm"], ezcomm.StringTran["StrLow"], ezcomm.StringTran["StrNon"],
	}
	verboseSel.SetSelected(verbose2Str())
	fyneRowLog = /*widget.*/ NewMultiLineEntry() //.NewList.NewTextGrid()
	fyneRowLog.Wrapping = fyne.TextWrapWord
	fyneRowLog.Disable()
	top := container.NewVBox(logBut, rowVerbose, verboseSel)
	return container.NewTabItem(ezcomm.StringTran["StrInfLog"],
		container.NewBorder(top, nil, nil, nil, fyneRowLog))
}

func tabLogShown() {
	logBut.Refresh()
}
