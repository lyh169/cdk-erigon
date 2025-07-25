package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ledgerwatch/erigon/cmd/datastreamer/config"
	"github.com/ledgerwatch/erigon/cmd/datastreamer/util"
	"github.com/ledgerwatch/erigon/zk/datastream/types"
	jclient "github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/urfave/cli/v2"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon/ethclient"
	"github.com/ledgerwatch/erigon/zk/datastream/client"
	"github.com/ledgerwatch/erigon/zk/datastream/proto/github.com/0xPolygonHermez/zkevm-node/state/datastream"
)

var (
	startBatchNum uint64
	endBatchNum   uint64
)

var Command = cli.Command{
	Action: run,
	Name:   "checker",
	Usage:  "check the data of datastream",
	Flags: []cli.Flag{
		&util.CfgFileFlag,
		&cli.Uint64Flag{
			Name:        "startBatch",
			Usage:       "check the datastream of the start batch bumber",
			Destination: &startBatchNum,
			Value:       0,
		},
		&cli.Uint64Flag{
			Name:        "endBatch",
			Usage:       "check the datastream of the end batch bumber",
			Destination: &endBatchNum,
			Value:       0,
		},
	},
}

func run(cliCtx *cli.Context) error {
	configFile := cliCtx.String(util.CfgFileFlag.Name)
	if configFile == "" {
		return errors.New("The flag cfg is not set")
	}

	cfg, err := config.GetConf(configFile)
	if err != nil {
		return err
	}

	log.Init(log.Config{
		Environment: "development",
		Level:       cfg.LogLevel,
		Outputs:     []string{"stderr"},
	})

	if cfg.Version == 0 {
		cfg.Version = 2
	}
	ctx := cliCtx.Context
	client := client.NewClient(ctx, cfg.DatastreamUrl, false, cfg.Version, 500, 0)
	defer client.Stop()
	if err := client.Start(); err != nil {
		panic(err)
	}

	// create bookmark
	bookmark := types.NewBookmarkProto(0, datastream.BookmarkType_BOOKMARK_TYPE_L2_BLOCK)

	var previousFile *types.FileEntry
	var lastBlockRoot common.Hash
	progressBatch := uint64(0)
	progressBlock := uint64(0)
	lastSeenBatch := uint64(0)
	lastSeenBlock := uint64(0)

	endBlockNum := uint64(0)
	if endBatchNum == 0 {
		ret, err := BatchNumber(cfg.L2Url)
		if err == nil && ret != 0 {
			endBatchNum = ret
		}
		log.Info("current ", " endBatchNum ", endBatchNum)
	}

	if startBatchNum > endBatchNum {
		return errors.New("The flag startBatch is bigger than endBatch")
	}

	if startBatchNum > 1 {
		LastBatch, err := BatchByNumber(cfg.L2Url, startBatchNum-1)
		if err != nil {
			log.Error("BatchByNumber fail ", " error ", err)
			return err
		}
		if len(LastBatch.Blocks) == 0 {
			log.Error("This batch is empty ", " batch number ", startBatchNum-1)
			return err
		}

		eclient, err := ethclient.Dial(cfg.L2Url)
		if err != nil {
			log.Error("ethclient.Dial fail ", " error ", err)
			return err
		}

		block, err := eclient.BlockByHash(ctx, LastBatch.Blocks[len(LastBatch.Blocks)-1])
		if err != nil {
			log.Error("eclient.BlockByHash fail ", " error ", err)
			return err
		}

		bookmark = types.NewBookmarkProto(startBatchNum, datastream.BookmarkType_BOOKMARK_TYPE_BATCH)
		progressBatch = startBatchNum
		progressBlock = block.NumberU64()
		lastSeenBatch = startBatchNum - 1
		lastSeenBlock = block.NumberU64()
		lastBlockRoot = block.Root()

		log.Info("param is ", " progressBatch ", progressBatch, " progressBlock ", progressBlock, " lastSeenBatch ", lastSeenBatch, " lastBlockRoot ", lastBlockRoot)
	}

	function := func(file *types.FileEntry) error {
		switch file.EntryType {
		case types.EntryTypeL2BlockEnd:
			log.Debug("EntryTypeL2BlockEnd")
			if previousFile != nil && previousFile.EntryType != types.EntryTypeL2Block && previousFile.EntryType != types.EntryTypeL2Tx {
				return fmt.Errorf("unexpected entry type before l2 block end: %v", previousFile.EntryType)
			}
			if endBlockNum > 0 && progressBlock > endBlockNum {
				log.Info("success check end ", " endblockNum is ", endBlockNum)
				return fmt.Errorf("success check end endblockNum is %d", endBlockNum)
			}
		case types.BookmarkEntryType:
			bookmark, err := types.UnmarshalBookmark(file.Data)
			if err != nil {
				return err
			}
			if bookmark.BookmarkType() == datastream.BookmarkType_BOOKMARK_TYPE_BATCH {
				log.Debug("***************************** BookmarkEntryType, progressBatch ***************************", " number ", bookmark.Value)
				progressBatch = bookmark.Value
				if previousFile != nil && previousFile.EntryType != types.EntryTypeBatchEnd {
					return fmt.Errorf("unexpected entry type before batch bookmark type: %v, bookmark batch number: %d", previousFile.EntryType, bookmark.Value)
				}
			}
			if bookmark.BookmarkType() == datastream.BookmarkType_BOOKMARK_TYPE_L2_BLOCK {
				log.Debug("BookmarkEntryType ", " progressBlock ", bookmark.Value)
				progressBlock = bookmark.Value
				if previousFile != nil &&
					previousFile.EntryType != types.EntryTypeBatchStart &&
					previousFile.EntryType != types.EntryTypeL2BlockEnd {
					return fmt.Errorf("unexpected entry type before block bookmark type: %v, bookmark block number: %d", previousFile.EntryType, bookmark.Value)
				}
			}
		case types.EntryTypeBatchStart:
			batchStart, err := types.UnmarshalBatchStart(file.Data)
			if err != nil {
				return err
			}
			log.Debug("EntryTypeBatchStart ", " progressBatch ", batchStart.Number)

			if lastSeenBatch+1 != batchStart.Number {
				return fmt.Errorf("unexpected batch %d, expected %d", batchStart.Number, lastSeenBatch+1)
			}

			lastSeenBatch = batchStart.Number
			progressBatch = batchStart.Number

			if previousFile != nil {
				if previousFile.EntryType != types.BookmarkEntryType {
					return fmt.Errorf("unexpected entry type before batch start: %v, batchStart Batch number: %d", previousFile.EntryType, batchStart.Number)
				} else {
					bookmark, err := types.UnmarshalBookmark(previousFile.Data)
					if err != nil {
						return err
					}
					if bookmark.BookmarkType() != datastream.BookmarkType_BOOKMARK_TYPE_BATCH {
						return fmt.Errorf("unexpected bookmark type before batch start: %v, batchStart Batch number: %d", bookmark.BookmarkType(), batchStart.Number)
					}
				}
			}
		case types.EntryTypeBatchEnd:
			if previousFile != nil &&
				previousFile.EntryType != types.EntryTypeL2BlockEnd &&
				previousFile.EntryType != types.EntryTypeL2Tx &&
				previousFile.EntryType != types.EntryTypeL2Block &&
				previousFile.EntryType != types.EntryTypeBatchStart {
				return fmt.Errorf("unexpected entry type before batch end: %v", previousFile.EntryType)
			}
			batchEnd, err := types.UnmarshalBatchEnd(file.Data)
			if err != nil {
				return err
			}
			log.Debug("EntryTypeBatchEnd ", " batch number ", batchEnd.Number)

			if batchEnd.Number != progressBatch {
				return fmt.Errorf("batch end number mismatch: %d, expected: %d", batchEnd.Number, progressBatch)
			}
			if batchEnd.StateRoot != lastBlockRoot && !checkIsEmptyBatch(cfg.L2Url, batchEnd.Number) {
				return fmt.Errorf("batch end state root mismatch: %x, expected: %x", batchEnd.StateRoot, lastBlockRoot)
			}
			if endBatchNum > 0 && batchEnd.Number >= endBatchNum {
				log.Info("success check end ", " endBatchNum is ", endBatchNum)
				return fmt.Errorf("success check end endBatchNum is %d", endBatchNum)
			}
		case types.EntryTypeL2Tx:
			if previousFile != nil && previousFile.EntryType != types.EntryTypeL2Tx && previousFile.EntryType != types.EntryTypeL2Block {
				return fmt.Errorf("unexpected entry type before l2 tx: %v", previousFile.EntryType)
			}
		case types.EntryTypeL2Block:
			l2Block, err := types.UnmarshalL2Block(file.Data)
			if err != nil {
				return err
			}
			log.Debug("EntryTypeL2Block ", " l2 block number ", l2Block.L2BlockNumber)
			if l2Block.L2BlockNumber > 0 && lastSeenBlock+1 != l2Block.L2BlockNumber {
				return fmt.Errorf("unexpected block %d, expected %d", l2Block.L2BlockNumber, lastSeenBlock+1)
			}

			lastSeenBlock = l2Block.L2BlockNumber
			progressBlock = l2Block.L2BlockNumber
			if previousFile != nil {
				if previousFile.EntryType != types.BookmarkEntryType && !previousFile.IsL2BlockEnd() {
					return fmt.Errorf("unexpected entry type before l2 block: %v, block number: %d", previousFile.EntryType, l2Block.L2BlockNumber)
				} else {
					bookmark, err := types.UnmarshalBookmark(previousFile.Data)
					if err != nil {
						return err
					}
					if bookmark.BookmarkType() != datastream.BookmarkType_BOOKMARK_TYPE_L2_BLOCK {
						return fmt.Errorf("unexpected bookmark type before l2 block: %v, block number: %d", bookmark.BookmarkType(), l2Block.L2BlockNumber)
					}

				}
			}
			lastBlockRoot = l2Block.StateRoot
		case types.EntryTypeGerUpdate:
			return nil
		default:
			return fmt.Errorf("unexpected entry type: %v", file.EntryType)
		}

		previousFile = file
		return nil
	}

	// send start command
	err = client.ExecutePerFile(bookmark, function)
	log.Info("progress ", " block ", progressBlock)
	log.Info("progress ", " batch ", progressBatch)
	if err != nil && !strings.Contains(err.Error(), "success") {
		panic(err)
	}
	log.Info("success")
	return nil
}

type ZkEVMBatch struct {
	AccInputHash   string        `json:"accInputHash"`
	Blocks         []common.Hash `json:"blocks"`
	BatchL2Data    string        `json:"batchL2Data"`
	Coinbase       string        `json:"coinbase"`
	GlobalExitRoot string        `json:"globalExitRoot"`
	LocalExitRoot  string        `json:"localExitRoot"`
	StateRoot      string        `json:"stateRoot"`
	Closed         bool          `json:"closed"`
	Timestamp      string        `json:"timestamp"`
}

func checkIsEmptyBatch(url string, number uint64) bool {
	batch, err := BatchByNumber(url, number)
	if err != nil {
		log.Warn("checkIsEmptyBatch ", " batch number ", number, " error ", err)
		return false
	}
	return len(batch.Blocks) == 0
}

func BatchByNumber(url string, number uint64) (*ZkEVMBatch, error) {
	response, err := jclient.JSONRPCCall(url, "zkevm_getBatchByNumber", number)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf(response.Error.Message)
	}

	var result ZkEVMBatch
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func BatchNumber(url string) (uint64, error) {
	response, err := jclient.JSONRPCCall(url, "zkevm_batchNumber")
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf(response.Error.Message)
	}

	var result hexutil.Uint64
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return 0, err
	}

	return uint64(result), nil
}
