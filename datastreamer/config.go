package datastreamer

import (
	"github.com/0xPolygonHermez/zkevm-node/config/types"
)

// Config represents the configuration of a data streamer
type Config struct {
	// Port to listen on
	Port uint16 `mapstructure:"Port"`
	// Filename of the binary data file
	Filename string `mapstructure:"Filename"`
	// WaitInterval is the time the data streamer waits until
	WaitInterval types.Duration `mapstructure:"WaitInterval"`
	// MaxBatchLimit is the maximum number of blocks to be read from the DB
	MaxBatchLimit uint64 `mapstructure:"MaxBatchLimit"`
}
