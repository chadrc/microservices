#!/bin/bash

cat ./services.conf | while IFS=: read -r service port; do
    sudo docker rm -f `cat ./temp/${service}.containerid`
    rm -f ./temp/${service}.containerid
done