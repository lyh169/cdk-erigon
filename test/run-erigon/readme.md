# How to run
1. DownLoad the code and make docker images
```shell
git clone https://github.com/MerlinLayer2/cdk-erigon.git
cd cdk-erigon 
docker build -t cdk-erigon -f ./Dockerfile .
```
2. Run the deploy cdk-erigon components
```shell
cd test/run-erigon
make run
```
3. Send tx
```shell
cast send --json --rpc-url "http://127.0.0.1:8124"  --legacy --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"  "0xB9Ce0C6f1Dfd196a3890F0dB994f9Cf2C4606042" --value 10ether
```
4. Stop the componments
```shell
make stop
```
