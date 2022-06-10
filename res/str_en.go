package res

import (
	"gitee.com/bon-ami/eztools/v4"
)

func I18n_en() {
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
}
