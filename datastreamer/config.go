package datastreamer

import (
	"github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/0xPolygonHermez/zkevm-node/config/types"
)

// Config represents the configuration of a data streamer
type Config struct {
	// Port to listen on
	Port uint16 `mapstructure:"Port"`
	// Filename of the binary data file
	Filename string `mapstructure:"Filename"`
	// Log is the log configuration
	Log log.Config `mapstructure:"Log"`
	// WaitPeriodReadDB is the time the data streamer waits until
	WaitPeriodReadDB types.Duration `mapstructure:"WaitPeriodReadDB"`
	// MaxBlockLimit is the maximum number of blocks to be read from the DB
	MaxBlockLimit uint64 `mapstructure:"MaxBlockLimit"`
}
