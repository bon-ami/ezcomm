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
		StrFntRch   = "the quick brown fox jumps over the lazy dog"
		StrReboot4Change     = "restart of this application will show the change"

		StrVer       = "version info"
		StrHlp       = "help info"
		StrV         = "log connection events"
		StrVV        = "also log connetion termination details"
		StrVVV       = "also log I/O details"
		StrLogFn     = "log file name"
		StrCfgFn     = "cfg file name"
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
		StrInfLog        = "log"
	`)
}