# Staging: for building go app (OSSMediatorCollector)
FROM golang:1.23.4 AS builder

# download dependencies for build
RUN go install github.com/jstemmer/go-junit-report@latest

# create required directory for OSSMediatorCollector
RUN mkdir /OSSMediatorCollector

# copy project directory to be built
COPY . /OSSMediatorCollector/.

# set the working directory
WORKDIR /OSSMediatorCollector/

ARG BUILD_CMD="build test build_storesecret"
ARG VERSION="1.1"
# build the go app (OSSMediatorCollector)
RUN make $BUILD_CMD VERSION=$VERSION

# staging: base package for app deployment
FROM alpine:3.21.2

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /collector/bin

# copy compiled binary from the builder stage
COPY --from=builder /OSSMediatorCollector/bin/. /collector/bin/

WORKDIR /collector/bin

ENTRYPOINT ["./collector"]

