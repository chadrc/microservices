#!/bin/bash

sudo docker rm -f `cat ./temp/loginservice.containerid`
rm -f ./temp/loginservice.containerid