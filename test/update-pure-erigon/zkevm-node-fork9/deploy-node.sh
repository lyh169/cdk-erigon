#!/bin/bash
make stop
sleep 1
docker ps
sleep 3

docker-compose up -d zkevm-state-db
docker-compose up -d zkevm-pool-db
docker-compose up -d zkevm-event-db
docker-compose up -d cdk-data-availability-db
sleep 1

docker-compose up -d zkevm-mock-l1-network
sleep 3
cast send --json --rpc-url "http://127.0.0.1:8545" --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" 'setupCommittee(uint256,string[],bytes)' 1 '["http://cdk-data-availability:8444"]' 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
#cast send --json --rpc-url "http://127.0.0.1:8545" --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" 'setupCommittee(uint256,string[],bytes)' 1 '["http://127.0.0.1:8444"]' 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
sleep 3

docker-compose up -d cdk-data-availability
docker-compose up -d zkevm-sync
sleep 3
docker-compose up -d zkevm-prover
sleep 3

docker-compose up -d zkevm-eth-tx-manager
docker-compose up -d zkevm-approve
docker-compose up -d zkevm-sequencer
docker-compose up -d zkevm-sequence-sender
docker-compose up -d zkevm-l2gaspricer
docker-compose up -d zkevm-aggregator
docker-compose up -d zkevm-json-rpc
sleep 3
docker ps -a
