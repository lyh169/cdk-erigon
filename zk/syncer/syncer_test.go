package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/ethclient"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/stretchr/testify/require"
)

func TestL1(t *testing.T) {
	//client, err := ethclient.Dial("http://43.134.127.80:9545")
	//client, err := ethclient.Dial("http://127.0.0.1:8545")
	client, err := ethclient.Dial("http://43.134.127.80:8123")
	require.NoError(t, err)

	blk, err := client.BlockNumber(context.Background())
	require.NoError(t, err)
	rpc.SafeBlockNumber.Int64()
	latestBlock, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blk)))
	require.NoError(t, err)
	fmt.Println("latestBlock", latestBlock, blk)
}

func TestDiffentBlockHash(t *testing.T) {
	srcClient, err := ethclient.Dial("https://rpc.merlinchain.io")
	require.NoError(t, err)
	dstClient, err := ethclient.Dial("http://35.247.156.252:8123")
	require.NoError(t, err)

	from := int64(11000600) // mainnet
	to := int64(12222828)

	var m ethconfig.Merlin
	var sBlk *types.Header
	var dBlk *types.Header
	wg := sync.WaitGroup{}
	for i := from; i < to; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			sBlk = getBlockFromClient(srcClient, big.NewInt(i))
		}()
		go func() {
			defer wg.Done()
			dBlk = getBlockFromClient(dstClient, big.NewInt(i))
		}()
		wg.Wait()
		if sBlk.ReceiptHash != dBlk.ReceiptHash {
			m.Headers = append(m.Headers, sBlk)
		}
		if i%100 == 0 {
			fmt.Println("current query blocknum is", i, "headercount", len(m.Headers))
			if len(m.Headers) > 0 {
				err = WriteFile("dynamic-mynetwork-block.json", &m)
				require.NoError(t, err)
			}
		}
	}

	if len(m.Headers) > 0 {
		err = WriteFile("dynamic-mynetwork-block.json", &m)
		require.NoError(t, err)
	}
}

func getBlockFromClient(client *ethclient.Client, num *big.Int) *types.Header {
	sBlk, err := client.HeaderByNumber(context.Background(), num)
	if err != nil {
		time.Sleep(time.Second)
		sBlk, err = client.HeaderByNumber(context.Background(), num)
		if err != nil {
			panic(fmt.Sprintf("BlockByNumber %d, error %s", num.Uint64(), err.Error()))
		}
		return sBlk
	}
	return sBlk
}

func TestBatchDiffentBlockHash(t *testing.T) {
	//srcClient, err := ethclient.Dial("https://rpc.merlinchain.io")
	//require.NoError(t, err)
	//dstClient, err := ethclient.Dial("http://35.247.156.252:8123")
	//require.NoError(t, err)

	srcClient, err := rpc.Dial("https://rpc.merlinchain.io")
	require.NoError(t, err)
	dstClient, err := rpc.Dial("http://35.247.156.252:8123")
	require.NoError(t, err)

	from := int64(11296900) // mainnet
	to := int64(12222828)

	var m ethconfig.Merlin
	var sBlk []*types.Header
	var dBlk []*types.Header
	wg := sync.WaitGroup{}
	const batchNum = 100
	for i := from; i < to; i = i + batchNum {
		end := i + batchNum
		if end >= to {
			end = to - 1
		}
		wg.Add(2)
		go func() {
			defer wg.Done()
			sBlk, err = batchGetHeaderFromClient(srcClient, i, end)
			if err != nil {
				sBlk, err = batchGetHeaderFromClient(srcClient, i, end)
				if err != nil {
					panic(err)
				}
			}
		}()
		go func() {
			defer wg.Done()
			dBlk, err = batchGetHeaderFromClient(dstClient, i, end)
			if err != nil {
				dBlk, err = batchGetHeaderFromClient(dstClient, i, end)
				if err != nil {
					panic(err)
				}
			}
		}()
		wg.Wait()

		require.Equal(t, len(sBlk), len(dBlk))

		for j := 0; j < len(sBlk); j++ {
			if sBlk[j].ReceiptHash != dBlk[j].ReceiptHash {
				m.Headers = append(m.Headers, sBlk[j])
			}
		}
		if end%100 == 0 {
			fmt.Println("current query blocknum is", end, "headercount", len(m.Headers))
			if len(sBlk) > 0 {
				fmt.Println("current number is", sBlk[0].Number.Uint64(), "block hash is ", sBlk[0].Hash())
			}
			fmt.Println()
			if len(m.Headers) > 0 {
				err = WriteFile("dynamic-mynetwork-block.json", &m)
				require.NoError(t, err)
			}
		}
	}

	if len(m.Headers) > 0 {
		err = WriteFile("dynamic-mynetwork-block.json", &m)
		require.NoError(t, err)
	}
}
func batchGetHeaderFromClient(client *rpc.Client, from, to int64) ([]*types.Header, error) {
	num := to - from
	headers := make([]*types.Header, num)
	reqs := make([]rpc.BatchElem, num)
	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(i) + from)), false},
			Result: &headers[i],
		}
	}
	if err := client.BatchCallContext(context.Background(), reqs); err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, reqs[i].Error
		}
		if headers[i] == nil {
			return nil, fmt.Errorf("got null header for  %d of block", int64(i)+from)
		}
	}
	return headers, nil
}

func WriteFile(outFile string, m *ethconfig.Merlin) error {
	data, err := json.MarshalIndent(*m, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(outFile, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
