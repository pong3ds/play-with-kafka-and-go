#!/bin/bash

export HOST_IP_ADDRESS=$(ipconfig getifaddr $(route -n get default|awk '/interface/ { print $2 }'))
docker-compose up -d
