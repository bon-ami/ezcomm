#!/bin/bash
if [ $# -lt 1 ]; then
	echo "Version X.X.X missing"
else
	echo building Linux $1
	go build -v -ldflags="-X main.Ver=$1" -o EZComm_cmd
fi
