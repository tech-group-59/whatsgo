build-linux:
	GOOS="linux" GOARCH="amd64" go build -o bin/whatsgo-amd64-linux ./cmd/whatsgo
build-mac:
	GOOS="darwin" GOARCH="arm64" go build -o bin/whatsgo-arm64-mac ./cmd/whatsgo
build-windows:
	GOOS="windows" GOARCH="amd64" go build -o bin/whatsgo-amd64-windows ./cmd/whatsgo.exe
build-all: build-linux build-mac build-windows

run:
	go build -o build/whatsgo ./cmd/whatsgo && ./build/whatsgo --config ./config/config.yaml

run-docker:
	docker build -t whatsgo . && docker run whatsgo
