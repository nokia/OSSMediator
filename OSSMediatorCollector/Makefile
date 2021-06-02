SHELL := /bin/bash
VERSION = $(shell cat ../VERSION)

all: clean docker_build

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for OSSMediatorCollector......"
	@echo "---------------------------------------------------------------------------------"
	@echo ""
	@echo "Docker Build..."
	@echo "..............."
	@cp ../VERSION .
	@docker build -t ossmediatorcollector:1 . --network host
	@docker create --name ossmediatorcollector ossmediatorcollector:1
	@docker cp ossmediatorcollector:/OSSMediatorCollector/OSSMediatorCollector-unittest-result.txt .
	@docker cp ossmediatorcollector:/OSSMediatorCollector/OSSMediatorCollector-lint-result.txt .
	@docker rm ossmediatorcollector
	@echo "docker build completed."

build:
	@echo Building OSSMediatorCollector
	@cd src/collector && go install -ldflags "-X main.appVersion=$(VERSION)" || (echo "OSSMediatorCollector build failed"; exit 1)
	@echo Running go lint
	@cd src/collector && go vet ./... > ../../OSSMediatorCollector-lint-result.txt
	@echo Build Successful.

build_package:
	@echo Creating OSSMediatorCollector package
	@mkdir -p bin
	@docker create --name ossmediatorcollector ossmediatorcollector:1
	@docker cp ossmediatorcollector:/OSSMediatorCollector/bin/collector ./bin/
	@docker rm ossmediatorcollector
	@mkdir -p package && mkdir -p package/bin && cp -R resources package && cp bin/collector package/bin && chmod 777 package/bin/collector
	@cd package && zip -r OSSMediatorCollector-$(VERSION).zip bin resources && rm -rf bin resources
	@echo Package created at OSSMediatorCollector-$(VERSION).zip

test:
	@echo "Started :OSSMediatorCollector Tests"
	@echo Running Tests
	@cd src/collector && go test ./... -coverprofile=../../coverage.out -v > ../../OSSMediatorCollector-unittest-result.txt || echo "OSSMediatorCollector unit tests failed."
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@echo Generating coverage report
	@gocov convert coverage.out
	@echo "Completed :OSSMediatorCollector Tests"

clean:
	@echo "Started :OSSMediatorCollector CleanUp"
	@rm -rf pkg package bin/collector VERSION OSSMediatorCollector-unittest-result.txt
	@echo "Completed :OSSMediatorCollector CleanUp"
