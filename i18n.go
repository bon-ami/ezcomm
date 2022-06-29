package ezcomm

import (
	"path/filepath"
	"strings"

	"gitee.com/bon-ami/eztools/v4"
	"github.com/bon-ami/go-findfont"
	"gitlab.com/bon-ami/ezcomm/res"
)

var (
	StringTran map[string]string
	StringIndx = [...]string{
		"StrInt",
		"StrCfg",
		"StrLcl",
		"StrRmt",
		"StrLst",
		"StrDis",
		"StrStp",
		"StrCon",
		"StrAdr",
		"StrPrt",
		"StrRec",
		"StrCnt",
		"StrSnd",
		"StrTo",
		"StrFrm",
		"StrFlw",
		"StrStb",
		"StrAll",
		"StrLog",
		"StrLang",
		"StrVbs",
		"StrHgh",
		"StrMdm",
		"StrLow",
		"StrNon",
		"StrFnt",
		"StrFnt4Lang",
		"StrReboot4Change",
		"StrFntRch",
		"StrVer",
		"StrHlp",
		"StrV",
		"StrVV",
		"StrVVV",
		"StrLogFn",
		"StrCfgFn",
		"StrFlowFnInf",
		"StrFlowOpenErr",
		"StrFlowOpenNot",
		"StrFlowRunning",
		"StrFlowRunNot",
		"StrOK",
		"StrNG",
		"StrFlowFinAs",
		"StrCut",
		"StrCpy",
		"StrPst",
		"StrCpyAll",
		"StrClr",
		"StrListeningOn",
		"StrConnFail",
		"StrConnected",
		"StrDisconnected",
		"StrNotInRec",
		"StrSvrIdle",
		"StrClntLft",
		"StrNoPeer4",
		"StrDisconnecting",
		"StrStopLstn",
		"StrFl2Rcv",
		"StrGotFromSw",
		"StrGotFrom",
		"StrUnknownDsc",
		"StrFl2Snd",
		"StrSw",
		"StrSnt2",
		"StrInfLog",
	}
)

func I18nInit() {
	eztools.InitLanguages()
	res.I18n_en()
	res.I18n_es()
	res.I18n_zhCN()
	StringTran = make(map[string]string)
}

func I18nLoad(lang string) (string, error) {
	langOut, err := eztools.LoadLanguage(lang)
	if err != nil {
		eztools.Log("failed to load language for", lang, err)
		return lang, err
	}
	if eztools.Debugging {
		LogPrintFunc("loading language", lang, langOut)
	}
	for _, i := range StringIndx {
		//LogPrintFunc("loading", i)
		str, err := eztools.GetLanguageStr(i)
		//LogPrintFunc("loading", str, err)
		if err != nil {
			LogPrintFunc("no translation for", i)
			continue
		}
		StringTran[i] = str
	}
	return langOut, err
}

// fontList contains font paths and names
var (
	fontList [2][]string
	fontMap  map[string]int
)

// ListSystemFonts get all system fonts with extensions
func ListSystemFonts(exts []string) []string {
	fontList[0] = findfont.ListWtSuffixes(exts)
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
