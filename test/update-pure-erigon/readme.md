# This test is update from the zkevm-node to cdk-erigon (it can replace all the zkevm-node components to cdk-erigon)
# How to run
1. DownLoad the code and make docker images
```shell
git clone https://github.com/0xPolygon/cdk-data-availability.git -b v0.0.7
cd cdk-data-availability && make build-docker
git clone https://github.com/0xPolygon/cdk-validium-node.git -b v0.7.0+cdk
cd cdk-validium-node && make build-docker
git clone https://github.com/0xPolygonHermez/cdk-erigon.git -b local-run
cd cdk-erigon make build-docker
```
2. Run the zkevm-node components: 
```shell
cd cdk-erigon/test/update-pure-erigon
sh deploy-zkevm-node.sh
```
4. Wait the zkevm-node up to the HaltOnBatchNumber(that is 5), you can query the verified batch:
```shell
cast --json --rpc-url "http://127.0.0.1:8124" rpc zkevm_verifiedBatchNumber
```
5. Run the update erigon shell script
```shell
sh update-erigon.sh
```
6. Stop the all componments
```shell
make stop
cd zkevm-node-fork9
make stop
cd ..
```
