package commands

import (
	"fmt"
	"testing"

	"github.com/gateway-fm/zkevm-data-streamer/datastreamer"
	"github.com/ledgerwatch/erigon/zk/datastream/server"
)

func TestUnwindDatastream(t *testing.T) {
	fileName := "/root/go/src/cdk-erigon/workdir/data-stream.bin"
	var dataStreamServerFactory = server.NewZkEVMDataStreamServerFactory()
	ds, err := dataStreamServerFactory.CreateStreamServer(
		0,
		0,
		1,
		datastreamer.StreamType(1),
		fileName,
		1,
		2,
		3,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	dataStreamServer := dataStreamServerFactory.CreateDataStreamServer(ds, 1001)
	dataStreamServer.UnwindToBlock(2)
}
