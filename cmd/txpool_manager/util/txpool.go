package util

import (
	"context"
	"time"

	"github.com/c2h5oh/datasize"
	gokzg4844 "github.com/crate-crypto/go-kzg-4844"
	mdbx2 "github.com/erigontech/mdbx-go/mdbx"
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/log/v3"
)

type Tx struct {
	TxSlotJ *TxSlotJson
	Sender  common.Address
	RawData *KV
}

type KV struct {
	Key   string
	Value string
}

type TxSlotJson struct {
	Rlp            string      // Is set to nil after flushing to db, frees memory, later we look for it in the db, if needed
	Value          uint256.Int // Value transferred by the transaction
	Tip            uint256.Int // Maximum tip that transaction is giving to miner/block proposer
	FeeCap         uint256.Int // Maximum fee that transaction burns and gives to the miner/block proposer
	SenderID       uint64      // SenderID - require external mapping to it's address
	Nonce          uint64      // Nonce of the transaction
	DataLen        int         // Length of transaction's data (for calculation of intrinsic gas)
	DataNonZeroLen int
	AlAddrCount    int         // Number of addresses in the access list
	AlStorCount    int         // Number of storage keys in the access list
	Gas            uint64      // Gas limit of the transaction
	IDHash         common.Hash // Transaction hash for the purposes of using it as a transaction Id
	Traced         bool        // Whether transaction needs to be traced throughout transaction pool code and generate debug printing
	Creation       bool        // Set to true if "To" field of the transaction is not set
	Type           byte        // Transaction type
	Size           uint32      // Size of the payload (without the RLP string envelope for typed transactions)

	// EIP-4844: Shard Blob Transactions
	BlobFeeCap  uint256.Int // max_fee_per_blob_gas
	BlobHashes  []common.Hash
	Blobs       [][]byte
	Commitments []gokzg4844.KZGCommitment
	Proofs      []gokzg4844.KZGProof
	To          common.Address
}

func OpenTxpoolDB(ctx context.Context, dbDir string) (kv.RwDB, error) {
	txPoolDB, err := mdbx.NewMDBX(log.New()).Label(kv.TxPoolDB).Path(dbDir).
		WithTableCfg(func(defaultBuckets kv.TableCfg) kv.TableCfg { return kv.TxpoolTablesCfg }).
		Flags(func(f uint) uint { return f ^ mdbx2.Durable | mdbx2.SafeNoSync }).
		GrowthStep(16 * datasize.MB).
		SyncPeriod(30 * time.Second).
		Open(ctx)
	if err != nil {
		return nil, err
	}

	return txPoolDB, nil
}
