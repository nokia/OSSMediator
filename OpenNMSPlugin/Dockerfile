# Staging: for building go app (OpenNMSPlugin)
FROM golang:latest as builder

# set environment variable
ENV GOPATH="/OpenNMSPlugin"
ENV GOBIN="/OpenNMSPlugin/bin"
ENV PATH=$PATH:$GOBIN

RUN apt-get update && apt-get -y install zip curl && apt-get clean

# create required directory for go microservice (OpenNMSPlugin)
RUN mkdir /OpenNMSPlugin && mkdir -p /OpenNMSPlugin/src/opennmsplugin/

# set the working directory
WORKDIR /OpenNMSPlugin/

#Install dep
RUN go get github.com/golang/dep/cmd/dep

# Install gometalinter
RUN go get gopkg.in/alecthomas/gometalinter.v2 && gometalinter.v2 --install

# Install plugins for unit tests
RUN go get github.com/jstemmer/go-junit-report && \
        go get github.com/axw/gocov/gocov

# download vendor dependencies; if Gopkg.* files is not modified, it will use docker Cache
COPY ./src/opennmsplugin/Gopkg.* ./src/opennmsplugin/
RUN cd ./src/opennmsplugin/ && dep ensure -vendor-only

# copy project directory to be built
COPY ./Makefile .
COPY ./VERSION ../.
COPY ./src/opennmsplugin ./src/opennmsplugin/
COPY ./resources ./resources/

# build the go app (OpenNMSPlugin)
RUN make build test
