FROM golang:1.12.6 AS builder
LABEL maintainer="Ming Cheng"

# Using 163 mirror for Debian Strech
#RUN sed -i 's/deb.debian.org/mirrors.163.com/g' /etc/apt/sources.list
#RUN apt-get update

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PACKAGE github.com/mingcheng/obsync.go
ENV GOPROXY https://goproxy.cn
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}
ENV TARGET_DIR ${BUILD_DIR}/target

# Print go version
RUN echo "GOROOT is ${GOROOT}"
RUN echo "GOPATH is ${GOPATH}"
RUN ${GOROOT}/bin/go version

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN make clean && \
  env GO111MODULE=on GO15VENDOREXPERIMENT=1 make build && \
  ${TARGET_DIR}/obsync -v && \
  mv ${TARGET_DIR}/obsync /usr/bin/obsync

# Stage2
FROM alpine:3.10

# @from https://mirrors.ustc.edu.cn/help/alpine.html
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

COPY --from=builder /usr/bin/obsync /bin/obsync
ENTRYPOINT ["/bin/obsync"]
