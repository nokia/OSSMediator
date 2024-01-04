# Staging: for building go app (ElasticsearchPlugin)
FROM golang:1.18 as builder

# download dependencies for build
RUN go install github.com/jstemmer/go-junit-report@latest

# copy project directory to be built
COPY . /ElasticsearchPlugin/.

# set the working directory
WORKDIR /ElasticsearchPlugin/

ARG BUILD_CMD="build test"

#unit test setup
RUN if echo "$BUILD_CMD" | grep -q "test" > /dev/null ; then \
    export JAVA_BIN="/usr/local/jre1.8.0_171/bin/" && \
    export PATH=$PATH:$GOBIN:$JAVA_BIN && \
    mkdir -p /tmp/java && \
    wget -O jre-8u171-linux-x64.tar.gz http://javadl.oracle.com/webapps/download/AutoDL?BundleId=233162_512cd62ec5174c3487ac17c61aaa89e8 && \
    tar -C /usr/local -xzf jre-8u171-linux-x64.tar.gz && \
    mkdir /elasticsearch/ && cd /elasticsearch && \
    wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.0.1-linux-x86_64.tar.gz && \
    tar -xzf elasticsearch-7.0.1-linux-x86_64.tar.gz && \
    # create required directory for ElasticsearchPlugin
    mkdir -p /ElasticsearchPlugin/ /elasticsearch/elasticsearch-7.0.1/data/ && \
    #Creating user to run elastic search
    groupadd -f -g 1000 elasticsearch && useradd -m -u 1000 -g 1000 -s /bin/bash elasticsearch && \
    chown -R elasticsearch /elasticsearch/elasticsearch-7.0.1/bin/elasticsearch /elasticsearch/elasticsearch-7.0.1/data/ /elasticsearch/elasticsearch-7.0.1/config/ /elasticsearch/elasticsearch-7.0.1/logs/; \
    fi

# build the go app (ElasticsearchPlugin)
RUN make $BUILD_CMD

# staging: base package for app deployment
FROM alpine:3.19.0

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /plugin/bin

# copy compiled binary from the builder stage
COPY --from=builder /ElasticsearchPlugin/bin/elasticsearchplugin /plugin/bin/

WORKDIR /plugin/bin

ENTRYPOINT ["./elasticsearchplugin"]
