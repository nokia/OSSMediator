SHELL := /bin/bash
VERSION = $(shell cat ../VERSION)

all: clean docker_build

docker_build:
	@echo "---------------------------------------------------------------------------------"
	@echo "Starting docker build process, for OpenNMSPlugin......"
	@echo "---------------------------------------------------------------------------------"
	@echo ""
	@echo "Docker Build..."
	@echo "..............."
	@cp ../VERSION .
	@docker build -t opennmsplugin:1 . --network host
	@echo "docker build completed."

build:
	@echo Building OpenNMSPlugin
	@cd src/opennmsplugin &&  go install -ldflags "-X main.appVersion=$(VERSION)" || (echo "OpenNMSPlugin build failed"; exit 1)
	@echo Build Successful.

build_package:
	@echo Creating OpenNMSPlugin package
	@mkdir -p bin
	@docker create --name opennmsplugin opennmsplugin:1
	@docker cp opennmsplugin:/OpenNMSPlugin/bin/opennmsplugin ./bin/
	@docker rm opennmsplugin
	@mkdir -p package && mkdir -p package/bin && cp -R resources package && cp bin/opennmsplugin package/bin && chmod 777 package/bin/opennmsplugin
	@cd package && zip -r OpenNMSPlugin-$(VERSION).zip bin resources && rm -rf bin resources
	@echo Package created at package/OpenNMSPlugin-$(VERSION).zip

test:
	@echo "Started :OpenNMSPlugin Tests"
	@echo Running Tests
	@cd src/opennmsplugin && go test ./... -v -coverprofile=../../coverage.out | go-junit-report > ../../unittest-result.xml
	@echo Running go coverage
	@go tool cover -func=coverage.out
	@gocov convert coverage.out
	@echo "Completed :OpenNMSPlugin Tests"

clean:
	@echo "Started :OpenNMSPlugin CleanUp"
	@rm -rf package bin/opennmsplugin VERSION
	@echo "Completed :OpenNMSPlugin CleanUp"
