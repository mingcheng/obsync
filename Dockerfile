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
FROM debian:stable

ENV TZ "Asia/Shanghai"
RUN sed -i 's/deb.debian.org/mirror.nju.edu.cn/g' /etc/apt/sources.list \
	&& sed -i 's/security.debian.org/mirror.nju.edu.cn/g' /etc/apt/sources.list \
	&& echo "Asia/Shanghai" > /etc/timezone \
	&& apt -y update \
	&& apt -y upgrade \
	&& apt -y install ca-certificates openssl tzdata curl dumb-init \
	&& apt -y autoremove

COPY --from=builder /bin/obsync /bin/obsync
VOLUME /etc/obsync.yaml

ENTRYPOINT ["dumb-init", "/bin/obsync"]
