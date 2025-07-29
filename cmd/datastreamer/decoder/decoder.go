package decoder

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	"github.com/fatih/color"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/erigon/cmd/datastreamer/config"
	"github.com/ledgerwatch/erigon/cmd/datastreamer/util"
	ctypes "github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/zk/datastream/types"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/urfave/cli/v2"

	"github.com/ledgerwatch/erigon/zk/datastream/proto/github.com/0xPolygonHermez/zkevm-node/state/datastream"
)

var (
	batchNum int64
	blockNum int64
	entryNum int64
)

var Command = cli.Command{
	Action: run,
	Name:   "decoder",
	Usage:  "decode the give batch or block of datastream",
	Flags: []cli.Flag{
		&util.CfgFileFlag,
		&cli.Int64Flag{
			Name:        "batchNum",
			Usage:       "decode the batch of datastream",
			Destination: &batchNum,
			Value:       -1,
		},
		&cli.Int64Flag{
			Name:        "blockNum",
			Usage:       "decode the block of datastream",
			Destination: &blockNum,
			Value:       -1,
		},
		&cli.Int64Flag{
			Name:        "entryNum",
			Usage:       "decode the entry of datastream",
			Destination: &entryNum,
			Value:       -1,
		},
	},
}

func run(cliCtx *cli.Context) error {
	configFile := cliCtx.String(util.CfgFileFlag.Name)
	if configFile == "" {
		return errors.New("The flag cfg is not set")
	}

	if batchNum < 0 && blockNum < 0 && entryNum < 0 {
		return errors.New("Not set the flag of batchNum | blockNum | entryNum")
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

	client, err := datastreamer.NewClient(cfg.DatastreamUrl, 1)
	if err != nil {
		log.Error("datastreamer newclient ", " err ", err)
		return err
	}

	if batchNum >= 0 {
		return decodeBatch(client)
	} else if blockNum >= 0 {
		return decodeL2Block(client)
	} else if entryNum >= 0 {
		return decodeEntry(client)
	}
	return nil
}

func decodeEntry(client *datastreamer.StreamClient) error {
	err := client.Start()
	if err != nil {
		return err
	}

	entry, err := client.ExecCommandGetEntry(uint64(entryNum))
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	printEntry(entry)
	return nil
}

func decodeL2Block(client *datastreamer.StreamClient) error {
	err := client.Start()
	if err != nil {
		return err
	}

	bookmark := types.NewBookmarkProto(uint64(blockNum), datastream.BookmarkType_BOOKMARK_TYPE_L2_BLOCK)

	marshalledBookMark, err := bookmark.Marshal()
	if err != nil {
		return err
	}

	firstEntry, err := client.ExecCommandGetBookmark(marshalledBookMark)
	if err != nil {
		log.Error(err)
		return err
	}
	printEntry(firstEntry)

	secondEntry, err := client.ExecCommandGetEntry(firstEntry.Number + 1)
	if err != nil {
		log.Error(err)
		return err
	}

	i := uint64(2)
	for secondEntry.Type == datastreamer.EntryType(datastream.EntryType_ENTRY_TYPE_TRANSACTION) {
		printEntry(secondEntry)
		entry, err := client.ExecCommandGetEntry(firstEntry.Number + i)
		if err != nil {
			log.Error(err)
			return err
		}
		secondEntry = entry
		i++
	}

	return nil
}

func decodeBatch(client *datastreamer.StreamClient) error {
	err := client.Start()
	if err != nil {
		return err
	}

	bookmark := types.NewBookmarkProto(uint64(batchNum), datastream.BookmarkType_BOOKMARK_TYPE_BATCH)

	marshalledBookMark, err := bookmark.Marshal()
	if err != nil {
		return err
	}

	firstEntry, err := client.ExecCommandGetBookmark(marshalledBookMark)
	if err != nil {
		log.Error(err)
		return err
	}
	printEntry(firstEntry)

	secondEntry, err := client.ExecCommandGetEntry(firstEntry.Number + 1)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	printEntry(secondEntry)

	i := uint64(2) //nolint:gomnd
	for {
		entry, err := client.ExecCommandGetEntry(firstEntry.Number + i)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		if entry.Type == datastreamer.EntryType(types.BookmarkEntryType) {
			bookmark, err := types.UnmarshalBookmark(entry.Data)
			if err != nil {
				return err
			}
			if bookmark.BookmarkType() == datastream.BookmarkType_BOOKMARK_TYPE_BATCH {
				break
			}
		}

		secondEntry = entry
		printEntry(secondEntry)
		i++
	}

	return nil
}

func printEntry(entry datastreamer.FileEntry) {
	fmt.Println("entry type", entry.Type)
	switch entry.Type {
	case datastreamer.EntryType(types.BookmarkEntryType):
		bookmark, err := types.UnmarshalBookmark(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "BookMark\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "Type............: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d (%s)\n", bookmark.BookmarkType(), datastream.BookmarkType_name[int32(bookmark.BookmarkType())]))
		printColored(color.FgGreen, "Value...........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", bookmark.Value))
	case datastreamer.EntryType(types.EntryTypeL2Block):
		l2Block, err := types.UnmarshalL2Block(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "L2 Block\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "L2 Block Number.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", l2Block.L2BlockNumber))
		printColored(color.FgGreen, "Batch Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", l2Block.BatchNumber))
		printColored(color.FgGreen, "Timestamp.......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d (%v)\n", l2Block.Timestamp, time.Unix(l2Block.Timestamp, 0)))
		printColored(color.FgGreen, "Delta Timestamp.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", l2Block.DeltaTimestamp))
		printColored(color.FgGreen, "L1 Block Hash...: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.L1BlockHash))
		printColored(color.FgGreen, "L1 InfoTree Idx.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", l2Block.L1InfoTreeIndex))
		printColored(color.FgGreen, "Block Hash......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.L2Blockhash))
		printColored(color.FgGreen, "State Root......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.StateRoot))
		printColored(color.FgGreen, "Global Exit Root: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.GlobalExitRoot))
		printColored(color.FgGreen, "Coinbase........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.Coinbase))
		printColored(color.FgGreen, "Block Gas Limit.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", l2Block.BlockGasLimit))
		printColored(color.FgGreen, "Block Info Root.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.BlockInfoRoot))

		if l2Block.Debug.Message != "" {
			printColored(color.FgGreen, "Debug...........: ")
			printColored(color.FgHiWhite, fmt.Sprintf("%s\n", l2Block.Debug))
		}

	case datastreamer.EntryType(types.EntryTypeL2BlockEnd):
		block, err := types.UnmarshalL2BlockEnd(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "L2 Block End\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "L2 Block Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", block.Number))

	case datastreamer.EntryType(types.EntryTypeBatchStart):
		batch, err := types.UnmarshalBatchStart(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "Batch Start\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "Batch Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", batch.Number))
		printColored(color.FgGreen, "Batch Type......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", datastream.BatchType_name[int32(batch.BatchType)]))
		printColored(color.FgGreen, "Fork ID.........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", batch.ForkId))
		printColored(color.FgGreen, "Chain ID........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", batch.ChainId))

		if batch.Debug.Message != "" {
			printColored(color.FgGreen, "Debug...........: ")
			printColored(color.FgHiWhite, fmt.Sprintf("%s\n", batch.Debug))
		}

	case datastreamer.EntryType(types.EntryTypeBatchEnd):
		batch, err := types.UnmarshalBatchEnd(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "Batch End\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "Batch Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", batch.Number))
		printColored(color.FgGreen, "State Root......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", batch.StateRoot))
		printColored(color.FgGreen, "Local Exit Root.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", batch.LocalExitRoot))

		if batch.Debug.Message != "" {
			printColored(color.FgGreen, "Debug...........: ")
			printColored(color.FgHiWhite, fmt.Sprintf("%s\n", batch.Debug))
		}

	case datastreamer.EntryType(types.EntryTypeL2Tx):
		dsTx, err := types.UnmarshalTx(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "L2 Transaction\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "L2 Block Number.: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", dsTx.L2BlockNumber))
		printColored(color.FgGreen, "Index...........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", dsTx.Index))
		printColored(color.FgGreen, "Is Valid........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%t\n", dsTx.IsValid))
		printColored(color.FgGreen, "Data............: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", hexutil.Encode(dsTx.Encoded)))
		printColored(color.FgGreen, "Effec. Gas Price: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", dsTx.EffectiveGasPricePercentage))
		printColored(color.FgGreen, "IM State Root...: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", dsTx.IntermediateStateRoot))

		tx, err := ctypes.DecodeTransaction(dsTx.Encoded)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		nonce := tx.GetNonce()
		printColored(color.FgGreen, "Nonce...........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", nonce))

		if dsTx.Debug.Message != "" {
			printColored(color.FgGreen, "Debug...........: ")
			printColored(color.FgHiWhite, fmt.Sprintf("%s\n", dsTx.Debug))
		}

	case datastreamer.EntryType(types.EntryTypeGerUpdate):
		updateGer, err := types.DecodeGerUpdateProto(entry.Data)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		printColored(color.FgGreen, "Entry Type......: ")
		printColored(color.FgHiYellow, "Update GER\n")
		printColored(color.FgGreen, "Entry Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", entry.Number))
		printColored(color.FgGreen, "Batch Number....: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", updateGer.BatchNumber))
		printColored(color.FgGreen, "Timestamp.......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%v (%d)\n", time.Unix(int64(updateGer.Timestamp), 0), updateGer.Timestamp))
		printColored(color.FgGreen, "Global Exit Root: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", updateGer.GlobalExitRoot))
		printColored(color.FgGreen, "Coinbase........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", updateGer.Coinbase))
		printColored(color.FgGreen, "Fork ID.........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", updateGer.ForkId))
		printColored(color.FgGreen, "Chain ID........: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%d\n", updateGer.ChainId))
		printColored(color.FgGreen, "State Root......: ")
		printColored(color.FgHiWhite, fmt.Sprintf("%s\n", updateGer.StateRoot))

		if updateGer.Debug.Message != "" {
			printColored(color.FgGreen, "Debug...........: ")
			printColored(color.FgHiWhite, fmt.Sprintf("%s\n", updateGer.Debug))
		}
	}
}

func printColored(color color.Attribute, text string) {
	colored := fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, text)
	fmt.Print(colored)
}
