#!/usr/bin/env bash

SERVICE_ROOT=github.com/chadrc/microservices

go install ${SERVICE_ROOT}/loginservice
go install ${SERVICE_ROOT}/loggerservice
go install ${SERVICE_ROOT}/cardgameservice