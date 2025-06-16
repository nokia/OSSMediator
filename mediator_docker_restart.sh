#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Restart OSS Mediator Container"
	echo "Options:"
	echo "  -h, --help, help               Show help"
	echo "  -a, --all, all                 Restart ALL containers"
	echo "  -os, --opensearch, opensearch  Restart OpenSearch data"
	echo "  -g, --grafana, grafana         Restart Grafana"
	echo "  -p, --plugin, plugin           Restart ElasticSearch plugin"
	echo "  -dc1, --dc1, dc1               Restart DC1 Collector"
	echo "  -dc4, --dc4, dc4               Restart DC4 Collector"
	echo "  -dc5, --dc5, dc5               Restart DC5 Collector"
	echo
}

# Function to restart all containers
restartContainers(){
	#docker restart $(docker ps -a -q)
	$(pwd)/mediator_docker_stop.sh all
	$(pwd)/mediator_docker_start.sh all
}

# Function to restart dc1 collector
restartDC1(){
	echo "Restarting DC1 Collector..."
	docker restart ndac_oss_collector_dc1
}

# Function to restart dc4 collector
restartDC4(){
	echo "Restarting DC4 Collector.."
	docker restart ndac_oss_collector_dc4
}

# Function to restart dc5 collector
restartDC5(){
	echo "Restarting DC5 Collector..."
	docker restart ndac_oss_collector_dc5
}

# Function to restart ES Plugin 
restartElasticSearchPlugin(){
	echo "Restarting ElasticSearch Plugin..."
	docker restart ndac_oss_elasticsearchplugin
}


# Function to restart OpenSearch and storage
restartOpenSearch(){
	echo "Restarting OpenSearch..."
	docker restart ndac_oss_opensearch
}

# Function to restart Grafana and storage
restartGrafana(){
	echo "Restarting Grafana..."
	docker restart ndac_oss_grafana
}


### MAIN start HERE ######

# Check if there are any containers running
if [ -z "$(docker ps -aq)" ]; then
	echo "There are no containers deployed!"
	#show_help
	exit 1

fi

# Check if there are any arguments
if [ $# -eq 0 ]; then
    echo "Invalid command. Use -h or --help for usage information."
else
	# Handle arguments
    # Wipe specific component
	    case $1 in
        -h|--help|help)
            show_help
            ;;
		-a|--all|all)
            restartContainers
            ;;
		-os|--opensearch|os)
			restartOpenSearch
			;;
		-p|--plugin|plugin)
			restartElasticSearchPlugin
			;;
		-g|--grafana|grafana)	
			restartGrafana
			;;	
		-dc1|--dc1|dc1)
			restartDC1
			;;
		-dc4|--dc4|dc4)
			restartDC4
			;;
		-dc5|--dc5|dc5)
			restartDC5
			;;
		*)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi
