@IF "%1" == "" (
	@ECHO "Version missing"
) ELSE (
	@ECHO building Windows %1
	go build -v -ldflags="-X main.Ver=%1" -o EZComm_%1_cmd.exe
)
