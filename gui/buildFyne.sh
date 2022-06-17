#!/bin/bash
if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
	echo "Version $1"
	echo building Android `grep "Build" FyneApp.toml`
	fyne package -os android -appVersion $1
	echo building Windows `grep "Build" FyneApp.toml`
	fyne package -os windows -appVersion $1
	mv gui.exe "EZ_Comm.exe"
	echo building Linux `grep "Build" FyneApp.toml`
	fyne package -os linux -appVersion $1
	echo `grep "Build" FyneApp.toml`
fi