package main

import (
	"gitlab.com/bon-ami/ezcomm/guiFyne"
)

var (
	// Ver & Bld may or may not be useful
	Ver, Bld string
)

type Guis interface {
	// GuiSetGlbPrm is run at the beginning,
	//   to initialize following for ezcomm.
	//   Ver, Bld, GuiConnected, GuiEnded, GuiLog, GuiRcv, GuiSnt
	GuiSetGlbPrm(Ver, Bld string)
	// GuiRun is run at the end to handle UI in the main thread
	GuiRun()
}

func main() {
	var gui guiFyne.GuiFyne
	gui.GuiSetGlbPrm(Ver, Bld)

	gui.GuiRun()
}
