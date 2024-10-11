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
2. Run the deploy cdk-erigon components
```shell
cd cdk-erigon/test/pure-erigon
sh delpoy-erigon.sh
```
3. Send tx
```shell
cast send --json --rpc-url "http://127.0.0.1:8124"  --legacy --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0xB9Ce0C6f1Dfd196a3890F0dB994f9Cf2C4606042" --value 10ether
```
4. Stop the componments
```shell
make stop
```
