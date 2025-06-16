#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Deletes OSS Mediator Setup"
    echo "If no option is provied will delete all containers and data"
    echo "Options:"
    echo "  -h, --help, help    Show help"	Show help
    echo "  -os, --opensearch, os"			Delete OpenSearch data   
    echo "  -g, --grafana, grafana"			Delete OpenSearch data   
	echo "  -p, --plugin, plugin"			Delete ElasticSearch plugin 
	echo "  -dc1, --dc1, dc1"					Delete DC1 Collector
	echo "  -dc4, --dc4, dc4"					Delete DC4 Collector					
	echo "  -dc5, --dc5, dc5"					Delete DC5 Collector
	echo
}

# Function to delete all containers
deleteDeployment(){
	echo "Removing OSS Mediator containers..."
	#docker stop $(docker ps -a -q)
	#docker rm $(docker ps -a -q)
	deleteElasticSearchPlugin
	deleteOpenSearch
	deleteGrafana
	deleteDC1
	deleteDC4
	deleteDC5
	rm -rf $(pwd)/collector/ $(pwd)/plugin/ $(pwd)/es_data/ $(pwd)/grafana_storage/ $(pwd)/reports
	echo "OSS Mediator deployment deleted!"
}

# Function to delete dc1 collector
deleteDC1(){
	echo "Removing dc1 collector..."
	docker stop ndac_oss_collector_dc1
	docker rm ndac_oss_collector_dc1
	echo "Removing dc1 collector data..."
	rm -rf $(pwd)/collector/logs/dc1
	rm -rf $(pwd)/collector/checkpoints/dc1
	rm -rf $(pwd)/reports/dc1
}

# Function to delete dc4 collector
deleteDC4(){
	echo "Removing dc4 collector..."
	docker stop ndac_oss_collector_dc4
	docker rm ndac_oss_collector_dc4
	echo "Removing dc4 collector data..."
	rm -rf $(pwd)/collector/logs/dc4
	rm -rf $(pwd)/collector/checkpoints/dc4
	rm -rf $(pwd)/reports/dc4
}

# Function to delete dc5 collector
deleteDC5(){
	echo "Removing dc5 collector..."
	docker stop ndac_oss_collector_dc5
	docker rm ndac_oss_collector_dc5
	echo "Removing dc5 collector data..."
	rm -rf $(pwd)/collector/logs/dc5
	rm -rf $(pwd)/collector/checkpoints/dc5
	rm -rf $(pwd)/reports/dc5
}

# Function to delete dc7 collector
#deleteDC7(){
#	docker stop ndac_oss_collector_dc7
#	docker rm ndac_oss_collector_dc7
#	rm -rf $(pwd)/collector/logs/dc7/
#	rm -rf $(pwd)/collector/checkpoints/dc7/
#	rm -rf $(pwd)/reports/dc7/
#}

# Function to delete ES Plugin 
deleteElasticSearchPlugin(){
	echo "Removing ElasticSearch plugin..."
	docker stop	ndac_oss_elasticsearchplugin
	docker rm ndac_oss_elasticsearchplugin
	echo "Removing ElasticSearch plugin data..."
	rm -rf $(pwd)/plugin/
}


# Function to delete OpenSearch and storage
deleteOpenSearch(){
	echo "Removing OpenSearch plugin..."
	docker stop ndac_oss_opensearch
	docker rm ndac_oss_opensearch
	echo "Removing OpenSearch data..."
	rm -rf $(pwd)/es_data/
}

# Function to delete Grafan and storage
deleteGrafana(){
	echo "Removing Grafana..."
	docker stop ndac_oss_grafana
	docker rm ndac_oss_grafana
	echo "Removing Grafana data..."
	rm -rf $(pwd)/grafana_storage/
}


# Check if there are any containers running
#if [ -z "$(docker ps -qa)" ]; then
#	echo "There are no containers deployed!"
#	show_help
#	exit 1
#
#fi

# Check if there are any arguments
if [ $# -eq 0 ]; then
    # No arguments, wipe everything.
    deleteDeployment
else
	# Handle arguments
    # Wipe specific component
	    case $1 in
        -h|--help|help)
            show_help
            ;;
		-os|--opensearch|os)
			deleteOpenSearch
			;;
		-p|--plugin|plugin)
			deleteElasticSearchPlugin
			;;
		-g|--grafana|grafana)	
			deleteGrafana
			;;	
		-dc1|--dc1|dc1)
			deleteDC1
			;;
		-dc4|--dc4|dc4)
			deleteDC4
			;;
		-dc5|--dc5|dc5)
			deleteDC5
			;;
		*)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi
