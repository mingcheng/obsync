PROJECT="obsync"
VERSION=`date +%Y%m%d`

ifneq ("$(wildcard /go)","")
   GOPATH=/go
   GOROOT=/usr/local/go
endif

TARGET=./target
SRC=./cmd/obsync
BINARY=$(TARGET)/obsync

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags="-X main.version=$(VERSION) -X 'main.commit=`git rev-parse HEAD`' -X 'main.date=`date`'"
GO=env $(GO_ENV) $(GOROOT)/bin/go

PACKAGES=`go list ./... | grep -v /vendor/`
GOFILES=`find . -name "*.go" -type f -not -path "./vendor/*"`

build: cmd/obsync
	@$(GO) build $(GO_FLAGS) -o ${BINARY} -tags=jsoniter $(SRC)

fmt:
	@gofmt -s -w ${GOFILES}

list:
	@echo ${PACKAGES}
	@echo ${VETPACKAGES}
	@echo ${GOFILES}

test:
	@go test -cpu=1,2,4 -v -tags integration ./...

install: build
	@$(GO) install $(GO_FLAGS) -tags=jsoniter $(SRC)

docker_image: clean
	@docker build -f ./Dockerfile -t obsync:$(VERSION) .

clean:
	@$(GO) clean ./...

.PHONY: fmt install test clean target docker_image
