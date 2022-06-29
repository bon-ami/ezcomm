#!/bin/bash
if [ $# -lt 2 ]; then
	echo "params: {resource file} {static name} [{output file, defaults to bundle.go}]"
else
        OP=""
        AP=""
        if [ $# -gt 2 ]; then
                OP="-o $3"
                if [ -e "$3" ]; then
                        AP=" -a"
                fi
        else
                if [ -e "bundle.go" ]; then
                        AP=" -a"
                fi
        fi
        echo fyne bundle${AP} --name \"$2\" --pkg guiFyne $OP \"$1\"
        fyne bundle${AP} --name "$2" --pkg guiFyne $OP "$1"
fi
