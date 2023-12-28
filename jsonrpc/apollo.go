package jsonrpc

import (
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/jsonrpc/types"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"golang.org/x/time/rate"
)

// ApolloConfig is the apollo RPC dynamic config
type ApolloConfig struct {
	EnableApollo         bool            `json:"enable"`
	BatchRequestsEnabled bool            `json:"batchRequestsEnabled"`
	BatchRequestsLimit   uint            `json:"batchRequestsLimit"`
	GasLimitFactor       float64         `json:"gasLimitFactor"`
	DisableAPIs          []string        `json:"disableAPIs"`
	RateLimit            RateLimitConfig `json:"rateLimit"`

	rateLimit map[string]*rate.Limiter
	sync.RWMutex
}

var apolloConfig = &ApolloConfig{}

// getApolloConfig returns the singleton instance
func getApolloConfig() *ApolloConfig {
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

func (c *ApolloConfig) setRateLimit(rateLimit RateLimitConfig) {
	if c == nil || !c.EnableApollo {
		return
	}
	c.RateLimit = rateLimit
	c.RateLimit.RateLimitApis = make([]string, len(rateLimit.RateLimitApis))
	copy(c.RateLimit.RateLimitApis, rateLimit.RateLimitApis)
	c.RateLimit.SpecialApis = make([]RateLimitItem, len(rateLimit.SpecialApis))
	copy(c.RateLimit.SpecialApis, rateLimit.SpecialApis)
	c.rateLimit = updateRateLimit(c.RateLimit)
}

// InitRateLimit initializes the rate limit config
func InitRateLimit(rateLimit RateLimitConfig) {
	getApolloConfig().Lock()
	defer getApolloConfig().Unlock()
	getApolloConfig().rateLimit = updateRateLimit(rateLimit)
}

// UpdateConfig updates the apollo config
func UpdateConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	getApolloConfig().BatchRequestsEnabled = apolloConfig.BatchRequestsEnabled
	getApolloConfig().BatchRequestsLimit = apolloConfig.BatchRequestsLimit
	getApolloConfig().GasLimitFactor = apolloConfig.GasLimitFactor
	getApolloConfig().setDisableAPIs(apolloConfig.DisableAPIs)
	getApolloConfig().setRateLimit(apolloConfig.RateLimit)
	getApolloConfig().Unlock()
}

func (e *EthEndpoints) isDisabled(rpc string) bool {
	if getApolloConfig().Enable() {
		getApolloConfig().RLock()
		defer getApolloConfig().RUnlock()
		return len(getApolloConfig().DisableAPIs) > 0 && types.Contains(getApolloConfig().DisableAPIs, rpc)
	}

	return len(e.cfg.DisableAPIs) > 0 && types.Contains(e.cfg.DisableAPIs, rpc)
}

func updateRateLimit(rateLimit RateLimitConfig) map[string]*rate.Limiter {
	log.Infof("rate limit config updated, config: %+v", rateLimit)
	if rateLimit.Enabled {
		log.Infof("rate limit enabled, api: %v, count: %d, duration: %d", rateLimit.RateLimitApis, rateLimit.RateLimitCount, rateLimit.RateLimitDuration)
		rlm := make(map[string]*rate.Limiter)
		for _, api := range rateLimit.RateLimitApis {
			rlm[api] = rate.NewLimiter(rate.Limit(rateLimit.RateLimitCount), rateLimit.RateLimitDuration)
		}
		for _, api := range rateLimit.SpecialApis {
			log.Infof("special api rate limit enabled, api: %v, count: %d, duration: %d", api.Api, api.Count, api.Duration)
			rlm[api.Api] = rate.NewLimiter(rate.Limit(api.Count), api.Duration)
		}
		return rlm
	}
	return nil
}

func methodRateLimitAllow(method string) bool {
	getApolloConfig().RLock()
	rlm := getApolloConfig().rateLimit
	getApolloConfig().RUnlock()
	if rlm != nil && rlm[method] != nil && !rlm[method].Allow() {
		return false
	}
	return true
}
