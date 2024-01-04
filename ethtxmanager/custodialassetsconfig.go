package ethtxmanager

import "github.com/ethereum/go-ethereum/common"

type CustodialAssetsConfig struct {
	// Enable is the flag to enable the custodial assets
	Enable bool

	// URL is the url to sign the custodial assets
	URL string

	// Symbol is the symbol of the network, 2 prd, 2882 devnet
	Symbol int

	// SequencerAddr is the address of the sequencer
	SequencerAddr common.Address

	// AggregatorAddr is the address of the aggregator
	AggregatorAddr common.Address
}
