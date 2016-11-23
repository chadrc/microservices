#!/bin/bash

# Create array of services based on command line params
# If no params were given use default service list
PARAM_COUNT=$#
if [ ${PARAM_COUNT} -ge 1 ] ; then
    COUNT=0
    SERVICES=()
    for SERVICE in $@; do
        SERVICES[${COUNT}]=${SERVICE}
        COUNT+=1
    done
else
    SERVICES=(
        loginservice
        loggerservice
        cardgameservice
    )
fi

# Loop through services and do docker build
for SERVICE in ${SERVICES[@]}; do
    if [ -d "./"${SERVICE} ]; then
        echo Building ${SERVICE}
        # Copy base Dockerfile to service root folder
        cp ./Dockerfile ./${SERVICE}/Dockerfile
        sudo docker build -t=${SERVICE} ./${SERVICE}
        rm -f ./${SERVICE}/Dockerfile
    else
        echo ${SERVICE} "does not exist."
    fi
done