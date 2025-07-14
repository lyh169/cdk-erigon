package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ledgerwatch/erigon-lib/kv"
)

func ReadFile(path string) ([]byte, error) {
	jsonFile, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = jsonFile.Close() }()

	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func CheckSupportTable(tableName string) error {
	if tableName == kv.RecentLocalTransaction || tableName == kv.PoolTransaction {
		return nil
	}
	return fmt.Errorf("not support this table %s", tableName)
}
