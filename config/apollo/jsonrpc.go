package apollo

import (
	"reflect"

	"github.com/0xPolygonHermez/zkevm-node/jsonrpc"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/apolloconfig/agollo/v4/storage"
)

func (c *Client) loadJsonRPC(value interface{}) {
	dstConf, err := c.unmarshal(value)
	if err != nil {
		log.Fatalf("failed to unmarshal json-rpc config: %v", err)
	}
	c.config.RPC = dstConf.RPC
	copy(c.config.RPC.DisableAPIs, dstConf.RPC.DisableAPIs)

	log.Infof("loaded json-rpc from apollo config: %+v", value.(string))
}

// fireJsonRPC fires the json-rpc config change
// BatchRequestsEnabled
// BatchRequestsLimit
// MaxLogsCount
// MaxLogsBlockRange
// MaxNativeBlockHashBlockRange
// GasLimitFactor
// DisableAPIs
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	newConf, err := c.unmarshal(value.NewValue)
	if err != nil {
		log.Errorf("failed to unmarshal l2gaspricer config: %v error: %v", value.NewValue, err)
		return
	}
	log.Infof("apollo l2gaspricer old config : %+v", c.config.L2GasPriceSuggester)
	log.Infof("apollo l2gaspricer config changed: %+v", value.NewValue.(string))
	c.Lock()
	defer c.Unlock()
	c.updateJsonRPC(&c.config.RPC, newConf.RPC)
}

// updateJsonRPC updates the json-rpc config
// BatchRequestsEnabled
// BatchRequestsLimit
// MaxLogsCount
// MaxLogsBlockRange
// MaxNativeBlockHashBlockRange
// GasLimitFactor
// DisableAPIs
func (c *Client) updateJsonRPC(dstConfig *jsonrpc.Config, srcConfig jsonrpc.Config) {
	if c == nil || !c.config.Apollo.Enable || dstConfig == nil {
		log.Infof("apollo is not enabled %v %v %v", c, dstConfig, srcConfig)
		return
	}
	if dstConfig.BatchRequestsEnabled != srcConfig.BatchRequestsEnabled {
		log.Infof("jsonrpc batch requests enabled changed from %v to %v",
			dstConfig.BatchRequestsEnabled, srcConfig.BatchRequestsEnabled)
		dstConfig.BatchRequestsEnabled = srcConfig.BatchRequestsEnabled
	}
	if dstConfig.BatchRequestsLimit != srcConfig.BatchRequestsLimit {
		log.Infof("jsonrpc batch requests limit changed from %v to %v",
			dstConfig.BatchRequestsLimit, srcConfig.BatchRequestsLimit)
		dstConfig.BatchRequestsLimit = srcConfig.BatchRequestsLimit
	}
	if dstConfig.MaxLogsCount != srcConfig.MaxLogsCount {
		log.Infof("jsonrpc max logs count changed from %v to %v",
			dstConfig.MaxLogsCount, srcConfig.MaxLogsCount)
		dstConfig.MaxLogsCount = srcConfig.MaxLogsCount
	}
	if dstConfig.MaxLogsBlockRange != srcConfig.MaxLogsBlockRange {
		log.Infof("jsonrpc max logs block range changed from %v to %v",
			dstConfig.MaxLogsBlockRange, srcConfig.MaxLogsBlockRange)
		dstConfig.MaxLogsBlockRange = srcConfig.MaxLogsBlockRange
	}
	if dstConfig.MaxNativeBlockHashBlockRange != srcConfig.MaxNativeBlockHashBlockRange {
		log.Infof("jsonrpc max native block hash block range changed from %v to %v",
			dstConfig.MaxNativeBlockHashBlockRange, srcConfig.MaxNativeBlockHashBlockRange)
		dstConfig.MaxNativeBlockHashBlockRange = srcConfig.MaxNativeBlockHashBlockRange
	}
	if dstConfig.GasLimitFactor != srcConfig.GasLimitFactor {
		log.Infof("jsonrpc gas limit factor changed from %v to %v",
			dstConfig.GasLimitFactor, srcConfig.GasLimitFactor)
		dstConfig.GasLimitFactor = srcConfig.GasLimitFactor
	}
	if !reflect.DeepEqual(dstConfig.DisableAPIs, srcConfig.DisableAPIs) {
		log.Infof("jsonrpc disable apis changed from %v to %v",
			dstConfig.DisableAPIs, srcConfig.DisableAPIs)
		copy(dstConfig.DisableAPIs, srcConfig.DisableAPIs)
	}
}

// FetchJsonRPCConfig fetches the json-rpc config, called from json-rpc module
func (c *Client) FetchJsonRPCConfig(config *jsonrpc.Config) {
	if c == nil || !c.config.Apollo.Enable || config == nil {
		return
	}

	c.RLock()
	c.RUnlock()

	c.updateJsonRPC(config, c.config.RPC)
}
