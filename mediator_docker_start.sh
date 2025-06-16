#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Start OSS Mediator Container"
	echo "Options:"
	echo "  -h, --help, help               Show help"
	echo "  -a, --all, all                 Start ALL containers"
	echo "  -os, --opensearch, opensearch  Start OpenSearch data"
	echo "  -g, --grafana, grafana         Start Grafana"
	echo "  -p, --plugin, plugin           Start ElasticSearch plugin"
	echo "  -dc1, --dc1, dc1               Start DC1 Collector"
	echo "  -dc4, --dc4, dc4               Start DC4 Collector"
	echo "  -dc5, --dc5, dc5               Start DC5 Collector"
	echo
}

# Function to start all containers
startContainers(){
	#docker start $(docker ps -a -q)
	startOpenSearch
	startGrafana
	startDC1 #;sleep 5
	startDC4 #;sleep 5
	startDC5 #;sleep 15
	startElasticSearchPlugin
}

# Function to start dc1 collector
startDC1(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc1"; then
		echo "DC1 Collector is running."
	else
		echo "Starting DC1 Collector..."
		docker start ndac_oss_collector_dc1
	fi
}

# Function to start dc4 collector
startDC4(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc4"; then
		echo "DC4 Collector is running."
	else
		echo "Starting DC4 Collector..."
		docker start ndac_oss_collector_dc4
	fi
}

# Function to start dc5 collector
startDC5(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc5"; then
		echo "DC5 Collector is running."
	else
		echo "Starting DC5 Collector..."
		docker start ndac_oss_collector_dc5
	fi
}

# Function to start ES Plugin 
startElasticSearchPlugin(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_elasticsearchplugin"; then
		echo "ElasticSearch Plugin is running."
	else
		echo "Starting ElasticSearch Plugin ..."
		docker start ndac_oss_elasticsearchplugin
	fi
}


# Function to start OpenSearch
startOpenSearch(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_opensearch"; then
		echo "OpenSearch is running."
	else
		echo "Starting OpenSearch..."
		docker start ndac_oss_opensearch
	fi
}

# Function to start Grafana
startGrafana(){

	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_grafana"; then
		echo "Grafana is running."
	else
		echo "Starting Grafana..."
		docker start ndac_oss_grafana
	fi
}


### MAIN STARTS HERE ######

# Check if there are any containers running
if [ -z "$(docker ps -aq)" ]; then
	echo "There are no containers deployed!"
	echo "Do you want to start up OSS Mediator? (y/n)"
	read -r response

	if [[ "$response" == "y" || "$response" == "Y" ]]; then
		$(pwd)/mediator_docker_startup.sh
		exit 0
	elif [[ "$response" == "n" || "$response" == "N" ]]; then
    		exit 0
	else
    		echo "Invalid input. Please enter 'y' or 'n'."
	fi

	#show_help
	#exit 1

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
            startContainers
            ;;
		-os|--opensearch|os)
			startOpenSearch
			;;
		-p|--plugin|plugin)
			startElasticSearchPlugin
			;;
		-g|--grafana|grafana)	
			startGrafana
			;;	
		-dc1|--dc1|dc1)
			startDC1
			;;
		-dc4|--dc4|dc4)
			startDC4
			;;
		-dc5|--dc5|dc5)
			startDC5
			;;
		*)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi
