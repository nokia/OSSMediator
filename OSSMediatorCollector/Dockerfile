# Staging: for building go app (OSSMediatorCollector)
FROM registry.prod.srv.da.nsn-rdnet.net/baseimage/golang:1.18 as builder
#FROM golang:1.18 as builder

# download dependencies for build
RUN go install github.com/jstemmer/go-junit-report@latest

# create required directory for OSSMediatorCollector
RUN mkdir /OSSMediatorCollector

# copy project directory to be built
COPY . /OSSMediatorCollector/.

# set the working directory
WORKDIR /OSSMediatorCollector/

# build the go app (OSSMediatorCollector)
RUN make build test build_storesecret

# staging: base package for app deployment
FROM registry.prod.srv.da.nsn-rdnet.net/baseimage/alpine:3.18.2
#FROM alpine:3.18.2

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /collector/bin

# copy compiled binary from the builder stage
COPY --from=builder /OSSMediatorCollector/bin/collector /collector/bin/

WORKDIR /collector/bin

ENTRYPOINT ["./collector"]

