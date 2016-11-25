#!/bin/bash

while IFS=: read -r service port; do
    docker rm -f ${service}-active
done < ./services.conf