package datastreamer

import (
	"context"
	"math/big"

	"github.com/0xPolygonHermez/zkevm-node/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
)

// stateInterface gathers the methods required to interact with the state.
type stateInterface interface {
	GetLastL2BlockHeader(ctx context.Context, dbTx pgx.Tx) (*types.Header, error)
	GetDSGenesisBlock(ctx context.Context, dbTx pgx.Tx) (*state.DSL2Block, error)
	GetDSBatches(ctx context.Context, firstBatchNumber, lastBatchNumber uint64, readWIPBatch bool, dbTx pgx.Tx) ([]*state.DSBatch, error)
	GetDSL2Blocks(ctx context.Context, firstBatchNumber, lastBatchNumber uint64, dbTx pgx.Tx) ([]*state.DSL2Block, error)
	GetDSL2Transactions(ctx context.Context, firstL2Block, lastL2Block uint64, dbTx pgx.Tx) ([]*state.DSL2Transaction, error)
	GetStorageAt(ctx context.Context, address common.Address, position *big.Int, root common.Hash) (*big.Int, error)
}
