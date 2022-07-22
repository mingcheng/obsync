PROJECT="obsync"
VERSION=$(shell date +%Y%m%d)
COMMIT_HASH=$(shell git rev-parse --short HEAD)

TARGET=./target
SRC=./cmd/obsync
BINARY=./obsync

GO_FLAGS=-ldflags=" \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT_HASH)' \
	-X 'main.date=$(shell date)'"
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
