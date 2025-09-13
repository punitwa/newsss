#!/bin/bash

# Docker Log Cleanup Script
# Clears logs for all containers every 20 minutes

echo "$(date): Starting Docker log cleanup"

# Clean logs for all containers (running and stopped) over 10MB
find /var/lib/docker/containers -name "*-json.log" -size +10M -exec ls -lh {} \; | while read -r line; do
    logfile=$(echo "$line" | awk '{print $NF}')
    size=$(echo "$line" | awk '{print $5}')
    echo "$(date): Found large log file: $logfile ($size)"
    
    if [ -f "$logfile" ]; then
        echo "$(date): Truncating $logfile"
        truncate -s 0 "$logfile"
        echo "$(date): Truncated $logfile"
    fi
done

# Also clean logs for running containers specifically
CONTAINERS=$(docker ps -q)
if [ -n "$CONTAINERS" ]; then
    echo "$(date): Cleaning logs for running containers: $CONTAINERS"
    
    for container in $CONTAINERS; do
        CONTAINER_NAME=$(docker inspect --format='{{.Name}}' $container 2>/dev/null | sed 's/\///')
        LOG_FILE=$(docker inspect --format='{{.LogPath}}' $container 2>/dev/null)
        
        if [ -f "$LOG_FILE" ]; then
            FILE_SIZE=$(stat -c%s "$LOG_FILE" 2>/dev/null || echo "0")
            if [ "$FILE_SIZE" -gt 10485760 ]; then  # 10MB
                echo "$(date): Truncating logs for $CONTAINER_NAME ($FILE_SIZE bytes)"
                truncate -s 0 "$LOG_FILE"
            fi
        fi
    done
fi

# Optional: Clean up old log files and unused Docker resources
echo "$(date): Cleaning up Docker system"
docker system prune -f --filter "until=24h" > /dev/null 2>&1

echo "$(date): Cleanup script finished"

