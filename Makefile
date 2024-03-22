build-linux:
	CPATH="/opt/homebrew/include" LIBRARY_PATH="/opt/homebrew/lib" GOOS="linux" GOARCH="amd64" go build -o build/whatsgo-amd64-linux ./cmd/whatsgo
build-mac:
	CPATH="/opt/homebrew/include" LIBRARY_PATH="/opt/homebrew/lib" GOOS="darwin" GOARCH="arm64" go build -o build/whatsgo-arm64-mac ./cmd/whatsgo
build-windows:
	CPATH="/opt/homebrew/include" LIBRARY_PATH="/opt/homebrew/lib" GOOS="windows" GOARCH="amd64" go build -o build/whatsgo-amd64-windows ./cmd/whatsgo.exe
build-all: build-linux build-mac build-windows

run:
	go build -o build/whatsgo ./cmd/whatsgo && ./build/whatsgo --config ./config/config.yaml

run-docker:
	docker build -t whatsgo . && docker run whatsgo

dcb:
	docker compose build

dcr:
	docker compose run --rm app

dcu:
	docker compose up
