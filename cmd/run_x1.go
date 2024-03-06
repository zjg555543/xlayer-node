package main

import (
	"github.com/0xPolygonHermez/zkevm-node/config"
	"github.com/0xPolygonHermez/zkevm-node/pool"
)

func initRunForX1(c *config.Config, components []string) {
	pool.SetL2BridgeAddr(c.NetworkConfig.L2BridgeAddr)
}
