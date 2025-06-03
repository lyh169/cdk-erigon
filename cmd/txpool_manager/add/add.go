package add

import (
	"encoding/json"
	"errors"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cmd/txpool_manager/util"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/urfave/cli/v2"
)

var (
	inputFile string
	tableName string
)

var Command = cli.Command{
	Action: run,
	Name:   "add",
	Usage:  "Add the tx in txpool at the txpool_manager",
	Flags: []cli.Flag{
		&utils.DataDirFlag,
		&cli.StringFlag{
			Name:        "inputFile",
			Usage:       "Will add tx hashs or keys in the file",
			Destination: &inputFile,
		},
		&cli.StringFlag{
			Name:        "table",
			Usage:       "The db table name",
			Destination: &tableName,
		},
	},
}

func run(cliCtx *cli.Context) error {
	dataDir := cliCtx.String(utils.DataDirFlag.Name)
	if tableName == "" {
		tableName = kv.PoolTransaction
	}
	err := util.CheckSupportTable(tableName)
	if err != nil {
		return err
	}

	if inputFile == "" {
		log.Error("The flag inputFile is not set ")
		return errors.New("The flag inputFile is not set")
	}

	poolDB, err := util.OpenTxpoolDB(cliCtx.Context, dataDir)
	if err != nil {
		log.Error("Failed to open txpool database ", " err ", err)
		return err
	}

	var kvs []*util.KV
	if tableName == kv.PoolTransaction {
		kvs, err = ReadAddPoolTransactionFile(inputFile)
		if err != nil {
			log.Error("Failed to ReadAddPoolTransactionFile ", " err ", err, " inputFile ", inputFile)
			return err
		}
	} else if tableName == kv.RecentLocalTransaction {
		kvs, err = ReadAddRecentLocalTransactionFile(inputFile)
		if err != nil {
			log.Error("Failed to ReadAddRecentLocalTransactionFile ", " err ", err, " inputFile ", inputFile)
			return err
		}
	}

	log.Info("Add ", "dataDir ", dataDir, " inputFile ", inputFile, " will add count ", len(kvs))

	count, err := util.Add(cliCtx.Context, poolDB, tableName, kvs)
	if err != nil {
		log.Error("Failed to add txs in the database ", " err ", err)
		return err
	}

	log.Info("The all add txs in database ", " count ", count)
	return nil
}

func ReadAddPoolTransactionFile(path string) ([]*util.KV, error) {
	data, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var txs []*util.Tx
	err = json.Unmarshal(data, &txs)
	if err != nil {
		return nil, err
	}
	kvs := make([]*util.KV, len(txs))
	for i, tx := range txs {
		if tx.RawData == nil {
			return nil, errors.New("the data is error or the table is incorrect")
		}
		kvs[i] = tx.RawData
	}
	return kvs, nil
}

func ReadAddRecentLocalTransactionFile(path string) ([]*util.KV, error) {
	data, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var kvs []*util.KV
	err = json.Unmarshal(data, &kvs)
	if err != nil {
		return nil, err
	}
	return kvs, nil
}
