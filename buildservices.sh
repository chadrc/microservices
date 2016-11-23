#!/bin/bash

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

SERVICE_ROOT=github.com/chadrc/microservices

for SERVICE in ${SERVICES[@]}; do
    if [ -d "./"${SERVICE} ]; then
        echo Building ${SERVICE}
        cp ./Dockerfile ./${SERVICE}/Dockerfile
        echo | sudo docker build -t=${SERVICE} ./${SERVICE}
        rm -f ./${SERVICE}/Dockerfile
    else
        echo ${SERVICE} "does not exist."
    fi
done