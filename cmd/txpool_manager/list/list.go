package list

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cmd/txpool_manager/util"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/urfave/cli/v2"
)

var (
	countOutput int
	chainID     uint64
	queryTable  string
	outputDir   string
)

var Command = cli.Command{
	Action: run,
	Name:   "list",
	Usage:  "List the content at the txpool_manager",
	Flags: []cli.Flag{
		&utils.DataDirFlag,
		&cli.IntFlag{
			Name:        "listCount",
			Usage:       "Number of transactions will be list, default is all",
			Destination: &countOutput,
			Value:       -1,
		},
		&cli.Uint64Flag{
			Name:        "chainID",
			Usage:       "The chain id",
			Destination: &chainID,
		},
		&cli.StringFlag{
			Name:        "table",
			Usage:       "The db table name",
			Destination: &queryTable,
		},
		&cli.StringFlag{
			Name:        "outputDir",
			Usage:       "The out put fail dir",
			Destination: &outputDir,
		},
	},
}

func run(cliCtx *cli.Context) error {
	dataDir := cliCtx.String(utils.DataDirFlag.Name)
	if queryTable == "" {
		queryTable = kv.PoolTransaction
	}
	err := util.CheckSupportTable(queryTable)
	if err != nil {
		return err
	}

	if chainID == 0 {
		log.Error("The flag chainID is not set")
		return errors.New("The chainID is not set")
	}
	if outputDir == "" {
		log.Error("The flag outputDir is not set")
		return errors.New("The outputDir is not set")
	}

	log.Info("Listing ", " dataDir ", dataDir, " table ", queryTable, " chainID ", chainID)

	poolDB, err := util.OpenTxpoolDB(cliCtx.Context, dataDir)
	if err != nil {
		log.Error("Failed to open txpool database ", " err ", err)
		return err
	}

	var count int
	if queryTable == kv.RecentLocalTransaction {
		count, err = ListAndWriteRecentLocalTransaction(cliCtx.Context, poolDB, queryTable, countOutput)
	} else if queryTable == kv.PoolTransaction {
		count, err = ListAndWritePoolTransaction(cliCtx.Context, poolDB, queryTable, chainID, countOutput)
	}
	if err != nil {
		log.Error("Failed to List the database ", " err ", err)
		return err
	}
	log.Info("The all List txs in database ", " count ", count)
	return nil
}

func ListAndWriteRecentLocalTransaction(ctx context.Context, poolDB kv.RwDB, table string, count int) (int, error) {
	kvs, err := util.ListRecentLocalTransaction(ctx, poolDB, count)
	if err != nil {
		log.Error("Failed to List the database ", " err ", err)
		return 0, err
	}
	timeNow := time.Now().Unix()
	fileName := fmt.Sprintf("%s-tx-%v.json", table, timeNow)
	err = WriteListKVSToFile(outputDir, fileName, kvs)
	if err != nil {
		log.Error("Failed to WriteListKVSToFile ", " err ", err)
		return 0, err
	}
	return len(kvs), nil
}

func WriteListKVSToFile(path, fileName string, kvs []*util.KV) error {
	name := filepath.Join(path, fileName)
	data, err := json.MarshalIndent(kvs, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, 0600)
}

func ListAndWritePoolTransaction(ctx context.Context, poolDB kv.RwDB, table string, chainID uint64, countOutput int) (int, error) {
	txs, err := util.List(ctx, poolDB, table, chainID, countOutput)
	if err != nil {
		log.Error("Failed to List the database ", " err ", err)
		return 0, err
	}
	timeNow := time.Now().Unix()
	fileName := fmt.Sprintf("%s-tx-%v.json", table, timeNow)
	err = WriteListTxsToFile(outputDir, fileName, txs)
	if err != nil {
		log.Error("Failed to WriteListTxsToFile ", " err ", err)
		return 0, err
	}
	return len(txs), nil
}

func WriteListTxsToFile(path, fileName string, txs []*util.Tx) error {
	name := filepath.Join(path, fileName)
	data, err := json.MarshalIndent(txs, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, 0600)
}
