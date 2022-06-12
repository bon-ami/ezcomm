package ezcomm

import (
	"path/filepath"
	"strings"

	"gitee.com/bon-ami/eztools/v4"
	"github.com/flopp/go-findfont"
	"gitlab.com/bon-ami/ezcomm/res"
)

var (
	StrInt           = "interative"
	StrCfg           = "config"
	StrLcl           = "local"
	StrRmt           = "remote"
	StrLst           = "listen"
	StrDis           = "disconnect peer"
	StrStp           = "stop listening"
	StrCon           = "connect"
	StrAdr           = "address"
	StrPrt           = "port"
	StrRec           = "content records"
	StrCnt           = "content"
	StrSnd           = "send"
	StrTo            = "send to"
	StrFrm           = "received from"
	StrFlw           = "flow"
	StrStb           = "standing by"
	StrAll           = "select all"
	StrLog           = "log file"
	StrLang          = "languages"
	StrVbs           = "verbose level"
	StrHgh           = "high"
	StrMdm           = "medium"
	StrLow           = "low"
	StrNon           = "none"
	StrFnt           = "fonts"
	StrFnt4Lang      = "set this font for this language"
	StrReboot4Change = "restart of this application will show the change"

	StrVer       = "version info"
	StrHlp       = "help info"
	StrV         = "log connection events"
	StrVV        = "also log connetion termination details"
	StrVVV       = "also log I/O details"
	StrLogFn     = "log file name"
	StrFlowFnInf = "input file name to control flow/interactions."

	StrFlowOpenErr   = "flow file opened ERR"
	StrFlowOpenNot   = "flow file NOT opened"
	StrFlowRunning   = "flow file running..."
	StrFlowRunNot    = "flow NOT run"
	StrOK            = "OK"
	StrNG            = "NG"
	StrFlowFinAs     = "flow file finished as "
	StrCut           = "Cut"
	StrCpy           = "Copy"
	StrPst           = "Paste"
	StrCpyAll        = "Copy all"
	StrClr           = "Clear"
	StrListeningOn   = "listening on"
	StrConnFail      = "fail to connect to "
	StrConnected     = "connected"
	StrDisconnected  = "disconnected"
	StrNotInRec      = "NOT in record!"
	StrSvrIdle       = "server idle"
	StrClntLft       = "clients left"
	StrNoPeer4       = "NO peer found for"
	StrDisconnecting = "disconnecting"
	StrStopLstn      = "stopped listening"
	StrFl2Rcv        = "failed to receive"
	StrGotFromSw     = "got from somewhere"
	StrGotFrom       = "got from"
	StrUnknownDsc    = "UNKNOWN peer disconnected"
	StrFl2Snd        = "failed to send"
	StrSw            = "somewhere"
	StrSnt2          = "sent to"
)

func i18nInit() {
	eztools.InitLanguages()
	res.I18n_en()
	res.I18n_es()
	res.I18n_zhCN()
}

func I18nLoad(lang string) {
	eztools.LoadLanguage(lang)
	StrLst = eztools.GetLanguageStr("StrLst")
	StrInt = eztools.GetLanguageStr("StrInt")
	StrCfg = eztools.GetLanguageStr("StrCfg")
	StrLcl = eztools.GetLanguageStr("StrLcl")
	StrRmt = eztools.GetLanguageStr("StrRmt")
	StrDis = eztools.GetLanguageStr("StrDis")
	StrStp = eztools.GetLanguageStr("StrStp")
	StrCon = eztools.GetLanguageStr("StrCon")
	StrAdr = eztools.GetLanguageStr("StrAdr")
	StrPrt = eztools.GetLanguageStr("StrPrt")
	StrRec = eztools.GetLanguageStr("StrRec")
	StrCnt = eztools.GetLanguageStr("StrCnt")
	StrSnd = eztools.GetLanguageStr("StrSnd")
	StrTo = eztools.GetLanguageStr("StrTo")
	StrFrm = eztools.GetLanguageStr("StrFrm")
	StrFlw = eztools.GetLanguageStr("StrFlw")
	StrStb = eztools.GetLanguageStr("StrStb")
	StrAll = eztools.GetLanguageStr("StrAll")
	StrLog = eztools.GetLanguageStr("StrLog")
	StrLang = eztools.GetLanguageStr("StrLang")
	StrVbs = eztools.GetLanguageStr("StrVbs")
	StrHgh = eztools.GetLanguageStr("StrHgh")
	StrMdm = eztools.GetLanguageStr("StrMdm")
	StrLow = eztools.GetLanguageStr("StrLow")
	StrNon = eztools.GetLanguageStr("StrNon")
	StrFnt = eztools.GetLanguageStr("StrFnt")
	StrFnt4Lang = eztools.GetLanguageStr("StrFnt4Lang")
	StrReboot4Change = eztools.GetLanguageStr("StrReboot4Change")
	StrVer = eztools.GetLanguageStr("StrVer")
	StrHlp = eztools.GetLanguageStr("StrHlp")
	StrV = eztools.GetLanguageStr("StrV")
	StrVV = eztools.GetLanguageStr("StrVV")
	StrVVV = eztools.GetLanguageStr("StrVVV")
	StrLogFn = eztools.GetLanguageStr("StrLogFn")
	StrFlowFnInf = eztools.GetLanguageStr("StrFlowFnInf")

	StrFlowOpenErr = eztools.GetLanguageStr("StrFlowOpenErr")
	StrFlowOpenNot = eztools.GetLanguageStr("StrFlowOpenNot")
	StrFlowRunning = eztools.GetLanguageStr("StrFlowRunning")
	StrFlowRunNot = eztools.GetLanguageStr("StrFlowRunNot")
	StrOK = eztools.GetLanguageStr("StrOK")
	StrNG = eztools.GetLanguageStr("StrNG")
	StrFlowFinAs = eztools.GetLanguageStr("StrFlowFinAs")
	StrCut = eztools.GetLanguageStr("StrCut")
	StrCpy = eztools.GetLanguageStr("StrCpy")
	StrPst = eztools.GetLanguageStr("StrPst")
	StrCpyAll = eztools.GetLanguageStr("StrCpyAll")
	StrClr = eztools.GetLanguageStr("StrClr")
	StrListeningOn = eztools.GetLanguageStr("StrListeningOn")
	StrConnFail = eztools.GetLanguageStr("StrConnFail")
	StrConnected = eztools.GetLanguageStr("StrConnected")
	StrDisconnected = eztools.GetLanguageStr("StrDisconnected")
	StrNotInRec = eztools.GetLanguageStr("StrNotInRec")
	StrSvrIdle = eztools.GetLanguageStr("StrSvrIdle")
	StrClntLft = eztools.GetLanguageStr("StrClntLft")
	StrNoPeer4 = eztools.GetLanguageStr("StrNoPeer4")
	StrDisconnecting = eztools.GetLanguageStr("StrDisconnecting")
	StrStopLstn = eztools.GetLanguageStr("StrStopLstn")
	StrFl2Rcv = eztools.GetLanguageStr("StrFl2Rcv")
	StrGotFromSw = eztools.GetLanguageStr("StrGotFromSw")
	StrGotFrom = eztools.GetLanguageStr("StrGotFrom")
	StrUnknownDsc = eztools.GetLanguageStr("StrUnknownDsc")
	StrFl2Snd = eztools.GetLanguageStr("StrFl2Snd")
	StrSw = eztools.GetLanguageStr("StrSw")
	StrSnt2 = eztools.GetLanguageStr("StrSnt2")
}

// fontList contains font paths and names
var (
	fontList [2][]string
	fontMap  map[string]int
)

func ListSystemFonts() []string {
	fontList[0] = findfont.List()
	if len(fontList[0]) < 1 {
		return nil
	}
	fontList[1] = make([]string, len(fontList[0]))
	fontMap = make(map[string]int)
	for i, v := range fontList[0] {
		fontFileName := filepath.Base(fontList[0][i])
		fontList[1][i] = strings.TrimSuffix(fontFileName, filepath.Ext(fontFileName))
		fontMap[v] = i
	}
	return fontList[1]
}

// MatchSystemFontsFromIndex returns font path from index into fontList[0 or 1]
func MatchSystemFontsFromIndex(indx int) string {
	/*ret, err := findfont.Find(font)
	if err != nil {
		eztools.Log("failed to find font", font, err)
	}*/
	return fontList[0][indx]
}

// MatchSystemFontsFromPath is reverse function for matchSystemFontsFromIndex
// Return value: eztools.InvalidID if not found
func MatchSystemFontsFromPath(str string) int {
	i, ok := fontMap[str]
	if !ok {
		return eztools.InvalidID
	}
	return i
}
