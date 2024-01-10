package etherman

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// GetZkEVMAddress returns the ZkEVM address
func (etherMan *Client) GetZkEVMAddress() (common.Address, error) {
	if etherMan == nil {
		return common.Address{}, fmt.Errorf("etherMan is nil")
	}
	return etherMan.l1Cfg.ZkEVMAddr, nil
}
