FROM golang:1.18 AS builder
LABEL maintainer="mingcheng<mingcheng@outook.com>"

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
FROM ubuntu:22.04

RUN sed -i "s@http://.*archive.ubuntu.com@http://repo.huaweicloud.com@g" /etc/apt/sources.list \
    && sed -i "s@http://.*security.ubuntu.com@http://repo.huaweicloud.com@g" /etc/apt/sources.list \
	&& apt -y update && apt -y upgrade \
	&& apt -y install ca-certificates openssl tzdata curl dumb-init \
	&& apt -y autoremove

ENV TZ "Asia/Shanghai"
COPY --from=builder /bin/obsync /bin/obsync
VOLUME /etc/obsync.yaml

ENTRYPOINT ["dumb-init", "/bin/obsync"]
