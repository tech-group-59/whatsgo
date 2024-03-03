build-linux:
	GOOS="linux" GOARCH="amd64" go build -o bin/whatsgo-amd64-linux
build-mac:
	GOOS="darwin" GOARCH="arm64" go build -o bin/whatsgo-arm64-mac
build-windows:
	GOOS="windows" GOARCH="amd64" go build -o bin/whatsgo-amd64-windows
build-all: build-linux build-mac build-windows

run:
	go build && ./whatsgo
