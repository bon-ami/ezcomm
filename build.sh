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

fa=(EZComm EZComm_debug EZComm.exe EZComm_debug.exe)
bld1 guiFyne ${@}
fa=(EZComm_debug)
bld1 cmd ${@}
