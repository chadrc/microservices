#!/bin/bash

SERVICES=(
loginservice
loggerservice
cardgameservice
)

SERVICE_ROOT=github.com/chadrc/microservices

for SERVICE in ${SERVICES[@]}; do
    echo Installing ${SERVICE}
    go install ${SERVICE_ROOT}/${SERVICE}
done