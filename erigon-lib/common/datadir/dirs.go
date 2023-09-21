/*
   Copyright 2021 Erigon contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package datadir

import (
	"github.com/ledgerwatch/erigon-lib/common/dir"
	"path/filepath"
)

// Dirs is the file system folder the node should use for any data storage
// requirements. The configured data directory will not be directly shared with
// registered services, instead those can use utility methods to create/access
// databases or flat files
type Dirs struct {
	DataDir         string
	RelativeDataDir string // like dataDir, but without filepath.Abs() resolution
	Chaindata       string
	Tmp             string
	Snap            string
	SnapIdx         string
	SnapHistory     string
	SnapDomain      string
	SnapAccessors   string
	TxPool          string
	Nodes           string
}

func New(datadir string) Dirs {
	relativeDataDir := datadir
	if datadir != "" {
		var err error
		absdatadir, err := filepath.Abs(datadir)
		if err != nil {
			panic(err)
		}
		datadir = absdatadir
	}

	dirs := Dirs{
		RelativeDataDir: relativeDataDir,
		DataDir:         datadir,
		Chaindata:       filepath.Join(datadir, "chaindata"),
		Tmp:             filepath.Join(datadir, "temp"),
		Snap:            filepath.Join(datadir, "snapshots"),
		SnapIdx:         filepath.Join(datadir, "snapshots", "idx"),
		SnapHistory:     filepath.Join(datadir, "snapshots", "history"),
		SnapDomain:      filepath.Join(datadir, "snapshots", "domain"),
		SnapAccessors:   filepath.Join(datadir, "snapshots", "accessor"),
		TxPool:          filepath.Join(datadir, "txpool"),
		Nodes:           filepath.Join(datadir, "nodes"),
	}
	dir.MustExist(dirs.Chaindata, dirs.Tmp,
		dirs.SnapIdx, dirs.SnapHistory, dirs.SnapDomain, dirs.SnapAccessors,
		dirs.TxPool, dirs.Nodes)
	return dirs
}