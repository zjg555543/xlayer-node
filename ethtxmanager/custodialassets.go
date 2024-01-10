package ethtxmanager

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygonzkevm"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

type trace string

const (
	sigLen         = 4
	hashLen        = 32
	proofLen       = 24
	traceID  trace = "traceID"
)

func getTraceID(ctx context.Context) (string, string) {
	if ctx == nil {
		return "", ""
	}
	return string(traceID), ctx.Value(traceID).(string)
}

func (t trace) String() string {
	return string(t)
}

type sequenceBatchesArgs struct {
	Batches            []polygonzkevm.PolygonZkEVMBatchData `json:"batches"`
	L2Coinbase         common.Address                       `json:"l2Coinbase"`
	SignaturesAndAddrs []byte                               `json:"signaturesAndAddrs"`
}

type verifyBatchesTrustedAggregatorArgs struct {
	PendingStateNum  uint64                  `json:"pendingStateNum"`
	InitNumBatch     uint64                  `json:"initNumBatch"`
	FinalNewBatch    uint64                  `json:"finalNewBatch"`
	NewLocalExitRoot [hashLen]byte           `json:"newLocalExitRoot"`
	NewStateRoot     [hashLen]byte           `json:"newStateRoot"`
	Proof            [proofLen][hashLen]byte `json:"proof"`
}

var (
	errCustodialAssetsNotEnabled = errors.New("custodial assets not enabled")
	errEmptyTx                   = errors.New("empty tx")
	errLoadAbi                   = errors.New("failed to load contract ABI")
	errGetMethodID               = errors.New("failed to get method ID")
	errUnpack                    = errors.New("failed to unpack data")
)

func (c *Client) signTx(mTx monitoredTx, tx *types.Transaction) (*types.Transaction, error) {
	if c == nil || !c.cfg.CustodialAssets.Enable {
		return nil, errCustodialAssetsNotEnabled
	}
	sender := mTx.from
	ctx := context.WithValue(context.Background(), traceID, uuid.New().String())
	mLog := log.WithFields(getTraceID(ctx))
	mLog.Infof("begin sign tx %x", tx.Hash())

	var ret *types.Transaction

	contractAddress, err := c.etherman.GetZkEVMAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get zkEVM address: %v", err)
	}

	switch sender {
	case c.cfg.CustodialAssets.SequencerAddr:
		args, err := c.unpackSequenceBatchesTx(tx)
		if err != nil {
			mLog.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
		}
		infos, err := args.marshal(contractAddress)
		if err != nil {
			mLog.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
		}
		ret, err = c.postSignRequestAndWaitResult(ctx, mTx, c.newSignRequest(c.cfg.CustodialAssets.OperateTypeSeq, sender, infos))
		if err != nil {
			mLog.Errorf("failed to post custodial assets: %v", err)
			return nil, fmt.Errorf("failed to post custodial assets: %v", err)
		}
	case c.cfg.CustodialAssets.AggregatorAddr:
		args, err := c.unpackVerifyBatchesTrustedAggregatorTx(tx)
		if err != nil {
			mLog.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
		}
		infos, err := args.marshal(contractAddress)
		if err != nil {
			mLog.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
		}
		ret, err = c.postSignRequestAndWaitResult(ctx, mTx, c.newSignRequest(c.cfg.CustodialAssets.OperateTypeAgg, sender, infos))
		if err != nil {
			mLog.Errorf("failed to post custodial assets: %v", err)
			return nil, fmt.Errorf("failed to post custodial assets: %v", err)
		}
	default:
		mLog.Errorf("unknown sender %s", sender.String())
		return nil, fmt.Errorf("unknown sender %s", sender.String())
	}

	return ret, nil
}

func (c *Client) unpackSequenceBatchesTx(tx *types.Transaction) (*sequenceBatchesArgs, error) {
	if tx == nil || len(tx.Data()) < sigLen {
		return nil, errEmptyTx
	}
	retArgs, err := unpack(tx.Data())
	if err != nil {
		return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
	}
	retBytes, err := json.Marshal(retArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
	}
	var args sequenceBatchesArgs
	err = json.Unmarshal(retBytes, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tx %x data: %v", tx.Hash(), err)
	}

	return &args, nil
}

func (c *Client) unpackVerifyBatchesTrustedAggregatorTx(tx *types.Transaction) (*verifyBatchesTrustedAggregatorArgs, error) {
	if tx == nil || len(tx.Data()) < sigLen {
		return nil, errEmptyTx
	}
	retArgs, err := unpack(tx.Data())
	if err != nil {
		return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
	}
	retBytes, err := json.Marshal(retArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
	}
	var args verifyBatchesTrustedAggregatorArgs
	err = json.Unmarshal(retBytes, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tx %x data: %v", tx.Hash(), err)
	}

	return &args, nil
}

func unpack(data []byte) (map[string]interface{}, error) {
	// load contract ABI
	zkAbi, err := abi.JSON(strings.NewReader(polygonzkevm.PolygonzkevmABI))
	if err != nil {
		return nil, errLoadAbi
	}

	decodedSig := data[:sigLen]

	// recover Method from signature and ABI
	method, err := zkAbi.MethodById(decodedSig)
	if err != nil {
		return nil, errGetMethodID
	}

	decodedData := data[sigLen:]

	// unpack method inputs
	// result, err := method.Inputs.Unpack(decodedData)
	result := make(map[string]interface{})
	err = method.Inputs.UnpackIntoMap(result, decodedData)
	if err != nil {
		return nil, errUnpack
	}

	return result, nil
}

type batchData struct {
	Transactions       string `json:"transactions"`
	TransactionsHash   string `json:"transactionsHash"`
	GlobalExitRoot     string `json:"globalExitRoot"`
	Timestamp          uint64 `json:"timestamp"`
	MinForcedTimestamp uint64 `json:"minForcedTimestamp"`
}

func (s *sequenceBatchesArgs) marshal(contractAddress common.Address) (string, error) {
	if s == nil {
		return "", fmt.Errorf("sequenceBatchesArgs is nil")
	}
	httpArgs := struct {
		Batches            []batchData    `json:"batches"`
		L2Coinbase         common.Address `json:"l2Coinbase"`
		SignaturesAndAddrs string         `json:"signaturesAndAddrs"`
		ContractAddress    common.Address `json:"contractAddress"`
	}{
		L2Coinbase:         s.L2Coinbase,
		SignaturesAndAddrs: hex.EncodeToString(s.SignaturesAndAddrs),
		ContractAddress:    contractAddress,
	}

	httpArgs.Batches = make([]batchData, 0, len(s.Batches))
	for _, batch := range s.Batches {
		httpArgs.Batches = append(httpArgs.Batches, batchData{
			Transactions:       hex.EncodeToString(batch.Transactions),
			TransactionsHash:   hex.EncodeToString(batch.TransactionsHash[:]),
			GlobalExitRoot:     hex.EncodeToString(batch.GlobalExitRoot[:]),
			Timestamp:          batch.Timestamp,
			MinForcedTimestamp: batch.MinForcedTimestamp,
		})
	}
	ret, err := json.Marshal(httpArgs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sequenceBatchesArgs: %v", err)
	}

	return string(ret), nil
}

func (v *verifyBatchesTrustedAggregatorArgs) marshal(contractAddress common.Address) (string, error) {
	if v == nil {
		return "", fmt.Errorf("verifyBatchesTrustedAggregatorArgs is nil")
	}
	httpArgs := struct {
		PendingStateNum  uint64           `json:"pendingStateNum"`
		InitNumBatch     uint64           `json:"initNumBatch"`
		FinalNewBatch    uint64           `json:"finalNewBatch"`
		NewLocalExitRoot string           `json:"newLocalExitRoot"`
		NewStateRoot     string           `json:"newStateRoot"`
		Proof            [proofLen]string `json:"proof"`
		ContractAddress  common.Address   `json:"contractAddress"`
	}{
		PendingStateNum:  v.PendingStateNum,
		InitNumBatch:     v.InitNumBatch,
		FinalNewBatch:    v.FinalNewBatch,
		NewLocalExitRoot: hex.EncodeToString(v.NewLocalExitRoot[:]),
		NewStateRoot:     hex.EncodeToString(v.NewStateRoot[:]),
		ContractAddress:  contractAddress,
	}
	for i, v := range v.Proof {
		httpArgs.Proof[i] = hex.EncodeToString(v[:])
	}

	ret, err := json.Marshal(httpArgs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal verifyBatchesTrustedAggregatorArgs: %v", err)
	}

	return string(ret), nil
}

func (c *Client) halt(err error) {
	for {
		log.Errorf("fatal error: %s", err)
		log.Error("halting the eth tx manager")
		time.Sleep(5 * time.Second) //nolint:gomnd
	}
}
