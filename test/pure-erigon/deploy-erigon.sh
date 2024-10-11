#!/bin/bash

#make stop
docker-compose -f docker-compose.yml down --remove-orphans
rm -rf datastream.bin && sudo rm -rf datastream.db
#docker ps
#sleep 3

docker-compose up -d cdk-aggregator-db
docker-compose up -d cdk-l1-sync-db
docker-compose up -d erigon-pool-db
docker-compose up -d cdk-data-availability-db

docker-compose up -d erigon-mock-l1-network
sleep 10
cast send --json --rpc-url "http://127.0.0.1:8545" --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" 'setupCommittee(uint256,string[],bytes)' 1 '["http://cdk-data-availability:8444"]' 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
cast send --json --rpc-url "http://127.0.0.1:8545" --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0x8dAF17A20c9DBA35f005b6324F493785D239719d" 'setTrustedSequencerURL(string)' 'http://erigon-seq:8123'
sleep 3

docker-compose up -d cdk-data-availability
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
