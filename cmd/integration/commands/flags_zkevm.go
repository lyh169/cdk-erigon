package commands

import "github.com/spf13/cobra"

var (
	unwindBatchNo   uint64
	unwindDsBlockNo uint64
)

func withUnwindBatchNo(cmd *cobra.Command) {
	cmd.Flags().Uint64Var(&unwindBatchNo, "unwind-batch-no", 0, "batch number to unwind to (this batch number will be the tip after unwind)")
}

func withDsUnwindBlockNumber(cmd *cobra.Command) {
	cmd.Flags().Uint64Var(&unwindDsBlockNo, "unwind-block-no", 0, "block number to unwind to (this block number will be the tip)")
}
