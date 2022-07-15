@IF "%1" == "" (
	ECHO "Version missing"
) ELSE (
	@echo building Linux %1
	cp FyneApp.toml FyneApp.bak
	fyne-cross linux -arch=amd64 -name ezcomm -app-version %1
	cp FyneApp.bak FyneApp.toml
)