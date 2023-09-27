SOURCE_PATH = /go/src/net-capture/
CONTAINER_AMD=net-capture-amd64
CONTAINER_ARM=net-capture-arm64
BIN_NAME = ./bin/net-capture
LDFLAGS = -ldflags "-extldflags \"-static\""

build-all: build-bin-linux-amd64 build-bin-linux-arm64 build-bin-mac-amd64 build-bin-mac-arm64 build-bin-windows-amd64

.PHONY: vendor
vendor:
	go mod vendor

build-bin-linux-amd64: vendor
	docker run --platform linux/amd64 --rm -v `pwd`:$(SOURCE_PATH) -t --env GOOS=linux --env GOARCH=amd64 -i $(CONTAINER_AMD) go build -mod=vendor -o $(BIN_NAME)_linux-amd64 -tags netgo $(LDFLAGS) ./cmd

build-bin-linux-arm64: vendor
	docker run --platform linux/arm64 --rm -v `pwd`:$(SOURCE_PATH) -t --env GOOS=linux --env GOARCH=arm64 -i $(CONTAINER_ARM) go build -mod=vendor -o $(BIN_NAME)_linux-arm64 -tags netgo $(LDFLAGS) ./cmd

build-bin-mac-amd64: vendor
	GOOS=darwin go build -mod=vendor -o $(BIN_NAME)_darwin-amd64 ./cmd

build-bin-mac-arm64: vendor
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -mod=vendor -o $(BIN_NAME)_darwin-arm64 ./cmd

build-bin-windows-amd64: vendor
	docker run -it --rm -v `pwd`:$(SOURCE_PATH) -w $(SOURCE_PATH) -e CGO_ENABLED=1 docker.elastic.co/beats-dev/golang-crossbuild:1.20-main --build-cmd "make build" -p "windows/amd64" .
	mv $(BIN_NAME) "$(BIN_NAME)_amd64.exe"

build:
	go build -mod=vendor -o $(BIN_NAME) $(LDFLAGS) ./cmd

build-container: build-container-amd64 build-container-arm64

build-container-amd64:
	docker buildx build --platform linux/amd64 -t $(CONTAINER_AMD) -f Dockerfile .

build-container-arm64:
	docker buildx build --platform linux/arm64 -t $(CONTAINER_ARM) -f Dockerfile .
