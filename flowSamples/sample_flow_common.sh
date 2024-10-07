#!/bin/bash

function flow_common_main {
	local direction=$1
	shift 1

	local testTemp=testdata/flow.bat.msg
	local param=
	local local=
	local remote=
	local file=
	until (( $# < 1 )) ; do
		local one=$1
		if [[ "$one" == --* ]]; then
			one=${one:1}
		fi
		echo processing $one.
		case $one in
			-local)
				if [ -n "$param" ]; then
					echo "$param NEEDS a value!"
				fi
				param="local" ;;
			-remote)
				if [ -n "$param" ]; then
					echo "$param NEEDS a value!"
				fi
				param="remote" ;;
			-file)
				if [ -n "$param" ]; then
					echo "$param NEEDS a value!"
				fi
				param="file" ;;
			*)
				if [ -z "$param" ]; then
					echo "$one is an ORPHAN or UNRECOGNIZED value!"
				else
					case $param in
						file) file=$one ;;
						local) local=$one ;;
						remote) remote=$one ;;
						*) echo "$param is UNRECOGNIZED!" ;;
					esac
					param=
				fi ;;
		esac
		shift 1
	done
	echo local=$local, remote=$remote, file=$file
	if [ -z "$file" -o -z "$local" ]; then
		return 1
	fi
	echo direction=$direction
	case "$direction" in
		in)
			echo >> $file
			echo IN: from $remote to $local >> $file
			mv $file $testTemp;;
		out)
			echo OUT: from $local to $remote >> $testTemp
			mv $testTemp $file;;
		*)
			echo "UNKNOWN caller!"; return 1;;
	esac
	return 0
}

if (( $# < 1 )); then
	echo "sample_flow_common.sh <in|out> [--local] [--remote] [--file]"
	exit 2
else
	flow_common_main $@
	exit $?
fi
