package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
)

var thm theme4Fonts

func guiFyne() {
	//eztools.Log("setting font=", cfgStruc.font)
	thm.SetFont(cfgStruc.font)
	ezcApp := app.New()
	ezcApp.Settings().SetTheme(&thm)
	/*icon, err := fyne.LoadResource("Icon.png")
	if err == nil {*/
	ezcApp.SetIcon(resourceIconPng)
	//}
	ezcWin := ezcApp.NewWindow(ezcName)

	contLcl := guiFyneMakeControlsLcl()
	contRmt := guiFyneMakeControlsRmt()

	cont := container.NewGridWithColumns(2, contLcl, contRmt)

	tabs := container.NewAppTabs(
		container.NewTabItem(StrInt, cont),
		container.NewTabItem(StrCfg, guiFyneMakeControlsCfg(ezcWin)),
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
