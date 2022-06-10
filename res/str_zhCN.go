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
		StrReboot4Change     = "应用重新启动后变化生效"
	`)
}
