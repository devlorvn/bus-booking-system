#!/bin/sh
set -e

# Delete old service if exists
docker service rm cluster-prune || true

# Create the global service
docker service create \
  --name cluster-prune \
  --mode global \
  --restart-condition none \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  192.168.220.128:5000/alpine:3.19 \
  sh -c 'echo -e "POST /v1.41/containers/prune?force=true HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n" | nc -U /var/run/docker.sock && echo -e "POST /v1.41/images/prune?all=true HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n" | nc -U /var/run/docker.sock && echo -e "POST /v1.41/volumes/prune HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n" | nc -U /var/run/docker.sock'

echo "cluster-prune service created successfully!"
