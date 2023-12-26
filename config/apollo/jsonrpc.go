package apollo

import (
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
// GasLimitFactor
// DisableAPIs
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	newConf, err := c.unmarshal(value.NewValue)
	if err != nil {
		log.Errorf("failed to unmarshal json-rpc config: %v error: %v", value.NewValue, err)
		return
	}
	log.Infof("apollo json-rpc old config : %+v", c.config.RPC)
	log.Infof("apollo json-rpc config changed: %+v", value.NewValue.(string))
	c.updateJsonRPC(newConf.RPC)
}

// updateJsonRPC updates the json-rpc config
// BatchRequestsEnabled
// BatchRequestsLimit
// GasLimitFactor
// DisableAPIs
func (c *Client) updateJsonRPC(srcConfig jsonrpc.Config) {
	if c == nil || !c.config.Apollo.Enable {
		log.Infof("apollo is not enabled %v %v", c, srcConfig)
		return
	}
	jsonrpc.UpdateConfig(srcConfig)
}
