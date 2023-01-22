all: build-windows build-linux build-mac

export CGO_ENABLED=0
NAME=markf

build-windows:
	GOOS=windows GOARCH=386 go build -o bin/$(NAME)-windows-i32.exe
	GOOS=windows GOARCH=amd64 go build -o bin/$(NAME)-windows-amd64.exe

build-linux:
	GOOS=linux GOARCH=386 go build -o bin/$(NAME)-linux-i32
	GOOS=linux GOARCH=amd64 go build -o bin/$(NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o bin/$(NAME)-linux-arm64

build-mac:
	GOOS=darwin GOARCH=arm64 go build -o bin/$(NAME)-mac-arm64
	GOOS=darwin GOARCH=amd64 go build -o bin/$(NAME)-mac-amd64

clean:
	rm -rf bin/*
	go clean