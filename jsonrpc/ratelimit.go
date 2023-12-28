package jsonrpc

import (
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/log"
	"golang.org/x/time/rate"
)

type RateLimit struct {
	rlm map[string]*rate.Limiter
	sync.RWMutex
}

var rateLimit = &RateLimit{}

// InitRateLimit initializes the rate limit config
func InitRateLimit(rlc RateLimitConfig) {
	rateLimit.Lock()
	defer rateLimit.Unlock()
	rateLimit.rlm = updateRateLimit(rlc)
}

func setRateLimit(rlc RateLimitConfig) {
	rateLimit.Lock()
	defer rateLimit.Unlock()
	rateLimit.rlm = updateRateLimit(rlc)
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
	rateLimit.RLock()
	rlm := rateLimit.rlm
	rateLimit.RUnlock()
	if rlm != nil && rlm[method] != nil && !rlm[method].Allow() {
		return false
	}
	return true
}
