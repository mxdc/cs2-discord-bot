.PHONY: build build-linux build-rpi build-mac-intel build-mac-arm build-windows clean tidy

BINARY_NAME=cs2-discord-bot
MAIN_PATH=./main.go

build:
	go mod tidy
	make build-all

build-linux:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-linux-amd64"

build-rpi:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=arm \
	GOARM=7 \
	go build -o $(BINARY_NAME)-linux-armv7 $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-linux-armv7"

build-mac-intel:
	CGO_ENABLED=0 \
	GOOS=darwin \
	GOARCH=amd64 \
	go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-darwin-amd64"

build-mac-arm:
	CGO_ENABLED=0 \
	GOOS=darwin \
	GOARCH=arm64 \
	go build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-darwin-arm64"

build-windows:
	CGO_ENABLED=0 \
	GOOS=windows \
	GOARCH=amd64 \
	go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Built: $(BINARY_NAME)-windows-amd64.exe"

build-all: build-linux build-rpi build-mac-intel build-mac-arm build-windows

clean:
	rm -rf \
		bin/ \
		go.sum \
		$(BINARY_NAME) \
		$(BINARY_NAME)-linux-amd64 \
		$(BINARY_NAME)-linux-armv7 \
		$(BINARY_NAME)-linux-arm64 \
		$(BINARY_NAME)-darwin-amd64 \
		$(BINARY_NAME)-darwin-arm64 \
		$(BINARY_NAME)-windows-amd64.exe

tidy:
	go mod tidy

.DEFAULT_GOAL := build
