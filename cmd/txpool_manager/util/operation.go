package util

import (
	"context"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/hexutil"
	"github.com/ledgerwatch/log/v3"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/types"
)

func List(ctx context.Context, db kv.RwDB, table string, chainID uint64, countOutput int) ([]*Tx, error) {
	if countOutput == 0 {
		return nil, nil
	}
	tx, err := db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parseCtx := types.NewTxParseContext(*uint256.NewInt(chainID))
	parseCtx.WithSender(false)

	i := 0
	it, err := tx.Range(table, nil, nil)
	if err != nil {
		return nil, err
	}
	var txs []*Tx
	for it.HasNext() {
		k, v, err := it.Next()
		if err != nil {
			return nil, err
		}
		addr, txRlp := *(*[20]byte)(v[:20]), v[20:]
		txn := &types.TxSlot{}

		_, err = parseCtx.ParseTransaction(txRlp, 0, txn, nil, false /* hasEnvelope */, false, nil)
		if err != nil {
			err = fmt.Errorf("err: %w, rlp: %x", err, txRlp)
			log.Warn("[txpool] fromDB: parseTransaction", " err ", err)
			continue
		}

		tx := &Tx{
			TxSlotJ: CovertToTxSlotJson(txn),
			Sender:  common.BytesToAddress(addr[:]),
			RawData: genKV(k, v),
		}

		txs = append(txs, tx)
		i++
		if countOutput >= 0 && i == countOutput {
			break
		}
	}
	return txs, nil
}

func ListRecentLocalTransaction(ctx context.Context, db kv.RwDB, countOutput int) ([]*KV, error) {
	if countOutput == 0 {
		return nil, nil
	}
	tx, err := db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	it, err := tx.Range(kv.RecentLocalTransaction, nil, nil)
	if err != nil {
		return nil, err
	}
	i := 0
	var kvs []*KV
	for it.HasNext() {
		k, v, err := it.Next()
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, genKV(k, v))
		i++
		if countOutput >= 0 && i == countOutput {
			break
		}
	}
	return kvs, nil
}

func CovertToTxSlotJson(txslot *types.TxSlot) *TxSlotJson {
	return &TxSlotJson{
		Rlp:            hexutil.Encode(txslot.Rlp),
		Value:          txslot.Value,
		Tip:            txslot.Tip,
		FeeCap:         txslot.FeeCap,
		SenderID:       txslot.SenderID,
		Nonce:          txslot.Nonce,
		DataLen:        txslot.DataLen,
		DataNonZeroLen: txslot.DataNonZeroLen,
		AlAddrCount:    txslot.AlAddrCount,
		AlStorCount:    txslot.AlStorCount,
		Gas:            txslot.Gas,
		IDHash:         common.BytesToHash(txslot.IDHash[:]),
		Traced:         txslot.Traced,
		Creation:       txslot.Creation,
		Type:           txslot.Type,
		Size:           txslot.Size,
		BlobFeeCap:     txslot.BlobFeeCap,
		BlobHashes:     txslot.BlobHashes,
		Blobs:          txslot.Blobs,
		Commitments:    txslot.Commitments,
		Proofs:         txslot.Proofs,
		To:             txslot.To,
	}
}

func genKV(k, v []byte) *KV {
	return &KV{hexutil.Encode(k), hexutil.Encode(v)}
}

func Add(ctx context.Context, db kv.RwDB, tableName string, kvs []*KV) (int, error) {
	count := 0
	if err := db.Update(ctx, func(tx kv.RwTx) error {
		for _, lkv := range kvs {
			if lkv.Key == "" {
				log.Warn("[txpool] Add tx to txpool fail key is empty")
				continue
			}
			key := common.FromHex(lkv.Key)
			value := common.FromHex(lkv.Value)
			has, err := tx.Has(tableName, key)
			if err != nil || has {
				log.Warn("[txpool] Add tx to txpool fail", " tx keys ", lkv.Key, "have in database", has, " err ", err)
				continue
			}

			if err := tx.Put(tableName, key, value); err != nil {
				log.Error("[txpool] Add tx to txpool fail", " txhash ", lkv.Key, " err ", err)
				return err
			}
			count++
		}
		return nil
	}); err != nil {
		return count, err
	}
	return count, nil
}

func DeleteAll(ctx context.Context, db kv.RwDB, table string) (int, error) {
	tx, err := db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	it, err := tx.Range(table, nil, nil)
	if err != nil {
		return 0, err
	}
	var txHashs []string
	for it.HasNext() {
		k, _, err := it.Next()
		if err != nil {
			return 0, err
		}
		txHashs = append(txHashs, hexutil.Encode(k))
	}
	if len(txHashs) == 0 {
		return 0, nil
	}
	return Delete(ctx, db, table, txHashs)
}

func Delete(ctx context.Context, db kv.RwDB, table string, txHashs []string) (int, error) {
	count := 0
	if err := db.Update(ctx, func(tx kv.RwTx) error {
		for _, txHash := range txHashs {
			if txHash == "" {
				log.Warn("[txpool] Delete tx in txpool fail key is empty")
				continue
			}
			hash, err := hexutil.Decode(txHash)
			if err != nil {
				log.Warn("[txpool] Delete txpool hash fail ", " txhash ", txHash, " Decode err ", err)
				continue
			}
			has, err := tx.Has(table, hash)
			if err != nil || !has {
				log.Warn("[txpool] Delete txpool hash fail ", " txhash key ", txHash, "is have", has, " err ", err)
				continue
			}
			if has {
				if err := tx.Delete(table, hash); err != nil {
					log.Warn("[txpool] Delete txpool hash fail ", " txhash ", txHash, " err ", err)
					continue
				}
				count++
			}
		}
		return nil
	}); err != nil {
		return count, err
	}
	return count, nil
}
