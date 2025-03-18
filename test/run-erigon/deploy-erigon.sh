#!/bin/bash

make stop
docker-compose -f docker-compose.yml down --remove-orphans
mkdir erigon && mkdir erigon/data


docker-compose up -d cdk-aggregator-db
docker-compose up -d erigon-pool-db
docker-compose up -d cdk-data-availability-db

docker-compose up -d erigon-mock-l1-network
sleep 3

docker-compose up -d erigon-prover
docker-compose up -d erigon-stateless-executor
sleep 3
docker-compose up -d cdk-data-availability
docker-compose up -d erigon-seq
docker-compose up -d erigon-rpc

sleep 3
docker-compose up -d cdk-aggregator
docker-compose up -d cdk-sequence-sender
docker-compose up -d erigon-pool-manager
sleep 3
docker ps -a
