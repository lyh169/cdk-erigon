package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	chain3 "github.com/ledgerwatch/erigon-lib/chain"
	common2 "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/core"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/turbo/debug"
	"github.com/ledgerwatch/erigon/zk/datastream/server"
	"github.com/ledgerwatch/log/v3"
	"github.com/spf13/cobra"
)

var datastream = &cobra.Command{
	Use: "datastream_unwind",
	Short: `revert datastream to target no.
Examples:
datastream_unwind --datadir=/datadirs/hermez-mainnet --unwind-block-no=2 --chain=hermez-bali --log.console.verbosity=4 --datadir-compare=/datadirs/pre-synced-block-100 # unwind to block 2
		`,
	Example: "go run ./cmd/integration datastream_unwind --config=... --verbosity=3 --unwind-block-no 2",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, _ := common2.RootContext()
		logger := debug.SetupCobra(cmd, "integration")
		db, err := openDB(dbCfg(kv.ChainDB, chaindata), true, logger)
		if err != nil {
			logger.Error("Opening DB", "error", err)
			return
		}
		defer db.Close()

		if err := unwindDatastream(ctx, db, logger); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error(err.Error())
			}
			return
		}
	},
}

func init() {
	withConfig(datastream)
	withChain(datastream)
	withDataDir2(datastream)
	withDsUnwindBlockNumber(datastream) // populates package global flag unwindBatchNo
	withUnwindBatchNo(datastream)
	rootCmd.AddCommand(datastream)
}

// unwindZk unwinds to the batch number set in the unwindBatchNo flag (package global)
func unwindDatastream(ctx context.Context, db kv.RwDB, logger log.Logger) error {
	fmt.Println(datadirCli, unwindDsBlockNo)
	dsFileName := path.Join(datadirCli, "data-stream.bin")
	var dataStreamServerFactory = server.NewZkEVMDataStreamServerFactory()
	ds, err := dataStreamServerFactory.CreateStreamServer(
		0,
		0,
		1,
		datastreamer.StreamType(1),
		dsFileName,
		1,
		2,
		3,
		nil,
	)
	if err != nil {
		fmt.Println("unwindDatastream", "error", err)
		return err
	}

	var genesis *types.Genesis
	if strings.HasPrefix(chain, "dynamic") {
		if config == "" {
			panic("Config file is required for dynamic chain")
		}

		params.DynamicChainConfigPath = filepath.Dir(config)
		genesis = core.GenesisBlockByChainName(chain)
		filename := path.Join(params.DynamicChainConfigPath, chain+"-conf.json")

		dConf := utils.DynamicConfig{}

		if _, err := os.Stat(filename); err == nil {
			dConfBytes, err := os.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(dConfBytes, &dConf); err != nil {
				panic(err)
			}
		}

		genesis.Timestamp = dConf.Timestamp
		genesis.GasLimit = dConf.GasLimit
		genesis.Difficulty = big.NewInt(dConf.Difficulty)
	} else {
		genesis = core.GenesisBlockByChainName(chain)
	}

	chainConfig, _, genesisErr := core.CommitGenesisBlock(db, genesis, "", logger)
	if _, ok := genesisErr.(*chain3.ConfigCompatError); genesisErr != nil && !ok {
		panic(genesisErr)
	}
	chainID := chainConfig.ChainID.Uint64()
	dataStreamServer := dataStreamServerFactory.CreateDataStreamServer(ds, chainID)
	if unwindDsBlockNo > 0 {
		dataStreamServer.UnwindToBlock(unwindDsBlockNo + 1) // todo if 10 needed use 11
	} else {
		dataStreamServer.UnwindToBatchStart(unwindBatchNo)
	}
	return nil
}
