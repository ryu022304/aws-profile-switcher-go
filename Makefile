linux:
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/awsp-linux ./main.go

win:
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/awsp-windows.exe ./main.go

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/mac-amd64/awsp-mac ./main.go

build:
	@make linux
	@make win
	@make mac
