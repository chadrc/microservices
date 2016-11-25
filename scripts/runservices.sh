#!/bin/bash

mkdir -p temp

while IFS=: read -r service port; do
    docker run --name=${service}-active -d -p ${port}:8080 ${service} > ./temp/${service}.containerid
done < ./services.conf
