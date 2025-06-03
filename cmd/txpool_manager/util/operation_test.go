package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/stretchr/testify/require"
)

func newTestTxpoolDB(tb testing.TB) kv.RwDB {
	tb.Helper()

	dir := fmt.Sprintf("/tmp/pool-db-temp_%v", time.Now().UTC().Format(time.RFC3339Nano))
	err := os.Mkdir(dir, 0775)

	if err != nil {
		tb.Fatal(err)
	}

	kvdb, err := OpenTxpoolDB(context.Background(), dir)
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			tb.Fatal(err)
		}
	})

	return kvdb
}

func TestAdd(t *testing.T) {
	tableName := kv.PoolTransaction
	testCases := []struct {
		name  string
		kvs   []*KV
		count int
	}{
		{
			name: "add_one",
			kvs: []*KV{
				{
					Key:   "0x0f94cba5054b22f56af1d456faa35ed481d44dff5a741f1522838d0ee45f1c9c",
					Value: "0xb9ce0c6f1dfd196a3890f0db994f9cf2c4606042f86d068435a4e90082520894c422419b895e728e359e255aa444744143401627881bc16d674ec80000808207f5a0aee5aba2682f50bc00dfac1c6462e75810a0bbc0d994203a7009dc2830d44edba02369a76f92ac26abf16479ab4f3a01359ee0831a13373f7df78cc93480bfd680",
				}},
			count: 1,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kvdb := newTestTxpoolDB(t)
			defer kvdb.Close()
			count, err := Add(context.Background(), kvdb, tableName, tc.kvs)
			require.NoError(t, err)
			require.Equal(t, tc.count, count)
		})
	}
}

func TestList(t *testing.T) {
	tableName := kv.PoolTransaction
	kvs, err := ReadTxpoolFile("../tests/txpool_data.json")
	require.NoError(t, err)
	kvdb := newTestTxpoolDB(t)
	defer kvdb.Close()
	_, err = Add(context.Background(), kvdb, tableName, kvs)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		kvs       []*KV
		chainID   uint64
		qcount    int
		realCount int
		txHashs   map[common.Hash]struct{}
		haveErr   bool
	}{
		{
			name:    "list_one_err",
			chainID: 1,
			qcount:  1,
			haveErr: true,
		},
		{
			name:    "list_one",
			chainID: 1001,
			qcount:  1,
			haveErr: false,
		},
		{
			name:    "list_two",
			chainID: 1001,
			qcount:  2,
			haveErr: false,
		},
		{
			name:    "list_all",
			chainID: 1001,
			qcount:  -1,
			haveErr: false,
			txHashs: map[common.Hash]struct{}{
				common.HexToHash("0x2edd0e406600e585ffbe03ffff30f1db6e2e81dd357d5a97d8d6132446ba65bd"): {},
			},
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txs, err := List(context.Background(), kvdb, kv.PoolTransaction, tc.chainID, tc.qcount)
			if tc.haveErr {
				require.Equal(t, len(txs), 0)
			} else {
				require.NoError(t, err)
				if tc.qcount != -1 {
					require.Equal(t, len(txs), tc.qcount)
				} else {
					require.NotEqual(t, len(txs), tc.qcount)
				}
				if len(tc.txHashs) != 0 {
					var isHave bool
					for _, tx := range txs {
						if _, ok := tc.txHashs[common.HexToHash(tx.RawData.Key)]; ok && tx.TxSlotJ.IDHash == common.HexToHash(tx.RawData.Key) {
							isHave = true
						}
					}
					require.True(t, isHave)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	chainID := uint64(1001)
	tableName := kv.PoolTransaction
	testCases := []struct {
		name         string
		kvs          []*KV
		dcount       int
		realCount    int
		deltxHashs   []string
		realdelHashs []string
		haveErr      bool
	}{
		{
			name:    "delete_one",
			dcount:  1,
			haveErr: false,
			deltxHashs: []string{
				"0x2edd0e406600e585ffbe03ffff30f1db6e2e81dd357d5a97d8d6132446ba65bd",
			},
			realdelHashs: []string{
				"0x2edd0e406600e585ffbe03ffff30f1db6e2e81dd357d5a97d8d6132446ba65bd",
			},
		},
		{
			name:    "delete_two",
			dcount:  2,
			haveErr: false,
			deltxHashs: []string{
				"0x2edd0e406600e585ffbe03ffff30f1db6e2e81dd357d5a97d8d6132446ba65bd",
				"0x3dd27860d93ea736db6a3bea51d39fdfc99cd61e69b1c64cda6361750db6a0e4",
			},
			realdelHashs: []string{
				"0x2edd0e406600e585ffbe03ffff30f1db6e2e81dd357d5a97d8d6132446ba65bd",
				"0x3dd27860d93ea736db6a3bea51d39fdfc99cd61e69b1c64cda6361750db6a0e4",
			},
		},
		{
			name:    "delete_one_not_exist",
			dcount:  0,
			haveErr: true,
			deltxHashs: []string{
				"0x8ccc5b3cbdcb23aca79f99933d21b7e364d763c30e168fe4a44e56567a27ad49",
			},
		},
		{
			name:    "delete_two_not_exist",
			dcount:  0,
			haveErr: true,
			deltxHashs: []string{
				"0x8ccc5b3cbdcb23aca79f99933d21b7e364d763c30e168fe4a44e56567a27ad49",
				"0x8ccc5b3cbdcb23aca79f99933d21b7e364d763c30e168fe4a44e56567a27ad4a",
			},
		},
		{
			name:    "delete_some_of_not_exist",
			dcount:  1,
			haveErr: true,
			deltxHashs: []string{
				"0x8ccc5b3cbdcb23aca79f99933d21b7e364d763c30e168fe4a44e56567a27ad49",
				"0x3dd27860d93ea736db6a3bea51d39fdfc99cd61e69b1c64cda6361750db6a0e4",
			},
			realdelHashs: []string{
				"0x3dd27860d93ea736db6a3bea51d39fdfc99cd61e69b1c64cda6361750db6a0e4",
			},
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// init db
			kvs, err := ReadTxpoolFile("../tests/txpool_data.json")
			require.NoError(t, err)
			kvdb := newTestTxpoolDB(t)
			defer kvdb.Close()
			_, err = Add(context.Background(), kvdb, tableName, kvs)
			require.NoError(t, err)

			count, err := Delete(context.Background(), kvdb, tableName, tc.deltxHashs)
			require.NoError(t, err)
			require.Equal(t, tc.dcount, count)

			txsMp, err := ListToMap(kvdb, tableName, chainID, -1)
			require.NoError(t, err)

			if len(tc.realdelHashs) != 0 {
				for _, hash := range tc.realdelHashs {
					if _, ok := txsMp[common.HexToHash(hash)]; ok {
						require.Errorf(t, errors.New("deleted hash should not exist"), "")
					}
				}
			}
		})
	}
}

func ListToMap(kvdb kv.RwDB, table string, chainID uint64, countOutput int) (map[common.Hash]*Tx, error) {
	txs, err := List(context.Background(), kvdb, table, chainID, countOutput)
	if err != nil {
		return nil, err
	}
	mp := make(map[common.Hash]*Tx, len(txs))
	for _, tx := range txs {
		mp[tx.TxSlotJ.IDHash] = tx
	}
	return mp, nil
}
func ReadTxpoolFile(path string) ([]*KV, error) {
	data, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	var txs []*Tx
	err = json.Unmarshal(data, &txs)
	if err != nil {
		return nil, err
	}
	kvs := make([]*KV, len(txs))
	for i, tx := range txs {
		kvs[i] = tx.RawData
	}
	return kvs, nil
}
