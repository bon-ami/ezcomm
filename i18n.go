package main

import (
	"gitee.com/bon-ami/eztools/v4"
	"github.com/BurntSushi/toml"
	"github.com/Xuanwo/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var i18nBundle *i18n.Bundle

func initStr() {
	i18nBundle = i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	/*i18nBundle.LoadMessageFile("res/en.toml")
	i18nBundle.LoadMessageFile("res/es.toml")
	i18nBundle.LoadMessageFile("res/zh-CN.toml")*/
	i18nBundle.MustParseMessageFileBytes([]byte(`
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
		StrAll = "select all"
	`), "en.toml")
	i18nBundle.MustParseMessageFileBytes([]byte(`
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
		StrAll = "select all"
	`), "es.toml")
	i18nBundle.MustParseMessageFileBytes([]byte(`
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
		StrAll = "全选"
	`), "zh-CN.toml")
}

func loadStrCurr() {
	tag, err := locale.Detect()
	if err != nil {
		eztools.LogPrint("failed to get locale!", err)
		return
	}
	//eztools.LogPrint(tag.String())
	loadStr(tag.String())
}

func loadStr(loc string) {
	localizer := i18n.NewLocalizer(i18nBundle, loc)
	StrLst = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrLst"})
	StrInt = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrInt"})
	StrCfg = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrCfg"})
	StrLcl = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrLcl"})
	StrRmt = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrRmt"})
	StrDis = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrDis"})
	StrStp = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrStp"})
	StrCon = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrCon"})
	StrAdr = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrAdr"})
	StrPrt = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrPrt"})
	StrRec = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrRec"})
	StrCnt = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrCnt"})
	StrSnd = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrSnd"})
	StrTo = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrTo"})
	StrFrm = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrFrm"})
	StrFlw = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrFlw"})
	StrAll = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "StrAll"})
}
