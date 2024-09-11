package txpool

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/kv"
	"github.com/gateway-fm/cdk-erigon-lib/kv/kvcache"
	"github.com/gateway-fm/cdk-erigon-lib/types"
	"github.com/ledgerwatch/log/v3"
	"github.com/status-im/keycard-go/hexutils"
)

type LimboBatchPersistentHelper struct {
	keyBytesBatch            []byte
	keyBytesBatchUint64Array []byte
	keyBytesBlock            []byte
	keyBytesTx               []byte
	bytes8Value              []byte
}

func newLimboBatchPersistentHelper() *LimboBatchPersistentHelper {
	return &LimboBatchPersistentHelper{
		keyBytesBatch:            make([]byte, 9),
		keyBytesBatchUint64Array: make([]byte, 17),
		keyBytesBlock:            make([]byte, 13),
		keyBytesTx:               make([]byte, 17),
		bytes8Value:              make([]byte, 8),
	}
}

func (h *LimboBatchPersistentHelper) setBatchesIndex(unckeckedBatchIndex, invalidBatchIndex int) {
	binary.LittleEndian.PutUint32(h.keyBytesBatch[1:5], uint32(unckeckedBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesBatch[5:9], uint32(invalidBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesBatchUint64Array[1:5], uint32(unckeckedBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesBatchUint64Array[5:9], uint32(invalidBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesBlock[1:5], uint32(unckeckedBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesBlock[5:9], uint32(invalidBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesTx[1:5], uint32(unckeckedBatchIndex))
	binary.LittleEndian.PutUint32(h.keyBytesTx[5:9], uint32(invalidBatchIndex))
}

func (h *LimboBatchPersistentHelper) setBlockIndex(blockIndex int) {
	binary.LittleEndian.PutUint32(h.keyBytesBlock[9:13], uint32(blockIndex))
	binary.LittleEndian.PutUint32(h.keyBytesTx[9:13], uint32(blockIndex))
}

func (h *LimboBatchPersistentHelper) setTxIndex(txIndex int) {
	binary.LittleEndian.PutUint32(h.keyBytesTx[13:17], uint32(txIndex))
}

func (p *TxPool) flushLockedLimbo(tx kv.RwTx) (err error) {
	if !p.ethCfg.Limbo {
		return nil
	}

	if err := tx.CreateBucket(TablePoolLimbo); err != nil {
		return err
	}

	if err := tx.ClearBucket(TablePoolLimbo); err != nil {
		return err
	}

	for hash, handled := range p.limbo.invalidTxsMap {
		hashAsBytes := hexutils.HexToBytes(hash)
		key := append([]byte{DbKeyInvalidTxPrefix}, hashAsBytes...)
		tx.Put(TablePoolLimbo, key, []byte{handled})
	}

	v := make([]byte, 0, 1024)
	for i, txSlot := range p.limbo.limboSlots.Txs {
		v = common.EnsureEnoughSize(v, 20+len(txSlot.Rlp))
		sender := p.limbo.limboSlots.Senders.At(i)

		copy(v[:20], sender)
		copy(v[20:], txSlot.Rlp)

		key := append([]byte{DbKeySlotsPrefix}, txSlot.IDHash[:]...)
		if err := tx.Put(TablePoolLimbo, key, v); err != nil {
			return err
		}
	}

	limboBatchPersistentHelper := newLimboBatchPersistentHelper()
	for i, limboBatch := range p.limbo.uncheckedLimboBatches {
		limboBatchPersistentHelper.setBatchesIndex(i, math.MaxUint32)
		if err = flushLockedLimboBatch(tx, limboBatch, limboBatchPersistentHelper); err != nil {
			return err
		}
	}
	for i, limboBatch := range p.limbo.invalidLimboBatches {
		limboBatchPersistentHelper.setBatchesIndex(math.MaxUint32, i)
		if err = flushLockedLimboBatch(tx, limboBatch, limboBatchPersistentHelper); err != nil {
			return err
		}
	}

	v = []byte{0}
	if p.limbo.awaitingBlockHandling.Load() {
		v[0] = 1
	}
	if err := tx.Put(TablePoolLimbo, []byte{DbKeyAwaitingBlockHandlingPrefix}, v); err != nil {
		return err
	}

	return nil
}

func flushLockedLimboBatch(tx kv.RwTx, limboBatch *LimboBatchDetails, limboBatchPersistentHelper *LimboBatchPersistentHelper) error {
	keyBytesBatch := limboBatchPersistentHelper.keyBytesBatch
	keyBytesBatchUint64Array := limboBatchPersistentHelper.keyBytesBatchUint64Array
	keyBytesBlock := limboBatchPersistentHelper.keyBytesBlock
	keyBytesTx := limboBatchPersistentHelper.keyBytesTx
	bytes8Value := limboBatchPersistentHelper.bytes8Value

	// Witness
	keyBytesBatch[0] = DbKeyBatchesWitnessPrefix
	if err := tx.Put(TablePoolLimbo, keyBytesBatch, limboBatch.Witness); err != nil {
		return err
	}

	// L1InfoTreeMinTimestamps
	keyBytesBatch[0] = DbKeyBatchesL1InfoTreePrefix
	copy(keyBytesBatchUint64Array, keyBytesBatch)
	for k, v := range limboBatch.L1InfoTreeMinTimestamps {
		binary.LittleEndian.PutUint64(keyBytesBatchUint64Array[9:17], uint64(k))
		binary.LittleEndian.PutUint64(bytes8Value[:], v)
		if err := tx.Put(TablePoolLimbo, keyBytesBatchUint64Array, bytes8Value); err != nil {
			return err
		}
	}

	// BatchNumber
	keyBytesBatch[0] = DbKeyBatchesBatchNumberPrefix
	binary.LittleEndian.PutUint64(bytes8Value[:], limboBatch.BatchNumber)
	if err := tx.Put(TablePoolLimbo, keyBytesBatch, bytes8Value); err != nil {
		return err
	}

	// ForkId
	keyBytesBatch[0] = DbKeyBatchesForkIdPrefix
	binary.LittleEndian.PutUint64(bytes8Value[:], limboBatch.ForkId)
	if err := tx.Put(TablePoolLimbo, keyBytesBatch, bytes8Value); err != nil {
		return err
	}

	for j, limboBlock := range limboBatch.Blocks {
		limboBatchPersistentHelper.setBlockIndex(j)

		// Block - Block number
		keyBytesBlock[0] = DbKeyBlockNumber
		binary.LittleEndian.PutUint64(bytes8Value[:], limboBlock.BlockNumber)
		if err := tx.Put(TablePoolLimbo, keyBytesBlock, bytes8Value); err != nil {
			return err
		}

		// Block - Timestamp
		keyBytesBlock[0] = DbKeyBlockTimestamp
		binary.LittleEndian.PutUint64(bytes8Value[:], limboBlock.Timestamp)
		if err := tx.Put(TablePoolLimbo, keyBytesBlock, bytes8Value); err != nil {
			return err
		}

		for k, limboTx := range limboBlock.Transactions {
			limboBatchPersistentHelper.setTxIndex(k)

			// Transaction - Rlp
			keyBytesTx[0] = DbKeyTxRlpPrefix
			if err := tx.Put(TablePoolLimbo, keyBytesTx, limboTx.Rlp[:]); err != nil {
				return err
			}

			// Transaction - Stream bytes
			keyBytesTx[0] = DbKeyTxStreamBytesPrefix
			if err := tx.Put(TablePoolLimbo, keyBytesTx, limboTx.StreamBytes[:]); err != nil {
				return err
			}

			// Transaction - Root
			keyBytesTx[0] = DbKeyTxRootPrefix
			if err := tx.Put(TablePoolLimbo, keyBytesTx, limboTx.Root[:]); err != nil {
				return err
			}

			// Transaction - Hash
			keyBytesTx[0] = DbKeyTxHashPrefix
			if err := tx.Put(TablePoolLimbo, keyBytesTx, limboTx.Hash[:]); err != nil {
				return err
			}

			// Transaction - Sender
			keyBytesTx[0] = DbKeyTxSenderPrefix
			if err := tx.Put(TablePoolLimbo, keyBytesTx, limboTx.Sender[:]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *TxPool) fromDBLimbo(ctx context.Context, tx kv.Tx, cacheView kvcache.CacheView) error {
	if !p.ethCfg.Limbo {
		return nil
	}

	p.limbo.limboSlots = &types.TxSlots{}
	parseCtx := types.NewTxParseContext(p.chainID)
	parseCtx.WithSender(false)

	it, err := tx.Range(TablePoolLimbo, nil, nil)
	if err != nil {
		return err
	}

	for it.HasNext() {
		k, v, err := it.Next()
		if err != nil {
			return err
		}

		switch k[0] {
		case DbKeyInvalidTxPrefix:
			hash := hexutils.BytesToHex(k[1:])
			p.limbo.invalidTxsMap[hash] = v[0]
		case DbKeySlotsPrefix:
			addr, txRlp := *(*[20]byte)(v[:20]), v[20:]
			txn := &types.TxSlot{}

			_, err = parseCtx.ParseTransaction(txRlp, 0, txn, nil, false /* hasEnvelope */, nil)
			if err != nil {
				err = fmt.Errorf("err: %w, rlp: %x", err, txRlp)
				log.Warn("[txpool] fromDB: parseTransaction", "err", err)
				continue
			}

			txn.SenderID, txn.Traced = p.senders.getOrCreateID(addr)
			binary.BigEndian.Uint64(v)

			// ValidateTx function validates a tx against current network state.
			// Limbo transactions are expected to be invalid according to current network state.
			// That's why there is no point to check it while recovering the pool from a database.
			// These transactions may become valid after some of the current tx in the pool are executed
			// so leave the decision whether a limbo transaction (or any other transaction that has been unwound) to the execution stage.
			// if reason := p.validateTx(txn, true, cacheView, addr); reason != NotSet && reason != Success {
			// 	return nil
			// }
			p.limbo.limboSlots.Append(txn, addr[:], true)
		case DbKeyBatchesWitnessPrefix:
			fromDBLimboBatch(p, -1, -1, k, v).Witness = v
		case DbKeyBatchesL1InfoTreePrefix:
			l1InfoTreeKey := binary.LittleEndian.Uint64(k[9:17])
			fromDBLimboBatch(p, -1, -1, k, v).L1InfoTreeMinTimestamps[l1InfoTreeKey] = binary.LittleEndian.Uint64(v)
		case DbKeyBatchesBatchNumberPrefix:
			fromDBLimboBatch(p, -1, -1, k, v).BatchNumber = binary.LittleEndian.Uint64(v)
		case DbKeyBatchesForkIdPrefix:
			fromDBLimboBatch(p, -1, -1, k, v).ForkId = binary.LittleEndian.Uint64(v)
		case DbKeyBlockNumber:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			fromDBLimboBatch(p, int(blockIndex), -1, k, v).Blocks[blockIndex].BlockNumber = binary.LittleEndian.Uint64(v)
		case DbKeyBlockTimestamp:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			fromDBLimboBatch(p, int(blockIndex), -1, k, v).Blocks[blockIndex].Timestamp = binary.LittleEndian.Uint64(v)
		case DbKeyTxRlpPrefix:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			txIndex := binary.LittleEndian.Uint32(k[13:17])
			fromDBLimboBatch(p, int(blockIndex), int(txIndex), k, v).Blocks[blockIndex].Transactions[txIndex].Rlp = v
		case DbKeyTxStreamBytesPrefix:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			txIndex := binary.LittleEndian.Uint32(k[13:17])
			fromDBLimboBatch(p, int(blockIndex), int(txIndex), k, v).Blocks[blockIndex].Transactions[txIndex].StreamBytes = v
		case DbKeyTxRootPrefix:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			txIndex := binary.LittleEndian.Uint32(k[13:17])
			copy(fromDBLimboBatch(p, int(blockIndex), int(txIndex), k, v).Blocks[blockIndex].Transactions[txIndex].Root[:], v)
		case DbKeyTxHashPrefix:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			txIndex := binary.LittleEndian.Uint32(k[13:17])
			copy(fromDBLimboBatch(p, int(blockIndex), int(txIndex), k, v).Blocks[blockIndex].Transactions[txIndex].Hash[:], v)
		case DbKeyTxSenderPrefix:
			blockIndex := binary.LittleEndian.Uint32(k[9:13])
			txIndex := binary.LittleEndian.Uint32(k[13:17])
			copy(fromDBLimboBatch(p, int(blockIndex), int(txIndex), k, v).Blocks[blockIndex].Transactions[txIndex].Sender[:], v)
		case DbKeyAwaitingBlockHandlingPrefix:
			p.limbo.awaitingBlockHandling.Store(v[0] != 0)
		default:
			panic("Invalid key")
		}

	}

	return nil
}

func fromDBLimboBatch(p *TxPool, blockIndex, txIndex int, k, v []byte) *LimboBatchDetails {
	unckeckedBatchIndex := binary.LittleEndian.Uint32(k[1:5])
	invalidBatchIndex := binary.LittleEndian.Uint32(k[5:9])
	if unckeckedBatchIndex != math.MaxUint32 {
		p.limbo.resizeUncheckedBatches(int(unckeckedBatchIndex), blockIndex, txIndex)
		return p.limbo.uncheckedLimboBatches[unckeckedBatchIndex]
	}

	p.limbo.resizeInvalidBatches(int(invalidBatchIndex), blockIndex, txIndex)
	return p.limbo.invalidLimboBatches[invalidBatchIndex]
}
