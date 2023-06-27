#!/bin/bash
#S=jarsigner
S=apksigner
K=ez.jks
P=passwd
L=keyez
A=EZComm
W=guiFyne
D=_debug

function chkMS() {
	if [ "MSYS2" == "`grep ^NAME /etc/os-release|awk -F '=' '{print $2}'`" ]; then
		return 0
	else
		return 1
	fi
}

function bld1() {
	local OS1=$1
	local VAR1=$2
	local SUFF
	local TZ=.tar.xz
	case ${OS1} in
		windows)
			SUFF=.exe
			if [[ `go env GOHOSTOS` == linux ]]; then
				if [ -f /usr/bin/x86_64-w64-mingw32-gcc ]; then
					export CGO_ENABLED=1
					export CC=/usr/bin/x86_64-w64-mingw32-gcc
				else
					echo NO cross compiler found!
					return 1
				fi
			fi
			;&
		linux)
			[ -f "${A}${TZ}" ] && rm ${A}${TZ}
			[ -f "${W}${SUFF}" ] && rm ${W}${SUFF}
			[ -f "${A}${SUFF}" ] && rm ${A}${SUFF}
			;;
		android*)
			if [ -n "${VAR1}" ]; then
				SUFF=${D}.apk
				if [ -z "$S" ]; then
					echo "NO signer"
					return 1
				fi
			else
				SUFF=.apk
				S=
			fi
			[ -f "${A}.apk" ] && rm ${A}.apk
			[ -f "${A}.aab" ] && rm ${A}.aab
			;;
	esac
	fyne package -os ${OS1} ${VAR1} -appVersion $V
	local ret=$?
	case ${OS1} in
		windows)
			if [[ `go env GOHOSTOS` == linux ]]; then
				# TODO: env restore
				unset CGO_ENABLED
				unset CC
			fi
			;&
		linux)
			if [ $ret -ne 0 ]; then
				return 2
			fi
			if [ -f "${A}${TZ}" ]; then
				echo "Extracting ${W}${SUFF} from ${A}${TZ}"
				tar -xf ${A}${TZ} usr/local/bin/${W}
				if (( "$?" == 0 )); then
					mv usr/local/bin/${W} ${W}${SUFF}
					rmdir usr/local/bin
					rmdir usr/local
					rmdir usr
					rm ${A}${TZ}
				else
					echo "${W} NOT found in ${A}${TZ}"
					return 2
				fi
			fi
			local DSUFF
			if [ -z "${VAR1}" ]; then
				# debug
				DSUFF=${D}${SUFF}
			else
				DSUFF=${SUFF}
			fi
			[ -f "${W}${SUFF}" ] && mv ${W}${SUFF} ${A}${DSUFF}
			if [ $? -eq 0 ]; then
				return 0
			fi
			if [ "${DSUFF}" == "${SUFF}" ]; then
				if [ -f "${A}${SUFF}" ]; then
					return 0
				fi
			else
				[ -f "${A}${SUFF}" ] && mv ${A}${SUFF} ${A}${DSUFF}
				if [ $? -eq 0 ]; then
					return 0
				fi
			fi
			echo "${W}${SUFF} and ${A}${SUFF} NOT found!"
			return 3
			;;
		android*)
			if [ $ret -ne 0 ]; then
				return 2
			fi
			if [ -f "${A}.apk" ]; then
				if [ -z "${S}" ]; then
					# debug
					mv ${A}.apk ${A}${SUFF}
					echo done with ${A}${SUFF}
					return 0
				fi
			else
				return 1
			fi

			# retail
			if [[ "${S}" != *.bat && `go env GOHOSTOS` == windows ]]; then
				S="${S}.bat"
			fi
			if [ -f "${A}.apk" -a -f "$K" ]; then
				$S sign --ks-key-alias $L --ks-pass file:$P --ks $K ${A}.apk
				if [ $? -ne 0 ]; then
					return 3
				fi
			else
				echo "NOT personally signed"
				ls "$S" "$K"
				return 3
			fi
			#fyne release -os android -appID io.sourceforge.ezproject.ezcomm -appVersion $V -appBuild 1 -keyStore $K
			;;
	esac
	return 0
}

function loop() {
	local VAR=("")
	local V

	if [ $# -gt 0 ]; then
		VAR[1]="--release"
		V=$1
		echo "Version $V"
	else
		echo "Version X.X.X = 0.0.0"
		V=0.0.0
	fi
	local OSS=("android/arm64" "windows")
	chkMS
	if (( $? != 0 )); then
		OSS[2]="linux"
	fi
	if [ ! -d "$ANDROID_HOME/ndk-bundle" -a ! -d "$ANDROID_NDK_HOME" ]; then
		echo NO NDK found!
		return 1
	fi
	cp FyneApp.toml FyneApp.bak
	local i=0
	while (( $i < ${#VAR[@]} )); do
		for O in ${OSS[@]}; do
			cp FyneApp.bak FyneApp.toml
			echo building $O ${VAR[$i]} $(grep "Build" FyneApp.toml)
			bld1 "$O" "${VAR[$i]}"
			local ret=$?
			if [ $ret -ne 0 ]; then
				return $ret
			fi
		done
		let i+=1
	done
	return 0
}

loop $@
exit $?
