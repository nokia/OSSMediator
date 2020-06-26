# Staging: for building go app (ElasticsearchPlugin)
FROM golang:1.11 as builder

# set environment variable
ENV GOPATH="/ElasticsearchPlugin"
ENV GOBIN="/ElasticsearchPlugin/bin"
ENV JAVA_BIN="/usr/local/jre1.8.0_171/bin/"
ENV PATH=$PATH:$GOBIN:$JAVA_BIN

# Install java for running elasticsearch for go test
RUN mkdir -p /tmp/java && \
        wget -O jre-8u171-linux-x64.tar.gz http://javadl.oracle.com/webapps/download/AutoDL?BundleId=233162_512cd62ec5174c3487ac17c61aaa89e8 && \
        tar -C /usr/local -xzf jre-8u171-linux-x64.tar.gz

# Install elasticsearch for unit test
RUN mkdir /elasticsearch/ && cd /elasticsearch && \
        wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.0.1-linux-x86_64.tar.gz && \
        tar -xzf elasticsearch-7.0.1-linux-x86_64.tar.gz

# create required directory for go microservice (ElasticsearchPlugin)
RUN mkdir -p /ElasticsearchPlugin/src/elasticsearchplugin/ /elasticsearch/elasticsearch-7.0.1/data/

# set the working directory
WORKDIR /ElasticsearchPlugin/

#Install dep
RUN go get github.com/golang/dep/cmd/dep

# Install plugins for unit tests
RUN go get github.com/axw/gocov/gocov

# download vendor dependencies; if Gopkg.* files is not modified, it will use docker Cache
COPY ./src/elasticsearchplugin/Gopkg.* ./src/elasticsearchplugin/
RUN cd ./src/elasticsearchplugin/ && dep ensure -vendor-only

# copy project directory to be built
COPY ./Makefile .
COPY ./VERSION ../.
COPY ./src/elasticsearchplugin ./src/elasticsearchplugin/
COPY ./resources ./resources/

#Creating user to run elastic search
RUN groupadd -f -g 1000 elasticsearch && useradd -m -u 1000 -g 1000 -s /bin/bash elasticsearch
RUN chown -R elasticsearch /elasticsearch/elasticsearch-7.0.1/bin/elasticsearch \
        /elasticsearch/elasticsearch-7.0.1/data/ \
        /elasticsearch/elasticsearch-7.0.1/config/ \
        /elasticsearch/elasticsearch-7.0.1/logs/

# build the go app (ElasticsearchPlugin)
RUN make build test
