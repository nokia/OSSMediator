SHELL := /bin/bash

all: clean docker_build_all copy_binary

docker_build_all:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build and test process, for OSSMediatorCollector......"
	@echo "---------------------------------------------------------------------------------"
	@docker build -t ossmediatorcollector:$(shell cat ../VERSION) . --network host
	@echo "docker build completed."

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for OSSMediatorCollector......"
	@echo "---------------------------------------------------------------------------------"
	@docker build -t ossmediatorcollector:$(shell cat ../VERSION) . --network host --build-arg BUILD_CMD="build build_storesecret"
	@echo "docker build completed."

copy_binary:
	@docker create --name ossmediatorcollector ossmediatorcollector:$(shell cat ../VERSION)
	@mkdir -p bin
	@docker cp ossmediatorcollector:/collector/bin/. ./bin/
	@docker rm ossmediatorcollector

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
