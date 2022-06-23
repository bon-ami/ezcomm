package main

import (
	"gitlab.com/bon-ami/ezcomm"
	"gitlab.com/bon-ami/ezcomm/guiFyne"
)

var (
	// Ver & Bld may or may not be useful
	Ver, Bld string
	gui      guiFyne.GuiFyne
)

func main() {
	ezcomm.SetGui(gui)
	gui.Run(Ver, Bld)
}
