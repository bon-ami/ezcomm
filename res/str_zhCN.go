package res

import (
	"gitee.com/bon-ami/eztools/v4"
)

func I18n_zhCN() {
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
		StrFntRch   = "世界你好"
		StrReboot4Change     = "应用重新启动后变化生效"

		StrVer       = "版本信息"
		StrHlp       = "帮助信息"
		StrV         = "记录连接事件"
		StrVV        = "并记录断连详情"
		StrVVV       = "再记录收发详情"
		StrLogFn     = "日志文件名"
		StrCfgFn     = "设置文件名"
		StrFlowFnInf = "输入文件名能控制流程"

		StrFlowOpenErr   = "打开流程文件出错"
		StrFlowOpenNot   = "未打开流程文件"
		StrFlowRunning   = "正在运行流程"
		StrFlowRunNot    = "流程未运行"
		StrOK            = "通过"
		StrNG            = "错误"
		StrFlowFinAs     = "流程结束为"
		StrCut           = "剪切"
		StrCpy           = "拷贝"
		StrPst           = "粘贴"
		StrCpyAll        = "拷贝所有"
		StrClr           = "清空"
		StrListeningOn   = "监听于"
		StrConnFail      = "无法连接到"
		StrConnected     = "已连接"
		StrDisconnected  = "已断连"
		StrNotInRec      = "不在记录中"
		StrSvrIdle       = "服务器空闲"
		StrClntLft       = "剩余客户端"
		StrNoPeer4       = "找不到对应对端"
		StrDisconnecting = "正在断连"
		StrStopLstn      = "已停止监听"
		StrFl2Rcv        = "无法收取"
		StrGotFromSw     = "收自某处"
		StrGotFrom       = "收自"
		StrUnknownDsc    = "未知对端断连"
		StrFl2Snd        = "无法发送"
		StrSw            = "某处"
		StrSnt2          = "已发送到"
		StrInfLog        = "日志"
	`)
}