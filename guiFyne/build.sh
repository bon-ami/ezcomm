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
        if [ -f "${A}.apk" ]; then
			rm ${A}.apk
        fi
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
                if [ -f "${A}_${V}${M}.apk" ]; then
					rm ${A}_${V}${M}.apk
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
        if [ -f "${A}.tar.xz" ]; then
			rm ${A}.tar.xz
        fi
        if [ -f "$W$2" ]; then
                rm ${W}$2
        elif [ -f "$A$2" ]; then
                rm ${A}$2
        fi
        if [[ `go env GOHOSTOS` == linux && "$1" == windows ]]; then
                export CGO_ENABLED=1
                export CC=/usr/bin/x86_64-w64-mingw32-gcc
        fi
        fyne package -os $1 $3 -appVersion $V
        if [[ `go env GOHOSTOS` == linux && "$1" == windows ]]; then
                unset CGO_ENABLED
                unset CC
        fi
        echo $?

        if [ -f "${A}.tar.xz" ]; then
			echo "Extracting ${W}$2 from ${A}.tar.xz"
			tar -xf ${A}.tar.xz usr/local/bin/${W}
			if (( "$?" == 0 )); then
				mv usr/local/bin/${W} ${W}$2
				rmdir usr/local/bin
				rmdir usr/local
				rmdir usr
				rm ${A}.tar.xz
			else
				echo "${W} NOT found in ${A}.tar.xz"
				return
			fi
        fi

        #if [ -z "$M" ]; then
                #M=_$V
        #fi
        if [ -f "$W$2" ]; then
                mv ${W}$2 ${A}$M$2
                ls -l ${A}$M$2
                echo ${W} done with ${A}$M$2
        elif [ -f "$A$2" ]; then
			if [ - z "$M" ]; then
				ls -l ${A}$2
				echo done with ${A}$2
			else
                mv ${A}$2 ${A}$M$2
                ls -l ${A}$M$2
                echo ${A}$2 done with ${A}$M$2
            fi
        else
			echo "$W$2 and $A$2 NOT found!"
        fi
}

if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
        V=$1
	echo "Version $V"

        cp FyneApp.toml FyneApp.bak

		#chkMS
		#if (( $? == 0 )); then
			if [ -f "${A}.apk" ]; then
					rm "${A}.apk"
			fi
			android1 "arm64"

			lwin1 windows .exe ""
			lwin1 windows .exe "--release"
		#else
			lwin1 linux "" ""
			lwin1 linux "" "--release"
		#fi

        echo `grep "Build" FyneApp.toml`
fi
