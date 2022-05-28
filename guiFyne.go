package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"gitee.com/bon-ami/eztools/v4"
)

func guiFyne() {
	ezcApp := app.New()
	/*icon, err := LoadResourceFromPath("icon.ico")
	if err == nil {
		ezcApp.SetIcon(icon)
	}*/
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
