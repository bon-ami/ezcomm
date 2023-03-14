#!/bin/bash
function cp1() {
	echo trying $1/$2
	[ -f $1/$2 ] && cp $1/$2 .
}

function bld1() {
	local d=$1
	shift 1
	pushd $d
	./build.sh ${@}
	popd
	for f in ${fa[@]}; do
		cp1 $d $f
	done
}

if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
	V=0.0.0
else
        V=$1
	V=${V#V}
	shift 1
fi
fa=(EZComm EZComm_debug EZComm.exe EZComm_debug.exe EZComm.apk EZComm_debug.apk)
bld1 guiFyne $V ${@}
fa=(EZComm_cmd EZComm_cmd.exe)
bld1 cmd $V ${@}
