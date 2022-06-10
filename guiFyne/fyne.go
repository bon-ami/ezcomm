package guiFyne

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
	"gitlab.com/bon-ami/ezcomm"
	"gitlab.com/bon-ami/ezcomm/res"
)

var thm theme4Fonts

func UI() {
	ezcomm.GuiConnected = GuiConnected
	ezcomm.GuiEnded = GuiEnded
	ezcomm.GuiLog = GuiLog
	ezcomm.GuiRcv = GuiRcv
	ezcomm.GuiSnt = GuiSnt

	//eztools.Log("setting font=", cfgStruc.font)
	thm.SetFont(ezcomm.CfgStruc.GetFont())
	ezcApp := app.New()
	ezcApp.Settings().SetTheme(&thm)
	/*icon, err := fyne.LoadResource("Icon.png")
	if err == nil {*/
	ezcApp.SetIcon(res.ResourceIconPng)
	//}
	ezcWin := ezcApp.NewWindow(ezcomm.EzcName)

	contLcl := guiFyneMakeControlsLcl()
	contRmt := guiFyneMakeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)

	tabs := container.NewAppTabs(
		container.NewTabItem(ezcomm.StrInt, cont),
		container.NewTabItem(ezcomm.StrCfg, guiFyneMakeControlsCfg(ezcWin)),
	)
	ezcWin.SetContent(tabs)

	/*selectCtrls = []*widget.SelectEntry{ sockLcl[0], sockLcl[1], //sockRmt[0], sockRmt[1],
		//sockRcv[0], sockRcv[0], //recRmt,
	}*/

	ezcWin.Show()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("to show UI")
	}
	ezcApp.Run()
	if eztools.Debugging && eztools.Verbose > 2 {
		eztools.Log("UI done")
	}
}
