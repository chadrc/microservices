#!/bin/bash

SERVICES=(
loginservice
loggerservice
cardgameservice
)

SERVICE_ROOT=github.com/chadrc/microservices

for SERVICE in ${SERVICES[@]}; do
    echo Building ${SERVICE}
    SERVICE_PATH=${SERVICE_ROOT}/${SERVICE}
    echo "      "Path: ${SERVICE_PATH}
    EXPORT_PATH=${GOPATH}/src/${SERVICE_ROOT}/${SERVICE}/build/${SERVICE}.service
    go build -o ${EXPORT_PATH} ${SERVICE_PATH}
    echo "      "Build: ${EXPORT_PATH}
done