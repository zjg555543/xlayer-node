package ethtxmanager

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygonzkevm"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

const (
	zkEVMAbi = `[{"inputs":[{"internalType":"contract IPolygonZkEVMGlobalExitRoot","name":"_globalExitRootManager","type":"address"},{"internalType":"contract IERC20Upgradeable","name":"_matic","type":"address"},{"internalType":"contract IVerifierRollup","name":"_rollupVerifier","type":"address"},{"internalType":"contract IPolygonZkEVMBridge","name":"_bridgeAddress","type":"address"},{"internalType":"contract IDataCommittee","name":"_dataCommitteeAddress","type":"address"},{"internalType":"uint64","name":"_chainID","type":"uint64"},{"internalType":"uint64","name":"_forkID","type":"uint64"},{"internalType":"uint256","name":"versionBeforeUpgrade","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"BatchAlreadyVerified","type":"error"},{"inputs":[],"name":"BatchNotSequencedOrNotSequenceEnd","type":"error"},{"inputs":[],"name":"ExceedMaxVerifyBatches","type":"error"},{"inputs":[],"name":"FinalNumBatchBelowLastVerifiedBatch","type":"error"},{"inputs":[],"name":"FinalNumBatchDoesNotMatchPendingState","type":"error"},{"inputs":[],"name":"FinalPendingStateNumInvalid","type":"error"},{"inputs":[],"name":"ForceBatchNotAllowed","type":"error"},{"inputs":[],"name":"ForceBatchTimeoutNotExpired","type":"error"},{"inputs":[],"name":"ForceBatchesAlreadyActive","type":"error"},{"inputs":[],"name":"ForceBatchesOverflow","type":"error"},{"inputs":[],"name":"ForcedDataDoesNotMatch","type":"error"},{"inputs":[],"name":"GlobalExitRootNotExist","type":"error"},{"inputs":[],"name":"HaltTimeoutNotExpired","type":"error"},{"inputs":[],"name":"InitBatchMustMatchCurrentForkID","type":"error"},{"inputs":[],"name":"InitNumBatchAboveLastVerifiedBatch","type":"error"},{"inputs":[],"name":"InitNumBatchDoesNotMatchPendingState","type":"error"},{"inputs":[],"name":"InvalidProof","type":"error"},{"inputs":[],"name":"InvalidRangeBatchTimeTarget","type":"error"},{"inputs":[],"name":"InvalidRangeForceBatchTimeout","type":"error"},{"inputs":[],"name":"InvalidRangeMultiplierBatchFee","type":"error"},{"inputs":[],"name":"NewAccInputHashDoesNotExist","type":"error"},{"inputs":[],"name":"NewPendingStateTimeoutMustBeLower","type":"error"},{"inputs":[],"name":"NewStateRootNotInsidePrime","type":"error"},{"inputs":[],"name":"NewTrustedAggregatorTimeoutMustBeLower","type":"error"},{"inputs":[],"name":"NotEnoughMaticAmount","type":"error"},{"inputs":[],"name":"OldAccInputHashDoesNotExist","type":"error"},{"inputs":[],"name":"OldStateRootDoesNotExist","type":"error"},{"inputs":[],"name":"OnlyAdmin","type":"error"},{"inputs":[],"name":"OnlyEmergencyState","type":"error"},{"inputs":[],"name":"OnlyNotEmergencyState","type":"error"},{"inputs":[],"name":"OnlyPendingAdmin","type":"error"},{"inputs":[],"name":"OnlyTrustedAggregator","type":"error"},{"inputs":[],"name":"OnlyTrustedSequencer","type":"error"},{"inputs":[],"name":"PendingStateDoesNotExist","type":"error"},{"inputs":[],"name":"PendingStateInvalid","type":"error"},{"inputs":[],"name":"PendingStateNotConsolidable","type":"error"},{"inputs":[],"name":"PendingStateTimeoutExceedHaltAggregationTimeout","type":"error"},{"inputs":[],"name":"SequenceZeroBatches","type":"error"},{"inputs":[],"name":"SequencedTimestampBelowForcedTimestamp","type":"error"},{"inputs":[],"name":"SequencedTimestampInvalid","type":"error"},{"inputs":[],"name":"StoredRootMustBeDifferentThanNewRoot","type":"error"},{"inputs":[],"name":"TransactionsLengthAboveMax","type":"error"},{"inputs":[],"name":"TrustedAggregatorTimeoutExceedHaltAggregationTimeout","type":"error"},{"inputs":[],"name":"TrustedAggregatorTimeoutNotExpired","type":"error"},{"inputs":[],"name":"VersionAlreadyUpdated","type":"error"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"newAdmin","type":"address"}],"name":"AcceptAdminRole","type":"event"},{"anonymous":false,"inputs":[],"name":"ActivateForceBatches","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"},{"indexed":false,"internalType":"bytes32","name":"stateRoot","type":"bytes32"},{"indexed":true,"internalType":"uint64","name":"pendingStateNum","type":"uint64"}],"name":"ConsolidatePendingState","type":"event"},{"anonymous":false,"inputs":[],"name":"EmergencyStateActivated","type":"event"},{"anonymous":false,"inputs":[],"name":"EmergencyStateDeactivated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"forceBatchNum","type":"uint64"},{"indexed":false,"internalType":"bytes32","name":"lastGlobalExitRoot","type":"bytes32"},{"indexed":false,"internalType":"address","name":"sequencer","type":"address"},{"indexed":false,"internalType":"bytes","name":"transactions","type":"bytes"}],"name":"ForceBatch","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint8","name":"version","type":"uint8"}],"name":"Initialized","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"},{"indexed":false,"internalType":"bytes32","name":"stateRoot","type":"bytes32"},{"indexed":true,"internalType":"address","name":"aggregator","type":"address"}],"name":"OverridePendingState","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"storedStateRoot","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"provedStateRoot","type":"bytes32"}],"name":"ProveNonDeterministicPendingState","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"}],"name":"SequenceBatches","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"}],"name":"SequenceForceBatches","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"newforceBatchTimeout","type":"uint64"}],"name":"SetForceBatchTimeout","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint16","name":"newMultiplierBatchFee","type":"uint16"}],"name":"SetMultiplierBatchFee","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"newPendingStateTimeout","type":"uint64"}],"name":"SetPendingStateTimeout","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"newTrustedAggregator","type":"address"}],"name":"SetTrustedAggregator","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"newTrustedAggregatorTimeout","type":"uint64"}],"name":"SetTrustedAggregatorTimeout","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"newTrustedSequencer","type":"address"}],"name":"SetTrustedSequencer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"newTrustedSequencerURL","type":"string"}],"name":"SetTrustedSequencerURL","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"newVerifyBatchTimeTarget","type":"uint64"}],"name":"SetVerifyBatchTimeTarget","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"newPendingAdmin","type":"address"}],"name":"TransferAdminRole","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"numBatch","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"forkID","type":"uint64"},{"indexed":false,"internalType":"string","name":"version","type":"string"}],"name":"UpdateZkEVMVersion","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"},{"indexed":false,"internalType":"bytes32","name":"stateRoot","type":"bytes32"},{"indexed":true,"internalType":"address","name":"aggregator","type":"address"}],"name":"VerifyBatches","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint64","name":"numBatch","type":"uint64"},{"indexed":false,"internalType":"bytes32","name":"stateRoot","type":"bytes32"},{"indexed":true,"internalType":"address","name":"aggregator","type":"address"}],"name":"VerifyBatchesTrustedAggregator","type":"event"},{"inputs":[],"name":"VERSION_BEFORE_UPGRADE","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"acceptAdminRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"sequencedBatchNum","type":"uint64"}],"name":"activateEmergencyState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"activateForceBatches","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"admin","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"batchFee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"","type":"uint64"}],"name":"batchNumToStateRoot","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"bridgeAddress","outputs":[{"internalType":"contract IPolygonZkEVMBridge","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"calculateRewardPerBatch","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"chainID","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"newStateRoot","type":"uint256"}],"name":"checkStateRootInsidePrime","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint64","name":"pendingStateNum","type":"uint64"}],"name":"consolidatePendingState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"dataCommitteeAddress","outputs":[{"internalType":"contract IDataCommittee","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"deactivateEmergencyState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes","name":"transactions","type":"bytes"},{"internalType":"uint256","name":"maticAmount","type":"uint256"}],"name":"forceBatch","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"forceBatchTimeout","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"","type":"uint64"}],"name":"forcedBatches","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"forkID","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getForcedBatchFee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"initNumBatch","type":"uint64"},{"internalType":"uint64","name":"finalNewBatch","type":"uint64"},{"internalType":"bytes32","name":"newLocalExitRoot","type":"bytes32"},{"internalType":"bytes32","name":"oldStateRoot","type":"bytes32"},{"internalType":"bytes32","name":"newStateRoot","type":"bytes32"}],"name":"getInputSnarkBytes","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getLastVerifiedBatch","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"globalExitRootManager","outputs":[{"internalType":"contract IPolygonZkEVMGlobalExitRoot","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"admin","type":"address"},{"internalType":"address","name":"trustedSequencer","type":"address"},{"internalType":"uint64","name":"pendingStateTimeout","type":"uint64"},{"internalType":"address","name":"trustedAggregator","type":"address"},{"internalType":"uint64","name":"trustedAggregatorTimeout","type":"uint64"}],"internalType":"struct PolygonZkEVM.InitializePackedParameters","name":"initializePackedParameters","type":"tuple"},{"internalType":"bytes32","name":"genesisRoot","type":"bytes32"},{"internalType":"string","name":"_trustedSequencerURL","type":"string"},{"internalType":"string","name":"_networkName","type":"string"},{"internalType":"string","name":"_version","type":"string"}],"name":"initialize","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"isEmergencyState","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"isForcedBatchDisallowed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"pendingStateNum","type":"uint64"}],"name":"isPendingStateConsolidable","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastBatchSequenced","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastForceBatch","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastForceBatchSequenced","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastPendingState","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastPendingStateConsolidated","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastTimestamp","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastVerifiedBatch","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lastVerifiedBatchBeforeUpgrade","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"matic","outputs":[{"internalType":"contract IERC20Upgradeable","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"multiplierBatchFee","outputs":[{"internalType":"uint16","name":"","type":"uint16"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"networkName","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"initPendingStateNum","type":"uint64"},{"internalType":"uint64","name":"finalPendingStateNum","type":"uint64"},{"internalType":"uint64","name":"initNumBatch","type":"uint64"},{"internalType":"uint64","name":"finalNewBatch","type":"uint64"},{"internalType":"bytes32","name":"newLocalExitRoot","type":"bytes32"},{"internalType":"bytes32","name":"newStateRoot","type":"bytes32"},{"internalType":"bytes32[24]","name":"proof","type":"bytes32[24]"}],"name":"overridePendingState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pendingAdmin","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pendingStateTimeout","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"pendingStateTransitions","outputs":[{"internalType":"uint64","name":"timestamp","type":"uint64"},{"internalType":"uint64","name":"lastVerifiedBatch","type":"uint64"},{"internalType":"bytes32","name":"exitRoot","type":"bytes32"},{"internalType":"bytes32","name":"stateRoot","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"initPendingStateNum","type":"uint64"},{"internalType":"uint64","name":"finalPendingStateNum","type":"uint64"},{"internalType":"uint64","name":"initNumBatch","type":"uint64"},{"internalType":"uint64","name":"finalNewBatch","type":"uint64"},{"internalType":"bytes32","name":"newLocalExitRoot","type":"bytes32"},{"internalType":"bytes32","name":"newStateRoot","type":"bytes32"},{"internalType":"bytes32[24]","name":"proof","type":"bytes32[24]"}],"name":"proveNonDeterministicPendingState","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"rollupVerifier","outputs":[{"internalType":"contract IVerifierRollup","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"internalType":"bytes","name":"transactions","type":"bytes"},{"internalType":"bytes32","name":"transactionsHash","type":"bytes32"},{"internalType":"bytes32","name":"globalExitRoot","type":"bytes32"},{"internalType":"uint64","name":"timestamp","type":"uint64"},{"internalType":"uint64","name":"minForcedTimestamp","type":"uint64"}],"internalType":"struct PolygonZkEVM.BatchData[]","name":"batches","type":"tuple[]"},{"internalType":"address","name":"l2Coinbase","type":"address"},{"internalType":"bytes","name":"signaturesAndAddrs","type":"bytes"}],"name":"sequenceBatches","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"bytes","name":"transactions","type":"bytes"},{"internalType":"bytes32","name":"globalExitRoot","type":"bytes32"},{"internalType":"uint64","name":"minForcedTimestamp","type":"uint64"}],"internalType":"struct PolygonZkEVM.ForcedBatchData[]","name":"batches","type":"tuple[]"}],"name":"sequenceForceBatches","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"","type":"uint64"}],"name":"sequencedBatches","outputs":[{"internalType":"bytes32","name":"accInputHash","type":"bytes32"},{"internalType":"uint64","name":"sequencedTimestamp","type":"uint64"},{"internalType":"uint64","name":"previousLastBatchSequenced","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint16","name":"newBatchFee","type":"uint16"}],"name":"setBatchFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"newforceBatchTimeout","type":"uint64"}],"name":"setForceBatchTimeout","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint16","name":"newMultiplierBatchFee","type":"uint16"}],"name":"setMultiplierBatchFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"newPendingStateTimeout","type":"uint64"}],"name":"setPendingStateTimeout","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newTrustedAggregator","type":"address"}],"name":"setTrustedAggregator","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"newTrustedAggregatorTimeout","type":"uint64"}],"name":"setTrustedAggregatorTimeout","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newTrustedSequencer","type":"address"}],"name":"setTrustedSequencer","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"newTrustedSequencerURL","type":"string"}],"name":"setTrustedSequencerURL","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"newVerifyBatchTimeTarget","type":"uint64"}],"name":"setVerifyBatchTimeTarget","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newPendingAdmin","type":"address"}],"name":"transferAdminRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"trustedAggregator","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"trustedAggregatorTimeout","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"trustedSequencer","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"trustedSequencerURL","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"_versionString","type":"string"}],"name":"updateVersion","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"verifyBatchTimeTarget","outputs":[{"internalType":"uint64","name":"","type":"uint64"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint64","name":"pendingStateNum","type":"uint64"},{"internalType":"uint64","name":"initNumBatch","type":"uint64"},{"internalType":"uint64","name":"finalNewBatch","type":"uint64"},{"internalType":"bytes32","name":"newLocalExitRoot","type":"bytes32"},{"internalType":"bytes32","name":"newStateRoot","type":"bytes32"},{"internalType":"bytes32[24]","name":"proof","type":"bytes32[24]"}],"name":"verifyBatches","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint64","name":"pendingStateNum","type":"uint64"},{"internalType":"uint64","name":"initNumBatch","type":"uint64"},{"internalType":"uint64","name":"finalNewBatch","type":"uint64"},{"internalType":"bytes32","name":"newLocalExitRoot","type":"bytes32"},{"internalType":"bytes32","name":"newStateRoot","type":"bytes32"},{"internalType":"bytes32[24]","name":"proof","type":"bytes32[24]"}],"name":"verifyBatchesTrustedAggregator","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"version","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`
	sigLen   = 4
	hashLen  = 32
	proofLen = 24
	traceID  = "traceID"
)

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

func (c *Client) signTx(sender common.Address, tx *types.Transaction) (*types.Transaction, error) {
	if c == nil || !c.cfg.CustodialAssetsConfig.Enable {
		return nil, errCustodialAssetsNotEnabled
	}
	ctx := context.WithValue(context.Background(), traceID, uuid.New().String())
	mLog := log.WithFields(traceID, ctx.Value(traceID))
	mLog.Infof("begin sign tx %x", tx.Hash())

	switch sender {
	case c.cfg.CustodialAssetsConfig.SequencerAddr:
		args, err := c.unpackSequenceBatchesTx(tx)
		if err != nil {
			mLog.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
		}
		infos, err := args.marshal()
		if err != nil {
			mLog.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
		}
		_, err = c.postSignRequestAndWaitResult(ctx, c.newSignRequest(operateTypeSeq, sender, string(infos)))
		if err != nil {
			mLog.Errorf("failed to post custodial assets: %v", err)
			return nil, fmt.Errorf("failed to post custodial assets: %v", err)
		}
	case c.cfg.CustodialAssetsConfig.AggregatorAddr:
		args, err := c.unpackVerifyBatchesTrustedAggregatorTx(tx)
		if err != nil {
			mLog.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to unpack tx %x data: %v", tx.Hash(), err)
		}
		infos, err := args.marshal()
		if err != nil {
			mLog.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
			return nil, fmt.Errorf("failed to marshal tx %x data: %v", tx.Hash(), err)
		}
		_, err = c.postSignRequestAndWaitResult(ctx, c.newSignRequest(operateTypeAgg, sender, string(infos)))
		if err != nil {
			mLog.Errorf("failed to post custodial assets: %v", err)
			return nil, fmt.Errorf("failed to post custodial assets: %v", err)
		}
	default:
		mLog.Errorf("unknown sender %s", sender.String())
		return nil, fmt.Errorf("unknown sender %s", sender.String())
	}

	return nil, nil
}

func (c *Client) unpackSequenceBatchesTx(tx *types.Transaction) (*sequenceBatchesArgs, error) {
	if tx == nil || len(tx.Data()) == 0 {
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
	if tx == nil || len(tx.Data()) == 0 {
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
	zkAbi, err := abi.JSON(strings.NewReader(zkEVMAbi))
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
	TransactionHash    string `json:"transactionHash"`
	GlobalExitRoot     string `json:"globalExitRoot"`
	Timestamp          uint64 `json:"timestamp"`
	MinForcedTimestamp uint64 `json:"minForcedTimestamp"`
}

func (s *sequenceBatchesArgs) marshal() (string, error) {
	if s == nil {
		return "", fmt.Errorf("sequenceBatchesArgs is nil")
	}
	httpArgs := struct {
		Batches            []batchData    `json:"batches"`
		L2Coinbase         common.Address `json:"l2Coinbase"`
		SignaturesAndAddrs string         `json:"signaturesAndAddrs"`
	}{
		L2Coinbase:         s.L2Coinbase,
		SignaturesAndAddrs: hex.EncodeToString(s.SignaturesAndAddrs),
	}

	httpArgs.Batches = make([]batchData, 0, len(s.Batches))
	for _, batch := range s.Batches {
		httpArgs.Batches = append(httpArgs.Batches, batchData{
			Transactions:       hex.EncodeToString(batch.Transactions),
			TransactionHash:    hex.EncodeToString(batch.TransactionsHash[:]),
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

func (v *verifyBatchesTrustedAggregatorArgs) marshal() (string, error) {
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
	}{
		PendingStateNum:  v.PendingStateNum,
		InitNumBatch:     v.InitNumBatch,
		FinalNewBatch:    v.FinalNewBatch,
		NewLocalExitRoot: hex.EncodeToString(v.NewLocalExitRoot[:]),
		NewStateRoot:     hex.EncodeToString(v.NewStateRoot[:]),
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
