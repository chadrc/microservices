#!/bin/bash

SERVICE_ROOT=github.com/chadrc/microservices

while IFS=: read -r service port; do
    echo "Installing ${service}..."
    go install ${SERVICE_ROOT}/${service}
    echo "Installed"
done < ./services.conf