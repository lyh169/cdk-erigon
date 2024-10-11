#!/bin/bash

# need halt the zkevm-node batch
cd ./zkevm-node-fork9
sh deploy-node.sh
cast send --json --rpc-url "http://127.0.0.1:8124"  --legacy --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0xB9Ce0C6f1Dfd196a3890F0dB994f9Cf2C4606042" --value 10ether
cd ..
