#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "stop OSS Mediator Container"
	echo "Options:"
	echo "  -h, --help, help               Show help"
	echo "  -a, --all, all                 Stop ALL containers"
	echo "  -os, --opensearch, opensearch  Stop OpenSearch data"
	echo "  -g, --grafana, grafana         Stop Grafana"
	echo "  -p, --plugin, plugin           Stop ElasticSearch plugin"
	echo "  -dc1, --dc1, dc1               Stop DC1 Collector"
	echo "  -dc4, --dc4, dc4               Stop DC4 Collector"
	echo "  -dc5, --dc5, dc5               Stop DC5 Collector"
	echo
}

# Function to stop all containers
stopContainers(){
	#docker stop $(docker ps -a -q)
	stopGrafana
	stopElasticSearchPlugin
	stopDC1
	stopDC4
	stopDC5
	stopOpenSearch
}

# Function to stop dc1 collector
stopDC1(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc1"; then
		echo "Stopping DC1 Collector..."
		docker stop ndac_oss_collector_dc1
	else
		echo "DC1 Collector is stopped."
	fi
}

# Function to stop dc4 collector
stopDC4(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc4"; then
		echo "Stopping DC4 Collector..."
		docker stop ndac_oss_collector_dc4		
	else
		echo "DC4 Collector is stopped."
	fi

}

# Function to stop dc5 collector
stopDC5(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_collector_dc5"; then
		echo "Stopping DC5 Collector..."
		docker stop ndac_oss_collector_dc5		
	else
		echo "DC5 Collector is stopped."
	fi
}

# Function to stop ES Plugin 
stopElasticSearchPlugin(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_elasticsearchplugin"; then
		echo "Stopping ElasticSearch Plugin..."
		docker stop ndac_oss_elasticsearchplugin		
	else
		echo "ElasticSearch Plugin is stopped."
	fi
}


# Function to stop OpenSearch and storage
stopOpenSearch(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_opensearch"; then
		echo "Stopping OpenSearch..."
		docker stop ndac_oss_opensearch		
	else
		echo "OpenSearch is stopped."
	fi
}

# Function to stop Grafana and storage
stopGrafana(){
	if docker ps --format '{{.Names}}' | grep -q "ndac_oss_grafana"; then
		echo "Stopping Grafana..."
		docker stop ndac_oss_grafana		
	else
		echo "Grafana is stopped."
	fi
}


### MAIN stopS HERE ######

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
            stopContainers
            ;;
		-os|--opensearch|os)
			stopOpenSearch
			;;
		-p|--plugin|plugin)
			stopElasticSearchPlugin
			;;
		-g|--grafana|grafana)	
			stopGrafana
			;;	
		-dc1|--dc1|dc1)
			stopDC1
			;;
		-dc4|--dc4|dc4)
			stopDC4
			;;
		-dc5|--dc5|dc5)
			stopDC5
			;;
		*)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi
