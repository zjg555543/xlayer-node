package jsonrpc

import (
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/jsonrpc/types"
)

// ApolloConfig is the apollo RPC dynamic config
type ApolloConfig struct {
	EnableApollo         bool     `json:"enable"`
	BatchRequestsEnabled bool     `json:"batchRequestsEnabled"`
	BatchRequestsLimit   uint     `json:"batchRequestsLimit"`
	GasLimitFactor       float64  `json:"gasLimitFactor"`
	DisableAPIs          []string `json:"disableAPIs"`

	sync.RWMutex
}

var apolloConfig = &ApolloConfig{}

// GetInstance returns the singleton instance
func GetInstance() *ApolloConfig {
	return apolloConfig
}

// Enable returns true if apollo is enabled
func (c *ApolloConfig) Enable() bool {
	if c == nil || !c.EnableApollo {
		return false
	}
	c.RLock()
	defer c.RUnlock()
	return c.EnableApollo
}

func (c *ApolloConfig) setDisableAPIs(disableAPIs []string) {
	if c == nil || !c.EnableApollo {
		return
	}
	c.DisableAPIs = make([]string, len(disableAPIs))
	copy(c.DisableAPIs, disableAPIs)
}

// UpdateConfig updates the apollo config
func UpdateConfig(apolloConfig Config) {
	GetInstance().Lock()
	GetInstance().EnableApollo = true
	GetInstance().BatchRequestsEnabled = apolloConfig.BatchRequestsEnabled
	GetInstance().BatchRequestsLimit = apolloConfig.BatchRequestsLimit
	GetInstance().GasLimitFactor = apolloConfig.GasLimitFactor
	GetInstance().setDisableAPIs(apolloConfig.DisableAPIs)
	GetInstance().Unlock()
}

func (e *EthEndpoints) isDisabled(rpc string) bool {
	if GetInstance().Enable() {
		GetInstance().RLock()
		defer GetInstance().RUnlock()
		return len(GetInstance().DisableAPIs) > 0 && types.Contains(GetInstance().DisableAPIs, rpc)
	}

	return len(e.cfg.DisableAPIs) > 0 && types.Contains(e.cfg.DisableAPIs, rpc)
}
