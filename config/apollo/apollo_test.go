package apollo

import (
	"testing"
	"time"

	nodeConfig "github.com/0xPolygonHermez/zkevm-node/config"
	"github.com/0xPolygonHermez/zkevm-node/config/types"
)

func TestApolloClient_LoadConfig(t *testing.T) {
	nc := &nodeConfig.Config{
		Apollo: types.ApolloConfig{
			IP:            "http://52.40.214.137:26657",
			AppID:         "x1-devnet",
			NamespaceName: "jsonrpc-ro.txt,jsonrpc-ro.properties",
			Enable:        true,
		},
	}
	client := NewClient(nc)

	client.LoadConfig()
	t.Log(nc.RPC)
	time.Sleep(2 * time.Minute)
	t.Log(nc.RPC)
}
