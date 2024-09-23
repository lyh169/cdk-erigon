package ethconfig

import (
	"encoding/json"
	"os"

	"github.com/ledgerwatch/erigon/core/types"
)

type Merlin struct {
	Headers []*types.Header `json:"headers"`
}

func ReadMerlinCfg(path string) (*Merlin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := Merlin{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

//func readFile(path string) ([]byte, error) {
//	jsonFile, err := os.Open(filepath.Clean(path))
//	if err != nil {
//		return nil, err
//	}
//	defer func() { _ = jsonFile.Close() }()
//
//	data, err := io.ReadAll(jsonFile)
//	if err != nil {
//		return nil, err
//	}
//	return data, nil
//}
