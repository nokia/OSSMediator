SHELL := /bin/bash
VERSION = $(shell cat ../VERSION)

all: clean docker_build build_package

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for OSSMediatorCollector......"
	@echo "---------------------------------------------------------------------------------"
	@echo ""
	@echo "Docker Build..."
	@echo "..............."
	@cp ../VERSION .
	@docker build -t ossmediatorcollector:1 . --network host
	@echo "docker build completed."

build:
	@echo Building OSSMediatorCollector
	@cd src/collector && go install -ldflags "-X main.appVersion=$(VERSION)" || (echo "OSSMediatorCollector build failed"; exit 1)
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
	@echo Running go metalinter
	@-cd src/collector && gometalinter.v2 ./... --checkstyle --vendor --deadline=100s > ../../report.xml || echo "Error: Check report.xml for lint errors."
	@echo Running Tests
	@cd src/collector && go test ./... -coverprofile=../../coverage.out -v | go-junit-report > ../../unittest-result.xml
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@echo Generating coverage report
	@gocov convert coverage.out
	@echo "Completed :OSSMediatorCollector Tests"

clean:
	@echo "Started :OSSMediatorCollector CleanUp"
	@rm -rf pkg package bin/collector VERSION
	@echo "Completed :OSSMediatorCollector CleanUp"
