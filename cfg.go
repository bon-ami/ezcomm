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

func (c ezcommCfg) SetFont(f string) {
	c.font = f
}

func (c ezcommCfg) GetFont() string {
	return c.font
}

var (
	CfgStruc ezcommCfg
	cfgPath  string
)

func WriteCfg() error {
	err := eztools.ErrNoValidResults
	if len(cfgPath) > 0 {
		err = eztools.XMLWrite(cfgPath, CfgStruc, "\t")
	}
	if err != nil {
		//log("failed to write config", cfgPath, err)
		cfgPath, err = eztools.XMLWriteDefault(EzcName, CfgStruc, "\t")
	}
	//resWriteCfg(err)
	if err != nil {
		cfgPath = ""
		return err
	}
	return err
}

/*func resWriteCfg(err error) {
	log("writing config", cfgPath,
		CfgStruc, "with (no?) error", err)
}*/

func WriterCfg(wrt io.WriteCloser) error {
	err := eztools.XMLWriter(wrt, CfgStruc, "\t")
	wrt.Close()
	//resWriteCfg(err)
	return err
}

// ReaderCfg read config from a Reader and use a file as log
// Parameters: paramLogI overrides log config from Reader
// Return values: log file path and error
func ReaderCfg(rdr io.ReadCloser, paramLogI string) (string, error) {
	if rdr != nil {
		setDefCfg()
		eztools.XMLReader(rdr, &CfgStruc)
		rdr.Close()
	}
	return procCfg(paramLogI)
}

// ReadCfg read config from a file and use a file as log
// Parameters: paramLogI overrides log config from Reader
// Return values: log file path and error
func ReadCfg(cfg, paramLogI string) (string, error) {
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

func procCfg(paramLogI string) (string, error) {
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
		//log("verbose", eztools.Verbose, ",log file =", paramLogO)
	}

	// anti-flood
	AntiFlood.Limit = CfgStruc.AntiFlood.Limit
	AntiFlood.Period = CfgStruc.AntiFlood.Period

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
	/*if CfgStruc.Verbose > 2 {
		log(CfgStruc)
	}*/
	return paramLogO, err
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

// SetLog sets a writer or file as log
// date and time will be prefixed to the file;
// none to a writer, otherwise.
func SetLog(fil string, wr io.Writer) (err error) {
	if wr == nil {
		wr, err = os.OpenFile(fil,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	}
	if err != nil {
		return err
	}
	if err = eztools.InitLogger(wr); err != nil {
		return err
	}
	if len(fil) > 0 {
		flags := eztools.LogFlagDateNTime
		/*if eztools.Verbose > 2 {
			flags = eztools.LogFlagDateTimeNFile
		}*/
		eztools.SetLogFlags(flags)
	} else {
		eztools.SetLogFlags(0)
	}
	return nil
}
