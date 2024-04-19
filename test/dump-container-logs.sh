#!/bin/bash

# Loop through all running Docker containers
docker ps --format "{{.ID}} {{.Names}}" | while read -r container_id container_name; do
    # Define the log file name
    log_file="$container_name.log"
    # Define the temporary log file name
    tmp_log_file="temp_$container_name.log"

    # Dump logs of the container to a temporary file
    docker logs "$container_id" &> "temp_$container_name.log"

    # Clean up the logs (remove ANSI escape codes) and save to the final log file
    sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g" "temp_$container_name.log" > "$log_file"

    # Remove the temporary log file
    rm "temp_$container_name.log"

    # Display a message indicating completion for each container
    echo "Logs saved to $log_file for container $container_name"
done