#!/bin/bash
#S=/d/sdk/build-tools/33.0.0/apksigner.bat
S=jarsigner
K=/d/ezproject/ez.jks
P=passwd
L=keyez
A=EZComm
W=guiFyne
E=ext

function android1() {
        local M=$1
        if [ -n "$M" ]; then
                M="/${M}"
        fi
	echo building Android $M `grep "Build" FyneApp.toml`
	fyne package -os android$M -appVersion $V
        if [ -f "${A}.apk" ]; then
                mv ${A}.apk ${A}_debug.apk
                echo done with ${A}_debug.apk
        fi
        echo "retail version now"
        cp FyneApp.bak FyneApp.toml
	fyne package -os android$M --release -appVersion $V
        if [ -f "${A}.apk" -a -n "$S" -a -f "$K" ]; then
                if [ -d "$E" ]; then
                        echo "removing previous temp dir $E"
                        rm -r "$E"
                fi
                echo "to remove Fyne certs"
                unzip -d "$E" ${A}.apk
                pushd "$E"
                rm META-INF/CERT*
                zip -0 -r ${A}.apk -xi *
                mv ${A}.apk ..
                popd
                rm -r "$E"
                M=$1
                if [ -n "$M" ]; then
                        M="_$M"
                fi
                #$S sign --ks $K --out ${A}_${M}.apk ${A}.apk
                if [ -f "$P" ]; then
                        P=`cat $P`
                        P="-storepass $P -keypass $P"
                else
                        echo "NO $P set for password"
                        P=
                fi
                $S -verbose -keystore $K $P -signedjar ${A}_${V}${M}.apk ${A}.apk $L
                if [ -f ${A}_${V}${M}.apk ]; then
                        echo done with ${A}_${V}${M}.apk
                        jarsigner -verbose -verify  ${A}_${V}${M}.apk
                fi
        else
                echo "NOT personally signed"
        fi
        #fyne release -os android -appID io.sourceforge.ezproject.ezcomm -appVersion $V -appBuild 1 -keyStore ../../ez.jks
}

function windows1() {
        local M=$1
        if [ -z "$M" ]; then
                M=_debug
        else
                M=
        fi
        cp FyneApp.bak FyneApp.toml
        echo building Windows $M `grep "Build" FyneApp.toml`
        fyne package -os windows $1 -appVersion $V
        if [ -f "$W" ]; then
                mv ${W}.exe "${A}$M.exe"
                echo done with ${A}$M.exe
        fi
}

if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
        V=$1
	echo "Version $V"

        cp FyneApp.toml FyneApp.bak
        if [ -f "${A}.apk" ]; then
                rm "${A}.apk"
        fi
        android1 "arm64"

        windows1 ""
        windows1 "--release"

        #cp FyneApp.bak FyneApp.toml
        #echo building Linux `grep "Build" FyneApp.toml`
        #fyne package -os linux -appVersion $V

        echo `grep "Build" FyneApp.toml`
fi
