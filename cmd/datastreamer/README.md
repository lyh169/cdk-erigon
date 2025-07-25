# datastreamer - tool for check and decode datastream data

In the root of `Erigon` project, use this command to build the commands:

```shell
    make datastreamer
```

It can then be run using the following command

```shell
    ./buid/bin/datastreamer sub-command options...
```
+ sub-command contains checker and decoder


## checker - Used for checker the data of datastream correctness

```shell
    datastreamer checker --startBatch=<number_integer>[optional] --endBatch=<number_integer>[optional]  --cfg=<datastreamerConfig.yaml>
```
+ The `startBatch` is the start check batch number of datastream, default is 0 
+ The `endBatch` is the end check batch number of datastream, default is current max batch number
+ The `cfg` is the checker config file, you can see ./config/datastreamerConfig.yaml.example

## operating example:
```shell
  ./datastreamer  checker --startBatch=253450 --endBatch=323450  --cfg=datastreamerConfig.yaml
  ./datastreamer  checker --cfg=datastreamerConfig.yaml
```

## decoder - Used for decoder the data of datastream
```shell
    datastreamer decoder --batchNum=<number_integer>[optional] --blockNum=<number_integer>[optional] --entryNum=<number_integer>[optional]  --cfg=<datastreamerConfig.yaml>
```
+ The `batchNum` is the decode batch number of datastream, default is -1(that is not active for this flag)
+ The `blockNum` is the decode block number of datastream, default is -1(that is not active for this flag)
+ The `entryNum` is the decode entry number of datastream, default is -1(that is not active for this flag)
+ The `cfg` is the checker config file, you can see ./config/datastreamerConfig.yaml.example

## operating example:
```shell
  ./datastreamer decoder --batchNum=100 --cfg=datastreamerConfig.yaml
  ./datastreamer decoder --blockNum=100 --cfg=datastreamerConfig.yaml
  ./datastreamer decoder --entryNum=100 --cfg=datastreamerConfig.yaml
```


