#!/bin/bash
if [ $# -lt 2 ]; then
	echo "params: {resource file} {static name} [{output file, defaults to bundle.go}]"
else
        OP=""
        AP=""
        AP_STR=" -a"
        if [ $# -gt 2 ]; then
                OP="-o $3"
                if [ -e "$3" ]; then
                        AP=$AP_STR
                fi
        else
                if [ -e "bundle.go" ]; then
                        AP=$AP_STR
                fi
        fi
        if [ "$AP" == "$AP_STR" ]; then
                echo "appending"
        fi
        echo fyne bundle${AP} --name \"$2\" --pkg main $OP \"$1\"
        fyne bundle${AP} --name "$2" --pkg main $OP "$1"
fi
