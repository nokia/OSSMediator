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
RUN make build test
