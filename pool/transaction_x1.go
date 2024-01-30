package pool

import (
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

// IsClaimTx checks, if tx is a claim tx
func (tx *Transaction) IsClaimTx(l2BridgeAddr common.Address, freeClaimGasLimit uint64) bool {
	if tx.To() == nil {
		return false
	}

	txGas := tx.Gas()
	if txGas > freeClaimGasLimit {
		return false
	}

	if *tx.To() == l2BridgeAddr &&
		strings.HasPrefix("0x"+common.Bytes2Hex(tx.Data()), BridgeClaimMethodSignature) {
		return true
	}
	return false
}
