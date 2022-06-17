@IF "%1" == "" (
	ECHO "Version missing"
) ELSE (
	@echo building Android
	fyne package -os android -appVersion %1
	@echo building Windows
	fyne package -os windows -appVersion %1
	@echo building Linux
	fyne package -os linux -appVersion %1
)