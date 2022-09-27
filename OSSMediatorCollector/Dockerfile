# Staging: for building go app (OSSMediatorCollector)
FROM golang:1.16 as builder

# create required directory for OSSMediatorCollector
RUN mkdir /OSSMediatorCollector

# copy project directory to be built
COPY . /OSSMediatorCollector/.
COPY ./VERSION ../.

# set the working directory
WORKDIR /OSSMediatorCollector/

# build the go app (OSSMediatorCollector)
RUN make build test build_storesecret

# staging: base package for app deployment
FROM alpine:latest

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /collector/bin

# copy compiled binary from the builder stage
COPY --from=builder /OSSMediatorCollector/bin/collector /collector/bin/

WORKDIR /collector/bin

ENTRYPOINT ["./collector"]

