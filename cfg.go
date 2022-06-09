package main

import (
	"os"

	"gitee.com/bon-ami/eztools/v4"
)

type ezcommFonts struct {
	Cmt    string `xml:",comment"`
	Locale string `xml:"locale,attr"`
	Font   string `xml:"font,attr"`
}
type ezcommCfg struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	//Txt      string `xml:",chardata"`
	Verbose  int
	LogFile  string
	Language string
	Fonts    []ezcommFonts
	font     string
}

var (
	cfgStruc ezcommCfg
	cfgPath  string
)

func writeCfg() error {
	var err error
	if len(cfgPath) > 0 {
		err = eztools.XMLWriteNoCreate(cfgPath, cfgStruc, "\t")
	} else {
		cfgPath, err = eztools.XMLWriteDefault(ezcName, cfgStruc, "\t")
	}
	if eztools.Debugging && eztools.Verbose > 1 {
		eztools.LogWtTime("writing config", cfgPath,
			cfgStruc, "with error", err)
	}
	if err != nil {
		return err
	}
	return nil
}

func readCfg(paramLogI string) {
	paramLogO := paramLogI
	var err error
	cfgPath, err = eztools.XMLReadDefault("", ezcName, &cfgStruc)
	if err == nil {
		if len(cfgStruc.LogFile) > 0 {
			if len(paramLogI) < 1 {
				paramLogO = cfgStruc.LogFile
			}
		}
		if cfgStruc.Verbose > 0 {
			if eztools.Verbose < cfgStruc.Verbose {
				eztools.Verbose = cfgStruc.Verbose
			}
		}
	}
	if eztools.Verbose > 0 {
		eztools.Debugging = true
	}
	if eztools.Debugging {
		if len(paramLogO) < 1 {
			paramLogO = ezcName + ".log"
		}
		//if eztools.Verbose > 1 {
		eztools.LogPrint("verbose", eztools.Verbose, ",log file =", paramLogO)
		//}
	}
	i18nInit()
	i18nLoad(cfgStruc.Language)
	matchFontFromLanguage()

	if len(paramLogO) > 0 {
		setLog(paramLogO)
	}
	return
}

func matchFontFromLanguage() {
	if len(cfgStruc.Language) < 1 {
		return
	}

	for _, font1 := range cfgStruc.Fonts {
		if font1.Locale == cfgStruc.Language {
			cfgStruc.font = font1.Font
			//eztools.Log("font formerly set", font1.Font)
			return
		}
	}
	cfgStruc.font = ""
	return
}

func setLog(fil string) error {
	logger, err := os.OpenFile(fil,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		if err = eztools.InitLogger(logger); err != nil {
			eztools.LogPrint(err)
		}
	} else {
		eztools.LogPrint("Failed to open log file "+fil, err)
	}
	return err
}
