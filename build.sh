#!/bin/bash
function cp1() {
	echo trying to copy $1/$2
	if [ -f $1/$2 ]; then
		cp $1/$2 .
		if [ $? -eq 0 ]; then
			return 0
		fi
	fi
	return 1
}

function bld1() {
	local d=$1
	shift 1
	pushd $d
	./build.sh ${@}
	local ret=$?
	popd
	if [ ${ret} -ne 0 ]; then
		return ${ret}
	fi
	for f in ${fa[@]}; do
		cp1 $d $f
		if [ $? -ne 0 ]; then
			echo FAILED
		fi
	done
	return 0
}

function chkMS() {
	if [ "MSYS2" == "`grep ^NAME /etc/os-release|awk -F '=' '{print $2}'`" ]; then
		return 0
	else
		return 1
	fi
}

if [ $# -gt 0 ]; then
	V=$1
	V=${V#V}
	V=${V#v}
	shift 1
else
	V=
fi
# GUI varies among retail/debug, MS/MSYS2
fa=(EZComm_debug.exe EZComm_debug.apk)
chkMS
MS=$?
let IND=${#fa[@]}
if (( $MS != 0 )); then
	fa[${IND}]=EZComm_debug
	let IND+=1
fi
if [ -n "$V" ]; then
	# retail
	for i in EZComm.exe EZComm.apk; do
		fa[${IND}]=$i
		let IND+=1
	done
	if (( $MS != 0 )); then
		fa[${IND}]=EZComm
		let IND+=1
	fi
fi
bld1 guiFyne $V ${@}
ret=$?
if [ ${ret} -ne 0 ]; then
	echo "failed ${ret}"
	exit ${ret}
fi
# cmd does not vary
fa=(EZComm_cmd EZComm_cmd.exe)
bld1 cmd $V ${@}
ret=$?
if [ ${ret} -ne 0 ]; then
	echo "failed ${ret}"
	exit ${ret}
fi
