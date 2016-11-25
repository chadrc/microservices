#!/bin/bash

mkdir -p temp
DOCKER_BRIDGE_IP=`ifconfig docker0 | grep "inet addr:" | cut -d: -f2 | awk '{print $1}'`
while IFS=: read -r service port; do
    docker run --name=${service}-active -d -e DOCKER_BRIDGE_IP=${DOCKER_BRIDGE_IP} -p ${port}:8080 ${service} > ./temp/${service}.containerid
done < ./services.conf
