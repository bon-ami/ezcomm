#!/bin/bash
function cp1() {
	echo trying $1/$2
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
		return 1
	fi
	for f in ${fa[@]}; do
		cp1 $d $f
		if [ $? -ne 0 ]; then
			return 2
		fi
	done
	return 0
}

if [ $# -lt 1 ]; then
	echo "Version (V)X.X.X missing"
	V=0.0.0
else
	V=$1
	V=${V#V}
	V=${V#v}
	shift 1
fi
fa=(EZComm EZComm_debug EZComm.exe EZComm_debug.exe EZComm.apk EZComm_debug.apk)
bld1 guiFyne $V ${@}
ret=$?
if [ ${ret} -ne 0 ]; then
	echo "failed ${ret}"
	exit ${ret}
fi
fa=(EZComm_cmd EZComm_cmd.exe)
bld1 cmd $V ${@}
ret=$?
if [ ${ret} -ne 0 ]; then
	echo "failed ${ret}"
	exit ${ret}
fi