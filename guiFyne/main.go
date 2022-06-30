package main

import (
	"gitlab.com/bon-ami/ezcomm"
)

var (
	// Ver & Bld may or may not be useful
	Ver, Bld string
	// hopefully, Other Guis also do
	gui GuiFyne
)

func main() {
	ezcomm.SetGui(gui)
	gui.Run(Ver, Bld)
}
