FROM golang:1.21 AS builder

ENV GOPATH /go
ENV GOROOT /usr/local/go

ENV PACKAGE github.com/mingcheng/obsync
ENV GOPROXY https://goproxy.cn,direct
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN make clean build && ./obsync -h && mv ./obsync /bin/obsync

# Stage2
FROM debian:stable
LABEL maintainer="mingcheng<mingcheng@outook.com>"

RUN apt -y update && apt -y install ca-certificates openssl tzdata curl dumb-init

ENV TZ "Asia/Shanghai"
COPY --from=builder /bin/obsync /bin/obsync
VOLUME /etc/obsync.yaml

ENTRYPOINT ["dumb-init", "/bin/obsync"]
