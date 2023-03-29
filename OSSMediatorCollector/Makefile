SHELL := /bin/bash

all: clean docker_build

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for OSSMediatorCollector......"
	@echo "---------------------------------------------------------------------------------"
	@echo ""
	@echo "Docker Build..."
	@echo "..............."
	@docker build -t ossmediatorcollector:$(shell cat ../VERSION) . --network host
	@docker build --target builder -t collector-builder:1 . --network host
	@docker create --name collector-builder collector-builder:1
	@mkdir -p bin
	@docker cp collector-builder:/OSSMediatorCollector/bin/collector ./bin/
	@docker cp collector-builder:/OSSMediatorCollector/bin/storesecret ./bin/
	@docker rm collector-builder
	@echo "docker build completed."

build:
	@echo Building OSSMediatorCollector
	@go mod download && go build -o bin/collector cmd/collector.go || (echo "OSSMediatorCollector build failed"; exit 1)
	@echo Running go lint
	@go vet ./... > lint-report.xml
	@echo Build Successful.

build_storesecret:
	@echo Building storesecret
	@cd storesecret && go mod download && go build -o ../bin/storesecret storesecret.go || (echo "storesecret build failed"; exit 1)
	@echo storesecret build successful.

build_package:
	@echo Creating OSSMediatorCollector package
	@mkdir -p package && mkdir -p package/bin && cp -R resources package && cp bin/* package/bin/ && chmod 777 package/bin/*
	@cd package && zip -r OSSMediatorCollector-$(shell cat ../VERSION).zip bin resources && rm -rf bin resources
	@echo Package created at package/OSSMediatorCollector-$(shell cat ../VERSION).zip

test:
	@echo "Started :OSSMediatorCollector Tests"
	@echo Running Tests
	@go test ./... -coverprofile=coverage.out -v | tee collector_test.log | go-junit-report > unittest-result.xml || echo "OSSMediatorCollector unit tests failed."
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@echo "Completed :OSSMediatorCollector Tests"

clean:
	@echo "Started :OSSMediatorCollector CleanUp"
	@rm -rf bin/ package VERSION
	@echo "Completed :OSSMediatorCollector CleanUp"
