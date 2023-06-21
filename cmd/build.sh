#!/bin/bash
if [ $# -lt 1 ]; then
	echo "Version X.X.X missing, with optional Build X"
	exit 1
else
        bld=$2
        if [ -z "$bld" ]; then
                bld=`date --rfc-3339=date`
        fi
	echo building Linux $1.$bld
	# nomsgpack for gin
	GOOS=linux go build -v -tags=nomsgpack -ldflags="-X main.Ver=$1" -ldflags="-X main.Bld=$bld" -o EZComm_cmd
	if [ $? -ne 0 ]; then
		exit 2
	fi
	echo building Windows $1.$bld
	GOOS=windows go build -v -tags=nomsgpack -ldflags="-X main.Ver=$1" -ldflags="-X main.Bld=$bld" -o EZComm_cmd.exe
	if [ $? -ne 0 ]; then
		exit 3
	fi
fi
