SHELL := /bin/bash
VERSION = $(shell cat ../VERSION)

all: clean docker_build

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for ElasticsearchPlugin......"
	@echo "---------------------------------------------------------------------------------"
	@echo ""
	@echo "Docker Build..."
	@echo "..............."
	@cp ../VERSION .
	@docker build -t elasticsearchplugin:1 . --network host
	@docker create --name elasticsearchplugin elasticsearchplugin:1
	@docker cp elasticsearchplugin:/ElasticsearchPlugin/elasticsearchplugin-unittest-result.txt .
	@docker rm elasticsearchplugin
	@echo "docker build completed."

build:
	@echo Building ElasticsearchPlugin
	@cd src/elasticsearchplugin &&  go install -ldflags "-X main.appVersion=$(VERSION)" || (echo "ElasticsearchPlugin build failed"; exit 1)
	@echo Build Successful.

build_package:
	@echo Creating ElasticsearchPlugin package
	@mkdir -p bin
	@docker create --name elasticsearchplugin elasticsearchplugin:1
	@docker cp elasticsearchplugin:/ElasticsearchPlugin/bin/elasticsearchplugin ./bin/
	@docker rm elasticsearchplugin
	@mkdir -p package && mkdir -p package/bin && cp -R resources package && cp bin/elasticsearchplugin package/bin && chmod 777 package/bin/elasticsearchplugin
	@cd package && zip -r ElasticsearchPlugin-$(VERSION).zip bin resources && rm -rf bin resources
	@echo Package created at package/ElasticsearchPlugin-$(VERSION).zip

test:
	@echo "Started :ElasticsearchPlugin Tests"
	@echo "---> Starting ElasticSearch"
	@cd /elasticsearch/elasticsearch-7.0.1/bin && nohup bash -c "./elasticsearch &" && sleep 20
	@echo Running Tests
	@cd src/elasticsearchplugin && go test -coverpkg=.,./config,./util ./... -v -coverprofile=../../coverage.out > ../../elasticsearchplugin-unittest-result.txt || echo "ElasticsearchPlugin unit tests failed."
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@gocov convert coverage.out
	@echo "Completed :ElasticsearchPlugin Tests"

clean:
	@echo "Started :ElasticsearchPlugin CleanUp"
	@rm -rf package bin/elasticsearchplugin VERSION elasticsearchplugin-unittest-result.txt
	@echo "Completed :ElasticsearchPlugin CleanUp"
