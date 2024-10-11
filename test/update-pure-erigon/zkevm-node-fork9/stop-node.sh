#!/bin/bash
#stop
docker stop zkevm-eth-tx-manager
docker stop zkevm-sequencer
docker stop zkevm-sequence-sender
docker stop zkevm-l2gaspricer
docker stop zkevm-aggregator
docker stop zkevm-json-rpc

docker stop zkevm-prover
docker stop zkevm-sync

docker-compose -f docker-compose.yml rm -f zkevm-eth-tx-manager
docker-compose -f docker-compose.yml rm -f zkevm-sequencer
docker-compose -f docker-compose.yml rm -f zkevm-sequence-sender
docker-compose -f docker-compose.yml rm -f zkevm-l2gaspricer
docker-compose -f docker-compose.yml rm -f zkevm-aggregator
docker-compose -f docker-compose.yml rm -f zkevm-json-rpc
docker-compose -f docker-compose.yml rm -f zkevm-prover
docker-compose -f docker-compose.yml rm -f zkevm-sync
docker-compose -f docker-compose.yml rm -f zkevm-approve
rm -rf datastream.bin && sudo rm -rf datastream.db
