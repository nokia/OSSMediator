SHELL := /bin/bash

all:
	@cd OSSMediatorCollector && make all
	@cd ElasticsearchPlugin && make all
	@cp OSSMediatorCollector/bin/storesecret MediatorSetup/.