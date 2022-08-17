#!/bin/bash
#S=/d/sdk/build-tools/33.0.0/apksigner.bat
S=jarsigner
K=/d/ezproject/ez.jks
P=passwd
L=keyez
A=EZComm
W=guiFyne
E=ext

function chkMS() {
	if [ "MSYS2" == "`grep ^NAME /etc/os-release|awk -F '=' '{print $2}'`" ]; then
		return 0
	else
		return 1
	fi
}

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

function lwin1() {
        local M=$3
        if [ -z "$M" ]; then
                M=_debug
        else
                M=
        fi
        cp FyneApp.bak FyneApp.toml
        echo building $1 $M `grep "Build" FyneApp.toml`
        fyne package -os $1 $3 -appVersion $V
        if [ -f "${A}.tar.xz" ]; then
			tar -xf ${A}.tar.xz usr/local/bin/${W}
			mv usr/local/bin/${W} ${W}
			rmdir usr/local/bin
			rmdir usr/local
			rmdir usr
        fi

        if [ -z "$M" ]; then
                M=_$V
        fi
        if [ -f "$W$2" ]; then
                mv ${W}$2 ${A}$M$2
                echo ${W} done with ${A}$M$2
        elif [ -f "$A$2" ]; then
                mv ${A}$2 ${A}$M$2
                echo ${A} done with ${A}$M$2
        else
			echo "$W$2 or $A$2 NOT found!"
        fi
}

if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
        V=$1
	echo "Version $V"

        cp FyneApp.toml FyneApp.bak

		chkMS
		if (( $? == 0 )); then
			if [ -f "${A}.apk" ]; then
					rm "${A}.apk"
			fi
			android1 "arm64"

			lwin1 windows .exe ""
			lwin1 windows .exe "--release"
		else
			lwin1 linux "" ""
			lwin1 linux "" "--release"
		fi

        #cp FyneApp.bak FyneApp.toml
        #echo building Linux `grep "Build" FyneApp.toml`
        #fyne package -os linux -appVersion $V

        echo `grep "Build" FyneApp.toml`
fi
