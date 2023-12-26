package jsonrpc

import (
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/jsonrpc/types"
)

type ApolloConfig struct {
	EnableApollo         bool     `json:"enable"`
	BatchRequestsEnabled bool     `json:"batchRequestsEnabled"`
	BatchRequestsLimit   uint     `json:"batchRequestsLimit"`
	GasLimitFactor       float64  `json:"gasLimitFactor"`
	DisableAPIs          []string `json:"disableAPIs"`

	sync.RWMutex
}

var apolloConfig = &ApolloConfig{}

func GetInstance() *ApolloConfig {
	return apolloConfig
}

func (c *ApolloConfig) Enable() bool {
	if c == nil || !c.EnableApollo {
		return false
	}
	c.RLock()
	defer c.RUnlock()
	return c.EnableApollo
}

func (c *ApolloConfig) GetBatchRequestsEnabled() bool {
	if c == nil || !c.EnableApollo {
		return false
	}
	c.RLock()
	defer c.RUnlock()

	return c.BatchRequestsEnabled
}

func (c *ApolloConfig) GetBatchRequestsLimit() uint {
	if c == nil || !c.EnableApollo {
		return 20
	}
	c.RLock()
	defer c.RUnlock()

	return c.BatchRequestsLimit
}

func (c *ApolloConfig) GetGasLimitFactor() float64 {
	if c == nil || !c.EnableApollo {
		return 1.0
	}
	c.RLock()
	defer c.RUnlock()

	return c.GasLimitFactor
}

func (c *ApolloConfig) GetDisableAPIs() []string {
	if c == nil || !c.EnableApollo {
		return nil
	}
	c.RLock()
	defer c.RUnlock()

	return c.DisableAPIs
}

func (c *ApolloConfig) setDisableAPIs(disableAPIs []string) {
	if c == nil || !c.EnableApollo {
		return
	}
	c.DisableAPIs = make([]string, len(disableAPIs))
	copy(c.DisableAPIs, disableAPIs)
}

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
		return len(GetInstance().DisableAPIs) > 0 && types.Contains(GetInstance().GetDisableAPIs(), rpc)
	}

	return len(e.cfg.DisableAPIs) > 0 && types.Contains(e.cfg.DisableAPIs, rpc)
}
