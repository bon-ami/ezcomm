package ezcomm

import (
	"os"

	"gitee.com/bon-ami/eztools/v4"
)

const (
	EzcName = "EZComm"
	DefAdr  = "localhost:"
	StrUdp  = "udp"
	StrTcp  = "tcp"
)

var (
	Ver, Bld string
)

type EzcommFonts struct {
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
	Fonts    []EzcommFonts
	font     string
}

func (c ezcommCfg) GetFont() string {
	return c.font
}

var (
	CfgStruc ezcommCfg
	cfgPath  string
)

func WriteCfg() error {
	var err error
	if len(cfgPath) > 0 {
		err = eztools.XMLWriteNoCreate(cfgPath, CfgStruc, "\t")
	} else {
		cfgPath, err = eztools.XMLWriteDefault(EzcName, CfgStruc, "\t")
	}
	if eztools.Debugging && eztools.Verbose > 1 {
		eztools.LogWtTime("writing config", cfgPath,
			CfgStruc, "with error", err)
	}
	if err != nil {
		return err
	}
	return nil
}

func ReadCfg(paramLogI string) {
	paramLogO := paramLogI
	var err error
	cfgPath, err = eztools.XMLReadDefault("", EzcName, &CfgStruc)
	if err == nil {
		if len(CfgStruc.LogFile) > 0 {
			if len(paramLogI) < 1 {
				paramLogO = CfgStruc.LogFile
			}
		}
		if CfgStruc.Verbose > 0 {
			if eztools.Verbose < CfgStruc.Verbose {
				eztools.Verbose = CfgStruc.Verbose
			}
		}
	}
	if eztools.Verbose > 0 {
		eztools.Debugging = true
	}
	if eztools.Debugging {
		if len(paramLogO) < 1 {
			paramLogO = EzcName + ".log"
		}
		//if eztools.Verbose > 1 {
		eztools.LogPrint("verbose", eztools.Verbose, ",log file =", paramLogO)
		//}
	}
	i18nInit()
	I18nLoad(CfgStruc.Language)
	MatchFontFromLanguage()

	if len(paramLogO) > 0 {
		setLog(paramLogO)
	}
	return
}

func MatchFontFromLanguage() {
	if len(CfgStruc.Language) < 1 {
		return
	}

	for _, font1 := range CfgStruc.Fonts {
		if font1.Locale == CfgStruc.Language {
			CfgStruc.font = font1.Font
			//eztools.Log("font formerly set", font1.Font)
			return
		}
	}
	CfgStruc.font = ""
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
