#!/bin/bash

mkdir -p temp

while IFS=: read -r service port; do
    docker run -d -p 8080:${port} ${service} > ./temp/${service}.containerid
done < ./services.conf
