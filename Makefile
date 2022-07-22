PROJECT="obsync"
VERSION=$(shell date +%Y%m%d)

TARGET=./target
SRC=./cmd/obsync
BINARY=./obsync

GO_FLAGS=-ldflags="-X main.version=$(VERSION) -X 'main.commit=`git rev-parse HEAD`' -X 'main.date=`date`'"
GO=$(shell which go)

build: cmd/obsync
	@$(GO) build $(GO_FLAGS) -o ${BINARY} $(SRC)

test:
	@go test -v ./cmd/... ./bucket/...

docker_image:
	@docker-compose build

clean:
	@$(GO) clean ./...
	@rm -f ${BINARY}

.PHONY: fmt install test clean target docker_image dist release
