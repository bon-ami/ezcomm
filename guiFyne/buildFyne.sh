#!/bin/bash
if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
        V=$1
	echo "Version $V"

        cp FyneApp.toml FyneApp.bak
	echo building Android `grep "Build" FyneApp.toml`
	fyne package -os android -appVersion $V

        cp FyneApp.bak FyneApp.toml
        echo building Windows `grep "Build" FyneApp.toml`
        fyne package -os windows -appVersion $V
        mv gui.exe "EZ_Comm.exe"

        cp FyneApp.bak FyneApp.toml
        echo building Linux `grep "Build" FyneApp.toml`
        fyne package -os linux -appVersion $V
        echo `grep "Build" FyneApp.toml`
fi
