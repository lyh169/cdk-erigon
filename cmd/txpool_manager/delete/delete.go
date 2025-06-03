package delete

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cmd/txpool_manager/util"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/zkevm/log"
	"github.com/urfave/cli/v2"
)

var (
	inputFile    string
	delTxKeysStr string
	tableName    string
	isDeleteAll  bool
)

var Command = cli.Command{
	Action: run,
	Name:   "delete",
	Usage:  "Delete the tx in txpool at the txpool_manager",
	Flags: []cli.Flag{
		&utils.DataDirFlag,
		&cli.StringFlag{
			Name:        "inputFile",
			Usage:       "Will deleting tx hashs or keys in the file",
			Destination: &inputFile,
		},
		&cli.StringFlag{
			Name:        "delTxKeys",
			Usage:       "Will deleting tx hashs or keys in the txpool, the format is: [txhash1,txhash2]",
			Destination: &delTxKeysStr,
		},
		&cli.StringFlag{
			Name:        "table",
			Usage:       "The db table name",
			Destination: &tableName,
		},
		&cli.BoolFlag{
			Name:        "all",
			Usage:       "Delete all the table of data",
			Destination: &isDeleteAll,
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

	var delKeys []string
	if !isDeleteAll {
		if len(delTxKeysStr) == 0 && inputFile == "" {
			log.Error("The flag delTxKeys and inputFile is not set")
			return errors.New("The flag delTxKeys and inputFile is not set")
		} else if len(delTxKeysStr) != 0 && inputFile != "" {
			log.Error("The flag delTxKeys and inputFile should only set one")
			return errors.New("The flag delTxKeys and inputFile should only set one")
		}
		delTxKeysStr = strings.Trim(delTxKeysStr, "[]")
		delKeys = strings.Split(delTxKeysStr, ",")
		for i := 0; i < len(delKeys); i++ {
			delKeys[i] = strings.TrimSpace(delKeys[i])
		}
		if inputFile != "" {
			if tableName == kv.PoolTransaction {
				delKeys, err = ReadDeletePoolTransactionFile(inputFile)
				if err != nil {
					log.Error("Failed to ReadDeleteFile ", " err ", err, " inputFile ", inputFile)
					return err
				}
			} else if tableName == kv.RecentLocalTransaction {
				delKeys, err = ReadDeleteRecentLocalTransactionFile(inputFile)
				if err != nil {
					log.Error("Failed to ReadDeleteFile ", " err ", err, " inputFile ", inputFile)
					return err
				}
			}
		}
	}

	log.Info("Listing ", "dataDir ", dataDir, " inputFile ", inputFile, " delKeys count ", len(delKeys))

	poolDB, err := util.OpenTxpoolDB(cliCtx.Context, dataDir)
	if err != nil {
		log.Error("Failed to open txpool database ", " err ", err)
		return err
	}
	var count int
	if isDeleteAll {
		count, err = util.DeleteAll(cliCtx.Context, poolDB, tableName)
		if err != nil {
			log.Error("Failed to Delete the database ", " err ", err)
			return err
		}
	} else {
		count, err = util.Delete(cliCtx.Context, poolDB, tableName, delKeys)
		if err != nil {
			log.Error("Failed to Delete the database ", " err ", err)
			return err
		}
	}
	log.Info("The all delete txs in database ", " count ", count)
	return nil
}

func ReadDeletePoolTransactionFile(path string) ([]string, error) {
	data, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var txs []*util.Tx
	var txHashs []string
	err = json.Unmarshal(data, &txs)
	if err != nil {
		err = json.Unmarshal(data, &txHashs)
		if err != nil {
			return nil, err
		}
		return txHashs, err
	}
	for _, tx := range txs {
		if tx.RawData == nil {
			return nil, errors.New("the data is error or the table is incorrect")
		}
		txHashs = append(txHashs, tx.RawData.Key)
	}
	return txHashs, nil
}

func ReadDeleteRecentLocalTransactionFile(path string) ([]string, error) {
	data, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var kvs []*util.KV
	err = json.Unmarshal(data, &kvs)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, kv := range kvs {
		keys = append(keys, kv.Key)
	}
	return keys, nil
}
