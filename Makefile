GOCOMPILER = go
SOURCE = main.go

linux: $(SOURCE)
	$(GOCOMPILER) build -o voyagemaster $(SOURCE)

win: $(SOURCE)
	GOOS=windows GOARCH=amd64 $(GOCOMPILER) build -o VoyageMaster.exe $(SOURCE)
