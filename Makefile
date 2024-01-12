SHELL := /bin/bash

all:
	@cd OSSMediatorCollector && make all
	@cd ElasticsearchPlugin && make all
	@cp OSSMediatorCollector/bin/storesecret MediatorSetup/.

docker_build:
	@cd OSSMediatorCollector && make docker_build copy_binary
	@cd ElasticsearchPlugin && make docker_build copy_binary
	@cp OSSMediatorCollector/bin/storesecret MediatorSetup/.
