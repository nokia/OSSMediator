SHELL := /bin/bash
VERSION = $(shell cat VERSION)

all:
	@cd OSSMediatorCollector && make all VERSION=$(VERSION)
	@cd ElasticsearchPlugin && make all VERSION=$(VERSION)
	@cp OSSMediatorCollector/bin/storesecret MediatorSetup/.

docker_build:
	@cd OSSMediatorCollector && make docker_build copy_binary VERSION=$(VERSION)
	@cd ElasticsearchPlugin && make docker_build copy_binary VERSION=$(VERSION)
	@cp OSSMediatorCollector/bin/storesecret MediatorSetup/.
