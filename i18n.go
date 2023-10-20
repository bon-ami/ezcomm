package ezcomm

import (
	"path/filepath"
	"strings"

	"gitee.com/bon-ami/eztools/v6"
	"github.com/bon-ami/go-findfont"
	"gitlab.com/bon-ami/ezcomm/res"
)

var (
	// StringTran maps StringIndx to resource
	StringTran map[string]string
	// StringIndx is key to StringTran
	StringIndx = [...]string{
		"StrURLReadMe",
		"StrEzcName",
		"StrEzpName",
		"StrInt",
		"StrFil",
		"StrDir",
		"StrFilO",
		"StrCfg",
		"StrLcl",
		"StrRmt",
		"StrLst",
		"StrLstFl",
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
		"StrLogExp",
		"StrLang",
		"StrVbs",
		"StrHgh",
		"StrMdm",
		"StrLow",
		"StrNon",
		"StrFnt",
		"StrFnt4Lang",
		"StrFnt4LangBuiltin",
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
		"StrAntiFld",
		"StrLmt",
		"StrPrd",
		"StrTooLarge2Rcv",
		"StrFileAwareness",
		"StrFileSave",
		"StrInfLan",
		"StrPokePeer",
		"StrDiscoverFail",
		"StrAlert",
		"StrNoPerm",
		"StrRcvFil",
		"StrDownloads",
		"StrExp",
		"StrRmAll",
		"StrDel",
		"StrCopied",
		"StrHeader",
		"StrValue",
		"StrSE2Del",
		"StrTxt2ShareBtCS",
		"StrLanHint",
	}
)

// I18nInit inits resource
func I18nInit() {
	eztools.InitLanguages()
	res.I18nEn()
	res.I18nEs()
	res.I18nJa()
	res.I18nZhcn()
	res.I18nZhtw()
	StringTran = make(map[string]string)
}

// I18nLoad loads a language
// Parameter: name of the language, as a param of eztools.AddLanguage.
//
//	It should be locale.
//
// Return values: input language, or detected one if no param input.
func I18nLoad(lang string) (string, error) {
	langOut, err := eztools.LoadLanguage(lang)
	if err != nil {
		eztools.Log("failed to load language for", lang, err)
		return lang, err
	}
	//eztools.Log("loading language", lang, langOut)
	for _, i := range StringIndx {
		//Log("loading", i)
		str, err := eztools.GetLanguageStr(i)
		//Log("loading", str, err)
		if err != nil {
			eztools.Log("NO translation for", i)
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
		Log("failed to find font", font, err)
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
