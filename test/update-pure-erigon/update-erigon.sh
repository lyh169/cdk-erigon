#!/bin/bash

sudo rm -rf erigon
docker-compose up -d  erigon-update-seq
sleep 100
docker stop erigon-update-seq
docker-compose -f docker-compose.yml rm -f erigon-update-seq

cd ./zkevm-node-fork9
sh stop-node.sh
cd ..

docker-compose up -d cdk-aggregator-db
docker-compose up -d cdk-l1-sync-db
docker-compose up -d erigon-pool-db


cast send --json --rpc-url "http://127.0.0.1:8545" --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0x8dAF17A20c9DBA35f005b6324F493785D239719d" 'setTrustedSequencerURL(string)' 'http://erigon-seq:8123'
sleep 3

docker-compose up -d erigon-prover
docker-compose up -d zkevm-approve
sleep 3

docker-compose up -d erigon-seq

sleep 10
docker-compose up -d cdk-aggregator
docker-compose up -d cdk-sequence-sender
docker-compose up -d erigon-pool-manager
docker-compose up -d erigon-rpc
sleep 3
docker ps -a
