@IF "%1" == "" (
	@ECHO "Version X.X.X missing, with optional Build X"
) ELSE (
        @SET BLD=%2
        @IF "%BLD%" == "" FOR /F "tokens=2" %%i IN ('date /t') DO SET bld=%%i
	@ECHO building Windows %1
	# nomsgpack for gin
	go build -v -tags=nomsgpack -ldflags="-X main.Ver=%1" -ldflags="-X main.Bld=$bld" -o EZComm_cmd.exe
)
