#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Run Hot backup by default"
    echo "Options:"
    echo "  -h, --help, help    Show help"
    echo "  -h, --hot, hot      Run HOT backup"
	echo "  -c, --cold, cold    Run COLD backup (OSS Mediator will reboot)"
	echo
}

version=$(cat VERSION)

stopMediatorContainers(){
	docker stop $(docker ps -a -q)
}

startMediatorContainers(){
	$(pwd)/mediator_docker_startup.sh	
}

backupMediator(){
	tar -czvf ../OSSMediator_"$version"_$(date +%d%m%Y_%H%M).tar.gz ../OSSMediator_$version
}


hotBackup(){

	backupMediator

}

coldBackup(){

	stopMediatorContainers
	backupMediator
	startMediatorContainers
}


# Check if there are any arguments
if [ $# -eq 0 ]; then
    # No arguments, run hot backup
    hotBackup
else
    # Handle arguments
    # Chose Hot or Cold
    case $1 in
        -h|--help|help)
            show_help
            ;;
        -c|--cold|cold)
			coldBackup
            ;;
		-h|--hot|hot)
			hotBackup
			;;
        *)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi
