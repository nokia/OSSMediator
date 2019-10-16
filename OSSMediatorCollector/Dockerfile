# Staging: for building go app (OSSMediatorCollector)
FROM golang:latest as builder

# set environment variable
ENV GOPATH="/OSSMediatorCollector"
ENV GOBIN="/OSSMediatorCollector/bin"
ENV PATH=$PATH:$GOBIN

RUN apt-get update && apt-get -y install zip curl && apt-get clean

# create required directory for go microservice (OSSMediatorCollector)
RUN mkdir /OSSMediatorCollector && mkdir -p /OSSMediatorCollector/src/collector/

# set the working directory
WORKDIR /OSSMediatorCollector/

#Install dep
RUN go get github.com/golang/dep/cmd/dep

# Install plugins for unit tests
RUN go get github.com/jstemmer/go-junit-report && \
        go get github.com/axw/gocov/gocov

# download vendor dependencies; if Gopkg.* files is not modified, it will use docker Cache
COPY ./src/collector/Gopkg.* ./src/collector/
RUN cd ./src/collector/ && dep ensure -vendor-only

# copy project directory to be built
COPY ./Makefile .
COPY ./VERSION ../.
COPY ./src/collector ./src/collector/
COPY ./resources ./resources/

# build the go app (OSSMediatorCollector)
RUN make build test
