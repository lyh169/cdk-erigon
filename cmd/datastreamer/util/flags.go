package util

import (
	"github.com/urfave/cli/v2"
)

var (
	CfgFileFlag = cli.StringFlag{
		Name:     "cfg",
		Usage:    "config path",
		Required: true,
	}
)
