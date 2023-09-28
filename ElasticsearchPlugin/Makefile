SHELL := /bin/bash

all: clean docker_build_all copy_binary

docker_build_all:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build and test process, for ElasticsearchPlugin......"
	@echo "---------------------------------------------------------------------------------"
	@docker build -t elasticsearchplugin:$(shell cat ../VERSION) . --network host
	@echo "docker build completed."

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for ElasticsearchPlugin......"
	@echo "---------------------------------------------------------------------------------"
	@docker build -t elasticsearchplugin:$(shell cat ../VERSION) . --network host --build-arg BUILD_CMD="build"
	@echo "docker build completed."

copy_binary:
	@docker create --name elasticsearchplugin elasticsearchplugin:$(shell cat ../VERSION)
	@mkdir -p bin
	@docker cp elasticsearchplugin:/plugin/bin/. ./bin/
	@docker rm elasticsearchplugin

build:
	@echo Building ElasticsearchPlugin
	@go mod download && go build -o bin/elasticsearchplugin cmd/main.go || (echo "ElasticsearchPlugin build failed"; exit 1)
	@echo Running go lint
	@go vet ./... > lint-report.xml
	@echo Build Successful.

build_package:
	@echo Creating ElasticsearchPlugin package
	@mkdir -p package && mkdir -p package/bin && cp -R resources package && cp bin/elasticsearchplugin package/bin && chmod 777 package/bin/elasticsearchplugin
	@cd package && zip -r ElasticsearchPlugin-$(shell cat ../VERSION).zip bin resources && rm -rf bin resources
	@echo Package created at package/ElasticsearchPlugin-$(shell cat ../VERSION).zip

test:
	@echo "Started :ElasticsearchPlugin Tests"
	@echo "---> Starting ElasticSearch"
	@runuser -l elasticsearch -c 'cd /elasticsearch/elasticsearch-7.0.1/bin && nohup bash -c "./elasticsearch -E http.port=9299 -E transport.tcp.port=9399 &" && sleep 20'
	@echo Running Tests
	@go test ./... -v -coverprofile=coverage.out | tee plugin_test.log | go-junit-report > unittest-result.xml || echo "ElasticsearchPlugin unit tests failed."
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@echo "Completed :ElasticsearchPlugin Tests"

clean:
	@echo "Started :ElasticsearchPlugin CleanUp"
	@rm -rf bin package VERSION
	@echo "Completed :ElasticsearchPlugin CleanUp"
