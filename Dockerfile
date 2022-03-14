FROM golang:1.17 AS builder
LABEL maintainer="mingcheng<mingcheng@outook.com>"

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PACKAGE github.com/mingcheng/obsync.go
ENV GOPROXY https://goproxy.cn,direct
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}
ENV TARGET_DIR ${BUILD_DIR}/target

# Print go version
RUN echo "GOROOT is ${GOROOT}"
RUN echo "GOPATH is ${GOPATH}"
RUN ${GOROOT}/bin/go version

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN make clean build && ${TARGET_DIR}/obsync -v && mv ${TARGET_DIR}/obsync /usr/bin/obsync

# Stage2
FROM debian:bullseye

ENV TZ "Asia/Shanghai"
RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& echo "Asia/Shanghai" > /etc/timezone \
	&& apt -y update \
	&& apt -y upgrade \
	&& apt -y install ca-certificates openssl tzdata curl dumb-init \
	&& apt -y autoremove

COPY --from=builder /usr/bin/obsync /bin/obsync
VOLUME /etc/obsync.json

ENTRYPOINT ["dumb-init", "/bin/obsync"]
