@REM @grep "func Test" *.go|gawk "{print $2}"|gawk -F "(" "{print $1}"
go test -list .*
@ECHO go test -v -verbose=3 -run
