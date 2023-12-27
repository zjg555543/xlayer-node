package sequencer

import (
	"sync"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/config/types"
)

// ApolloConfig is the apollo RPC dynamic config
type ApolloConfig struct {
	EnableApollo           bool
	FullBatchSleepDuration types.Duration

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

// UpdateConfig updates the apollo config
func UpdateConfig(apolloConfig Config) {
	GetInstance().Lock()
	GetInstance().EnableApollo = true
	GetInstance().FullBatchSleepDuration = apolloConfig.Finalizer.FullBatchSleepDuration
	GetInstance().Unlock()
}

func getFullBatchSleepDuration(localDuration, timestampResolution time.Duration) time.Duration {
	var ret time.Duration
	if GetInstance().Enable() {
		GetInstance().RLock()
		defer GetInstance().RUnlock()
		ret = GetInstance().FullBatchSleepDuration.Duration
	} else {
		ret = localDuration
	}
	if ret > timestampResolution {
		ret = timestampResolution
	}

	return ret
}
