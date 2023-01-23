all: build-windows build-linux build-mac sha256sums

export CGO_ENABLED=0
NAME=markf

build:
	go build -o bin/$(NAME)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(NAME)-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -o bin/$(NAME)-windows-i386.exe
	GOOS=windows GOARCH=arm64 go build -o bin/$(NAME)-windows-arm64

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(NAME)-linux-amd64
	GOOS=linux GOARCH=386 go build -o bin/$(NAME)-linux-i386
	GOOS=linux GOARCH=arm64 go build -o bin/$(NAME)-linux-arm64

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(NAME)-mac-amd64
	GOOS=darwin GOARCH=arm64 go build -o bin/$(NAME)-mac-arm64

sha256sums:
	cd bin && \
	for f in *; do sha256sum $$f > $$f.sha256; done && \
	sha256sum -c *.sha256

clean:
	rm -rf bin/*
	go clean
