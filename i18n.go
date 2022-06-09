package main

import (
	"path/filepath"
	"strings"

	"gitee.com/bon-ami/eztools/v4"
	"github.com/flopp/go-findfont"
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
)

func i18nInit() {
	eztools.InitLanguages()
	eztools.AddLanguage("en", `
		StrInt = "interative"
		StrCfg = "config"
		StrLcl = "local"
		StrRmt = "remote"
		StrLst = "listen"
		StrDis = "disconnect peer"
		StrStp = "stop listening"
		StrCon = "connect"
		StrAdr = "address"
		StrPrt = "port"
		StrRec = "content records"
		StrCnt = "content"
		StrSnd = "send"
		StrTo  = "send to"
		StrFrm = "received from"
		StrFlw = "flow"
		StrStb = "standing by"
		StrAll = "select all"
		StrLog  = "log file"
		StrLang = "languages"
		StrVbs  = "verbose level"
		StrHgh  = "high"
		StrMdm  = "medium"
		StrLow  = "low"
		StrNon  = "none"
		StrFnt      = "fonts"
		StrFnt4Lang = "set this font for this language"
		StrReboot4Change     = "restart of this application will show the change"
	`)
	eztools.AddLanguage("es", `
		StrInt = "interaccion"
		StrCfg = "configuraccion"
		StrLcl = "local"
		StrRmt = "remote"
		StrLst = "listen"
		StrDis = "disconnect peer"
		StrStp = "stop listening"
		StrCon = "connect"
		StrAdr = "address"
		StrPrt = "port"
		StrRec = "content records"
		StrCnt = "content"
		StrSnd = "send"
		StrTo  = "send to"
		StrFrm = "received from"
		StrFlw = "flow"
		StrStb = "standing by"
		StrAll = "select all"
		StrLog  = "log file"
		StrLang = "languages"
		StrVbs  = "verbose level"
		StrHgh  = "high"
		StrMdm  = "medium"
		StrLow  = "low"
		StrNon  = "none"
		StrFnt      = "fonts"
		StrFnt4Lang = "set this font for this language"
		StrReboot4Change     = "restart of this application will show the change"
	`)
	eztools.AddLanguage("zh-CN", `
		StrInt = "交互"
		StrCfg = "设置"
		StrLcl = "本地"
		StrRmt = "远端"
		StrLst = "监听"
		StrDis = "断开对端"
		StrStp = "停止监听"
		StrCon = "连接"
		StrAdr = "地址"
		StrPrt = "端口"
		StrRec = "内容记录"
		StrCnt = "内容"
		StrSnd = "发送"
		StrTo  = "发送到"
		StrFrm = "接收于"
		StrFlw = "流程"
		StrStb = "待命"
		StrAll = "全选"
		StrLog  = "日志文件"
		StrLang = "语言"
		StrVbs  = "日志级别"
		StrHgh  = "高"
		StrMdm  = "中"
		StrLow  = "少"
		StrNon  = "无"
		StrFnt      = "字体"
		StrFnt4Lang = "为此语言设置此字体"
		StrReboot4Change     = "应用重新启动后变化生效"
	`)
}

func i18nLoad(lang string) {
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
}

// fontList contains font paths and names
var (
	fontList [2][]string
	fontMap  map[string]int
)

func listSystemFonts() []string {
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

// matchSystemFontsFromIndex returns font path from index into fontList[0 or 1]
func matchSystemFontsFromIndex(indx int) string {
	/*ret, err := findfont.Find(font)
	if err != nil {
		eztools.Log("failed to find font", font, err)
	}*/
	return fontList[0][indx]
}

// matchSystemFontsFromPath is reverse function for matchSystemFontsFromIndex
// Return value: eztools.InvalidID if not found
func matchSystemFontsFromPath(str string) int {
	i, ok := fontMap[str]
	if !ok {
		return eztools.InvalidID
	}
	return i
}
