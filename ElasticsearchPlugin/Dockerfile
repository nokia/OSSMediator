# Staging: for building go app (ElasticsearchPlugin)
FROM golang:1.24.2 AS builder

# download dependencies for build
RUN go install github.com/jstemmer/go-junit-report@latest

# copy project directory to be built
COPY . /ElasticsearchPlugin/.

# set the working directory
WORKDIR /ElasticsearchPlugin/

ARG BUILD_CMD="build test"
ARG VERSION="1.1"

#unit test setup
RUN if echo "$BUILD_CMD" | grep -q "test" > /dev/null ; then \
    mkdir /opensearch && cd /opensearch && \
    wget https://artifacts.opensearch.org/releases/core/opensearch/2.19.0/opensearch-min-2.19.0-linux-x64.tar.gz && \
    tar -xzf opensearch-min-2.19.0-linux-x64.tar.gz && \
    export OPENSEARCH_JAVA_HOME=/opensearch/opensearch-2.19.0/jdk && \
    #Creating user to run opensearch \
    mkdir -p /opensearch/opensearch-2.19.0/data/ && \
    groupadd -f -g 1000 opensearch && useradd -m -u 1000 -g 1000 -s /bin/bash opensearch && \
    chown -R opensearch /opensearch/opensearch-2.19.0/jdk /opensearch/opensearch-2.19.0/bin/opensearch /opensearch/opensearch-2.19.0/data/ /opensearch/opensearch-2.19.0/config/ /opensearch/opensearch-2.19.0/logs/; \
    fi

# build the go app (ElasticsearchPlugin)
RUN make $BUILD_CMD VERSION=$VERSION

# staging: base package for app deployment
FROM alpine:3.21.3

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /plugin/bin

# copy compiled binary from the builder stage
COPY --from=builder /ElasticsearchPlugin/bin/elasticsearchplugin /plugin/bin/

WORKDIR /plugin/bin

ENTRYPOINT ["./elasticsearchplugin"]
