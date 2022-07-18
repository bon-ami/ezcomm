package ezcomm

import (
	"io"
	"os"
	"runtime"

	"gitee.com/bon-ami/eztools/v4"
)

const (
	EzcName      = "EZComm"
	DefAdr       = "localhost"
	StrUdp       = "udp"
	StrTcp       = "tcp"
	DefAntFldLmt = 10
	DefAntFldPrd = 60
)

var (
	Ver, Bld string
)

type EzcommFonts struct {
	// Cmt is not used
	Cmt string `xml:",comment"`
	// Locale is like zh-TW
	Locale string `xml:"locale,attr"`
	// Font is built-in names for fonts, which are locales,
	// or paths of fonts
	Font string `xml:"font,attr"`
}
type ezcommCfg struct {
	// Cmt = comments
	Cmt string `xml:",comment"`
	// Txt is not used
	//Txt      string `xml:",chardata"`
	// Verbose is level 0-3
	Verbose int
	// LogFile is path to log file, if enabled
	LogFile string
	// Language is locale
	Language string
	// Fonts are configurations of fonts
	Fonts []EzcommFonts
	// font is current font path, or built-in fonts, which are locales
	font string
	// AntiFlood stores anti-flood config
	AntiFlood struct {
		Limit, Period int64
	}
}

func (c ezcommCfg) GetFont() string {
	return c.font
}

var (
	CfgStruc  ezcommCfg
	cfgPath   string
	LogWtTime bool
)

func WriteCfg() error {
	err := eztools.ErrNoValidResults
	if len(cfgPath) > 0 {
		err = eztools.XMLWrite(cfgPath, CfgStruc, "\t")
	}
	if err != nil {
		Log("failed to write config", cfgPath, err)
		cfgPath, err = eztools.XMLWriteDefault(EzcName, CfgStruc, "\t")
	}
	resWriteCfg(err)
	if err != nil {
		cfgPath = ""
		return err
	}
	return err
}

func resWriteCfg(err error) {
	Log("writing config", cfgPath,
		CfgStruc, "with (no?) error", err)
}

func WriterCfg(wrt io.WriteCloser) error {
	err := eztools.XMLWriter(wrt, CfgStruc, "\t")
	wrt.Close()
	resWriteCfg(err)
	return err
}

func ReaderCfg(rdr io.ReadCloser, paramLogI string) error {
	if rdr != nil {
		setDefCfg()
		eztools.XMLReader(rdr, &CfgStruc)
		rdr.Close()
	}
	return procCfg(paramLogI)
}

func ReadCfg(cfg, paramLogI string) error {
	setDefCfg()
	cfgPath, _ = eztools.XMLReadDefault(cfg, EzcName, &CfgStruc)
	if len(cfgPath) < 1 && len(cfg) > 0 {
		// not exist yet?
		cfgPath = cfg
	}
	return procCfg(paramLogI)
}

func setDefCfg() {
	CfgStruc.AntiFlood.Limit = DefAntFldLmt
	CfgStruc.AntiFlood.Period = DefAntFldPrd
}

func procCfg(paramLogI string) error {
	paramLogO := paramLogI
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
	if eztools.Verbose > 0 {
		eztools.Debugging = true
	}
	if eztools.Debugging {
		if len(paramLogO) < 1 {
			switch runtime.GOOS {
			case "android":
				// logcat
				break
			default:
				paramLogO = EzcName + ".log"
			}
		}
		Log("verbose", eztools.Verbose, ",log file =", paramLogO)
	}
	if len(paramLogO) > 0 {
		setLog(paramLogO)
		LogWtTime = true
	}

	I18nInit()
	var (
		err  error
		lang string
	)
	if len(CfgStruc.Language) < 1 {
		// to avoid no fonts for current env, though I18nLoad can get system language from empty input
		lang, err = I18nLoad("en")
	} else {
		lang, err = I18nLoad(CfgStruc.Language)
	}
	if err == nil {
		CfgStruc.Language = lang
		MatchFontFromCurrLanguageCfg()
	}
	if CfgStruc.Verbose > 2 {
		Log(CfgStruc)
	}
	return err
}

func MatchFontFromCurrLanguageCfg() {
	if len(CfgStruc.Language) < 1 {
		return
	}

	if CfgStruc.Fonts != nil {
		for _, font1 := range CfgStruc.Fonts {
			if font1.Locale == CfgStruc.Language {
				CfgStruc.font = font1.Font
				//Log("font formerly set", font1.Font)
				return
			}
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
			Log(err)
		}
	} else {
		Log("Failed to open log file "+fil, err)
	}
	return err
}
