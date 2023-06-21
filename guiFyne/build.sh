#!/bin/bash
#S=/d/sdk/build-tools/33.0.0/apksigner.bat
#S=jarsigner
S=apksigner
K=ez.jks
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
	if [ ! -d "$ANDROID_HOME/ndk-bundle" -a ! -d "$ANDROID_NDK_HOME" ]; then
		echo NO NDK found!
		return 1
	fi
        local M=$1
        if [ -n "$M" ]; then
                M="/${M}"
        fi
	echo building Android $M `grep "Build" FyneApp.toml`
        [ -f "${A}.apk" ] && rm ${A}.apk
        [ -f "${A}.aab" ] && rm ${A}.aab
	fyne package -os android$M -appVersion $V
	if [ $? -ne 0 ]; then
		return 2
	fi
        if [ -f "${A}.apk" ]; then
                mv ${A}.apk ${A}_debug.apk
                echo done with ${A}_debug.apk
		else
			return 2
        fi
        echo "retail version now"
        cp FyneApp.bak FyneApp.toml
        [ -f "${A}.aab" ] && rm ${A}.aab
	fyne package -os android$M --release -appVersion $V
	if [ $? -ne 0 ]; then
		return 2
	fi
        if [ -f "${A}.apk" -a -n "$S" -a -f "$K" ]; then
			which $S
			if [ $? -ne 0 -a ! -f $S ]; then
				echo $S NOT found!
				return 2
			fi
                $S sign --ks-key-alias $L --ks-pass file:$P --ks $K ${A}.apk
				if [ $? -ne 0 ]; then
					return 3
				fi
        else
                echo "NOT personally signed"
                return 3
        fi
        #fyne release -os android -appID io.sourceforge.ezproject.ezcomm -appVersion $V -appBuild 1 -keyStore $K
	return 0
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
		if [ -f /usr/bin/x86_64-w64-mingw32-gcc ]; then
			export CGO_ENABLED=1
			export CC=/usr/bin/x86_64-w64-mingw32-gcc
		else
			echo NO cross compiler found!
			return 1
		fi
        fi
        fyne package -os $1 $3 -appVersion $V
		if [ $? -ne 0 ]; then
			return 2
		fi
        if [[ `go env GOHOSTOS` == linux && "$1" == windows ]]; then
                unset CGO_ENABLED
                unset CC
        fi

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
				return 2
			fi
        fi

        #if [ -z "$M" ]; then
                #M=_$V
        #fi
        if [ -f "$W$2" ]; then
                mv ${W}$2 ${A}$M$2
                ls -l ${A}$M$2
				if [ $? -ne 0 ]; then
					return 3
				fi
                echo ${W} done with ${A}$M$2
        elif [ -f "$A$2" ]; then
			if [ -z "$M" ]; then
				ls -l ${A}$2
				if [ $? -ne 0 ]; then
					return 3
				fi
				echo done with ${A}$2
			else
                mv ${A}$2 ${A}$M$2
                ls -l ${A}$M$2
				if [ $? -ne 0 ]; then
					return 3
				fi
                echo ${A}$2 done with ${A}$M$2
            fi
        else
			echo "$W$2 and $A$2 NOT found!"
			return 3
        fi
        return 0
}

if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
	exit 1
else
        V=$1
fi
echo "Version $V"

cp FyneApp.toml FyneApp.bak

	#chkMS
	#if (( $? == 0 )); then
		if [ -f "${A}.apk" ]; then
				rm "${A}.apk"
		fi
		android1 "arm64"
		if [ $? -ne 0 ]; then
			exit 1
		fi

		lwin1 windows .exe ""
		if [ $? -ne 0 ]; then
			exit 2
		fi
		lwin1 windows .exe "--release"
		if [ $? -ne 0 ]; then
			exit 3
		fi
	#else
		lwin1 linux "" ""
		if [ $? -ne 0 ]; then
			exit 4
		fi
		lwin1 linux "" "--release"
		if [ $? -ne 0 ]; then
			exit 5
		fi
	#fi

echo `grep "Build" FyneApp.toml`