package ethtxmanager

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type CustodialAssetsConfig struct {
	// Enable is the flag to enable the custodial assets
	Enable bool `mapstructure:"Enable"`

	// URL is the url to sign the custodial assets
	URL string `mapstructure:"URL"`

	// Symbol is the symbol of the network, 2 prd, 2882 devnet
	Symbol int `mapstructure:"Symbol"`

	// SequencerAddr is the address of the sequencer
	SequencerAddr common.Address `mapstructure:"SequencerAddr"`

	// AggregatorAddr is the address of the aggregator
	AggregatorAddr common.Address `mapstructure:"AggregatorAddr"`

	// WaitResultTimeout is the timeout to wait for the result of the custodial assets
	WaitResultTimeout time.Duration `mapstructure:"WaitResultTimeout"`
}
