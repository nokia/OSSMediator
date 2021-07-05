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
	@docker cp elasticsearchplugin:/ElasticsearchPlugin/elasticsearchplugin-lint-result.txt .
	@docker cp elasticsearchplugin:/ElasticsearchPlugin/coverage.html .
	@docker rm elasticsearchplugin
	@echo "docker build completed."

build:
	@echo Building ElasticsearchPlugin
	@go mod download && go build -ldflags "-X main.appVersion=$(VERSION)" -o bin/elasticsearchplugin cmd/main.go || (echo "ElasticsearchPlugin build failed"; exit 1)
	@echo Running go lint
	@go vet ./... > elasticsearchplugin-lint-result.txt
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
	@runuser -l elasticsearch -c 'cd /elasticsearch/elasticsearch-7.0.1/bin && nohup bash -c "./elasticsearch -E http.port=9299 -E transport.tcp.port=9399 &" && sleep 20'
	@echo Running Tests
	@go test ./... -v -coverprofile=coverage.out > elasticsearchplugin-unittest-result.txt || echo "ElasticsearchPlugin unit tests failed."
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Completed :ElasticsearchPlugin Tests"

clean:
	@echo "Started :ElasticsearchPlugin CleanUp"
	@rm -rf package bin/elasticsearchplugin VERSION elasticsearchplugin-unittest-result.txt elasticsearchplugin-lint-result.txt
	@echo "Completed :ElasticsearchPlugin CleanUp"
