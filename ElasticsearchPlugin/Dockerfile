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

# Install and start elastic search for unit test
RUN mkdir /elasticsearch/ && cd /elasticsearch && \
        wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.0.1-linux-x86_64.tar.gz && \
        tar -xzf elasticsearch-7.0.1-linux-x86_64.tar.gz

# create required directory for go microservice (ElasticsearchPlugin)
RUN mkdir /ElasticsearchPlugin && mkdir -p /ElasticsearchPlugin/src/elasticsearchplugin/

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
ARG UNAME=user
ARG UID=1000
ARG GID=1000
RUN groupadd -f -g $GID $UNAME && useradd -m -u $UID -g $GID -s /bin/bash $UNAME || true && chown -R $UNAME /elasticsearch/elasticsearch-7.0.1 /ElasticsearchPlugin
USER $UNAME

# build the go app (ElasticsearchPlugin)
RUN make build test
