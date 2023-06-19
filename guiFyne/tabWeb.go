package main

import (
	"html"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gitee.com/bon-ami/eztools/v6"
	"gitlab.com/bon-ami/ezcomm"
)

const (
	httpAppDir       = "appdir"
	httpSmsDir       = "/"
	httpSmsID        = "sms"
	httpTemplateHead = `<head>
        <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
        <title>`
	httpTemplateStyles = `</title>
        <style type="text/css">
                html,body{height: 100%;margin: 0;}
                .header{border-bottom:1px solid #ccc;}
                .sidebar{width:25%;float:left;}
                .main{width:75%;float:left;
                        display: flex;height: 100%;
                        flex: 1 1 auto;flex-flow: column;}
                .multiinput {border: solid 2px #999;}
        </style>
</head>
<body>
        <div class="sidebar">
                <div class="header">
                        <h3>
                                <a href="`
	httpTemplateEzc = `">`
	httpTemplateVer = `</a>
                        </h3>
                        <div align="right">`
	httpTemplatePrjLnk = `<div style="font-size:70%">of
<a href="`
	httpTemplatePrjName = `">`
	httpTemplateSMSForm = `</a>
                        </div></div>
                </div>
                <form action="`
	httpTemplateSMSCtrl = `" method="post" enctype = "application/x-www-form-urlencoded" >
			<textarea class="multiinput" name="`
	httpTemplateSMSHint = `" placeholder="`
	httpTemplateSMSText = ` ">`
	httpTemplateSubmit  = `</textarea>
                        <input type="submit"/>
                </form>
        </div>
        <div id="main" class="main">
                <iframe width="99%" height="100%" src="`
	httpTemplateTail = `">
                </iframe>
        </div>
</body>
</html>`
)

var (
	cntWeb               *widget.Entry
	hdrWebLcl, hdrWebRmt *widget.Accordion
	smsWeb               string
)

func httpTemplate2Page() (ret ezcomm.HTTPSvrBody) {
	ret.Tp = ezcomm.HTTPSvrBodyHTML
	ret.Str = eztools.HTMLHead + httpTemplateHead + ezcomm.EzcName +
		httpTemplateStyles +
		ezcomm.EzcURL + ezcomm.StringTran["StrURLReadMe"] +
		httpTemplateEzc + ezcomm.CfgStruc.EzcName() + " readme" +
		httpTemplateVer + ezcomm.Ver + "." + ezcomm.Bld +
		httpTemplatePrjLnk + eztools.EzpURL +
		httpTemplatePrjName + ezcomm.CfgStruc.EzpName() +
		httpTemplateSMSForm + httpSmsDir +
		httpTemplateSMSCtrl + httpSmsID +
		httpTemplateSMSHint + ezcomm.StringTran["StrTxt2ShareBtCS"] +
		httpTemplateSMSText + html.EscapeString(smsWeb) +
		httpTemplateSubmit + httpAppDir + httpTemplateTail
	return
}

// printIncomingWeb is for GET
func printIncomingWeb(ip string, req *http.Request, _ func(string) string) (
	int, ezcomm.HTTPSvrBody, map[string]string) {
	add2Rmt(0, ip)
	hdrWebRmt.Items = nil
	/*hdrs := req.Header
	hdrLst := make([]string, 0)*/
	for hdr1, cont1 := range req.Header {
		ent := NewMultiLineEntry()
		var val string
		//hdrLst = append(hdrLst, hdr1)
		for _, c1 := range cont1 {
			if len(val) > 0 {
				val += "\n"
			}
			val += strings.ReplaceAll(c1, ",", "\n")
			//hdrs[hdr1] = strings.Split(c1, ",")
			//Log(c1)
		}
		ent.SetText(val)
		hdrWebRmt.Append(widget.NewAccordionItem(hdr1, ent))
	}
	//hdrs[""] = hdrLst
	/*hdrWebRmt.RemoveAll()
	hdrWebRmt.Add(widget.NewTreeWithStrings(hdrs))*/
	hdrWebRmt.Refresh()
	//Log(hdrs)
	return http.StatusOK, httpTemplate2Page(), nil
}

// postIncomingWeb is for POST
func postIncomingWeb(ip string, req *http.Request,
	postFun func(string) string) (
	int, ezcomm.HTTPSvrBody, map[string]string) {
	smsWeb = postFun(httpSmsID)
	cntWeb.SetText(smsWeb)
	return http.StatusOK, httpTemplate2Page(), nil
}

/*func addOutgoingWebHdr(uid widget.TreeNodeID) {
	if uid == "+" {
		return
	}
	//hdrWebTree.IsBranch(uid)
}*/

func makeControlsWebHost() *fyne.Container {
	/*hdrWebTree = widget.NewTreeWithStrings(map[string][]string{
		"":  {"+", "-"},
		"+": {"0"},
		"-": {"0"},
	})
	hdrWebTree.OnSelected = addOutgoingWebHdr*/
	hdrWebLcl = widget.NewAccordion(widget.NewAccordionItem("+",
		widget.NewButton("+", func() {
			k := widget.NewEntry()
			v := widget.NewEntry()
			items := []*widget.FormItem{
				widget.NewFormItem(
					ezcomm.StringTran["StrHeader"], k),
				widget.NewFormItem(
					ezcomm.StringTran["StrValue"], v),
			}
			dialog.ShowForm("+", ezcomm.StringTran["StrOK"],
				ezcomm.StringTran["StrNG"], items,
				func(b bool) {
					if !b {
						return
					}
					var rmvHdr func()
					valEntry := NewMultiLineEntry()
					valEntry.OnChanged = func(str string) {
						//Log(k.Text, "->", v.Text, str)
						if !strings.HasPrefix(str, "\n") {
							return
						}
						rmvHdr()
					}
					valEntry.OnSubmitted = func(str string) {
						//Log(k.Text, ",", v.Text, ".", str)
						if len(str) > 0 {
							return
						}
						rmvHdr()
					}
					valEntry.PlaceHolder = ezcomm.StringTran["StrSE2Del"]
					valEntry.SetText(v.Text)
					hdrItem := widget.NewAccordionItem(
						k.Text, valEntry)
					hdrWebLcl.Append(hdrItem)
					rmvHdr = func() {
						hdrWebLcl.Remove(hdrItem)
					}
				}, ezcWin)
		})),
	)
	/*hdrWebLcl = container.NewBorder(nil, nil, nil, nil,
	hdrWebTree)*/
	cntWeb = widget.NewMultiLineEntry()
	cntWeb.SetPlaceHolder(ezcomm.StringTran["StrTxt2ShareBtCS"])
	txtRow := container.NewVBox(cntWeb,
		widget.NewButton(ezcomm.StringTran["StrOK"], func() {
			smsWeb = cntWeb.Text
		}))
	return container.NewBorder(container.NewVBox(makeControlsLclSocks(false)),
		txtRow, nil, nil, hdrWebLcl)
}

func makeControlsWebClnt() *fyne.Container {
	hdrWebRmt = widget.NewAccordion()
	/*hdrWebRmt = container.NewBorder(nil, nil, nil, nil,
	widget.NewLabel(ezcomm.StringTran["StrHeaders"]))*/
	return container.NewBorder(makeControlsRmtSocks(),
		nil, nil, nil, container.NewVScroll(hdrWebRmt))
}

func makeTabWeb() *container.TabItem {
	return container.NewTabItem("HTTP", //ezcomm.StringTran["StrHTTP"],
		container.NewGridWithColumns(2, makeControlsWebHost(), makeControlsWebClnt()))
}

func tabWebShown() {
	protRd.Refresh()
	lstBut.Refresh()
}
