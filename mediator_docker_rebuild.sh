#!/bin/bash

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Rebuilds whole setup by default"
    echo "Options:"
    echo "  -h, --help, help                           help Show this help message"
    echo "  -c, --container, container <container_id>  stop and remove the specified Docker container, then run startup"
    echo
    echo "If no arguments are provided, the script will run rebuild whole deployment sequence."
}

# Function to run the startup script
run_startup() {
    $(pwd)/mediator_docker_startup.sh
}

# Function to run the Wipe script
run_wipe() {
    $(pwd)/mediator_docker_wipe.sh
}

# Function to run the wipe and startup scripts
rebuild_deployment() {
    run_wipe
    run_startup
}


# Check if there are any arguments
if [ $# -eq 0 ]; then
    # No arguments, run wipe and startup
    rebuild_deployment
else
    # Handle arguments
    case $1 in
        -h|--help|help)
            show_help
            ;;
        -c|--container|container)
            # Check if a container ID or name is provided
            if [ -n "$2" ]; then
                # Check if the container is running
                if docker start $2 ; then
                    docker stop "$2"
                    docker rm "$2"
                    run_startup
                else
                    echo "Error: Container '$2' is not running."
                    exit 1
                fi
            else
                echo "Error: No container ID or name provided."
                exit 1
            fi
            ;;
        *)
            echo "Invalid command. Use -h or --help for usage information."
            exit 1
            ;;
    esac
fi

