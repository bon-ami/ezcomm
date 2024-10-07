package ezcomm

import (
	"io"
	"os"

	"gitee.com/bon-ami/eztools/v6"
)

const (
	// EzcName app name, such as for log file
	EzcName = "EZComm"
	// EzcURL URL of homepage
	EzcURL = "https://gitlab.com/bon-ami/ezcomm"
	// LogExt log file extension
	LogExt = ".log"
	// Localhost default peer address
	Localhost = "localhost" // use "" instead to listen on all interfaces
	// DefBrdAdr address for broadcast in lan discovery
	DefBrdAdr = "255.255.255.255" // broadcast addr
	// StrUDP is UDP protocol. otherwise, udp4, udp6, etc.
	StrUDP = "udp"
	// StrTCP is TCP protocol. otherwise, tcp4, tcp6, etc.
	StrTCP = "tcp"
	// DefAntFldLmt default anti-flood limit to tolerate a peer
	DefAntFldLmt = 10
	// DefAntFldPrd default anti-flood period to block a peer
	DefAntFldPrd = 60
)

var (
	// Ver = verison
	Ver string
	// Bld = build, a number or date
	Bld string
	// Vendor & AdditionalCfgPath are to locate config file {EzcName}.xml under
	// {system-app-local-dir}/{Vendor}/{EzcName}/{AdditionalCfgPath}
	// it may vary among GUI's
	Vendor string
	// AdditionalCfgPath is to locate config file. Refer to Vendor
	AdditionalCfgPath string
)

// Fonts is a font-locale match
type Fonts struct {
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
	// Language is locale
	Language string
	// Fonts are configurations of fonts
	Fonts []Fonts
	// font is current font path, or built-in fonts, which are locales
	font string
	// AntiFlood stores anti-flood config
	AntiFlood struct {
		Limit, Period int64
	}
}

func (c ezcommCfg) EzcName() string {
	if ret, ok := StringTran["StrEzcName"]; ok {
		return ret
	}
	return EzcName
}

func (c ezcommCfg) EzpName() string {
	if ret, ok := StringTran["StrEzpName"]; ok {
		return ret
	}
	return eztools.EzpName
}

func (c *ezcommCfg) SetFont(f string) {
	c.font = f
}

func (c ezcommCfg) GetFont() string {
	return c.font
}

var (
	// CfgStruc config structure
	CfgStruc ezcommCfg
	cfgPath  string
)

// WriteCfg to save config
func WriteCfg() error {
	err := eztools.ErrNoValidResults
	if len(cfgPath) > 0 {
		err = eztools.XMLWrite(cfgPath, CfgStruc, "\t")
	}
	if err == nil {
		return err
	}
	//log("failed to write config", cfgPath, err)
	var errs []error
	cfgPath, errs = eztools.XMLWriteDefault(Vendor,
		"", AdditionalCfgPath, EzcName, CfgStruc, "\t")
	if errs != nil {
		cfgPath = ""
		return errs[0]
	}
	return nil
}

// WriterCfg to save config using Writer
// Closer is closed before returning
func WriterCfg(wrt io.WriteCloser) error {
	err := eztools.XMLWriter(wrt, CfgStruc, "\t")
	wrt.Close()
	return err
}

// ReaderCfg reads config from a Reader
// Closer is closed before returning
func ReaderCfg(rdr io.ReadCloser, fallbackLang string) error {
	if rdr != nil {
		setDefCfg()
		if err := eztools.XMLReader(rdr, &CfgStruc); err != nil {
			if eztools.Debugging && eztools.Verbose > 1 {
				eztools.Log("failed to read config", err)
			}
		}
		rdr.Close()
	}
	return procCfg(fallbackLang)
}

// ReadCfg reads config from a file
func ReadCfg(cfg string, fallbackLang string) error {
	setDefCfg()
	var err []error
	cfgPath, err = eztools.XMLReadDefault(cfg, Vendor, "",
		AdditionalCfgPath, EzcName, &CfgStruc)
	if err != nil || len(cfgPath) < 1 && len(cfg) > 0 {
		// not exist yet?
		cfgPath = cfg
	}
	return procCfg(fallbackLang)
}

func setDefCfg() {
	CfgStruc.AntiFlood.Limit = DefAntFldLmt
	CfgStruc.AntiFlood.Period = DefAntFldPrd
}

func procCfg(fallbackLang string) error {
	if CfgStruc.Verbose > 0 {
		if eztools.Verbose < CfgStruc.Verbose {
			eztools.Verbose = CfgStruc.Verbose
		}
	}
	if eztools.Verbose > 0 {
		eztools.Debugging = true
	}
	//log("verbose", eztools.Verbose)

	// anti-flood
	AntiFlood.Limit = CfgStruc.AntiFlood.Limit
	AntiFlood.Period = CfgStruc.AntiFlood.Period

	I18nInit()
	var (
		err  error
		lang string
	)
	if len(CfgStruc.Language) < 1 {
		lang, err = I18nLoad(fallbackLang)
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
	return err
}

// MatchFontFromCurrLanguageCfg matches font for current language in config
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
	flags := eztools.LogFlagDateNTime
	eztools.SetLogFlags(flags)
	return nil
}

// DefLanPrt returns default port for lan discovery
func DefLanPrt() (ret int) {
	for _, c := range EzcName {
		ret += int(c)
	}
	return ret * 10 //==5550
}
