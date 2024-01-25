package datastreamer

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	dslog "github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/0xPolygonHermez/zkevm-node/event"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/0xPolygonHermez/zkevm-node/state"
	"github.com/ethereum/go-ethereum/common"
)

// DataStreamer represents a data streamer
type DataStreamer struct {
	cfg          Config
	dsLog        dslog.Config
	state        stateInterface
	eventLog     *event.EventLog
	streamServer *datastreamer.StreamServer
}

// New inits data streamer
func New(cfg Config, logCfg log.Config, state stateInterface, eventLog *event.EventLog) (*DataStreamer, error) {
	return &DataStreamer{
		cfg:      cfg,
		dsLog:    dslog.Config{Environment: dslog.LogEnvironment(logCfg.Environment), Level: logCfg.Level, Outputs: logCfg.Outputs},
		state:    state,
		eventLog: eventLog,
	}, nil
}

// Start starts the data streamer
func (s *DataStreamer) Start(ctx context.Context) {
	var err error
	s.streamServer, err = datastreamer.NewServer(s.cfg.Port, state.StreamTypeSequencer, s.cfg.Filename, &s.dsLog)
	if err != nil {
		log.Fatalf("failed to create stream server, err: %v", err)
	}

	err = s.streamServer.Start()
	if err != nil {
		log.Fatalf("failed to start stream server, err: %v", err)
	}

	err = state.GenerateDataStreamerFile(ctx, s.streamServer, s.state, false, nil)
	if err != nil {
		log.Fatalf("failed to generate data streamer file, err: %v", err)
	}

	s.loopDataStreamer(ctx)
}

func (s *DataStreamer) loopDataStreamer(ctx context.Context) {
	log.Infof("Starting data streamer loop")
	for {
		select {
		case <-ctx.Done():
			log.Infof("Loop data streamer is exiting")
			return
		default:
			time.Sleep(s.cfg.WaitInterval.Duration)

			latestBatchNum, latestBlockNum, err := s.getLatestBatchAndBlock(s.streamServer)
			if err != nil {
				log.Fatalf("Error getting current batch and block: %s", err.Error())
				break
			}

			err = s.handleDataStreamer(ctx, s.streamServer, s.state, latestBatchNum, latestBlockNum)
			if err != nil {
				log.Fatalf("Error sending data to streamer: %s", err.Error())
				break
			}
		}
	}
}

// handleDataStreamer generates or resumes a data stream file
func (s *DataStreamer) handleDataStreamer(ctx context.Context, streamServer *datastreamer.StreamServer, stateDB state.DSState, latestBatchNum uint64, latestBlockNum uint64) error {
	batches, err := stateDB.GetDSBatches(ctx, latestBatchNum, latestBatchNum+s.cfg.MaxBatchLimit, true, nil)
	if err != nil {
		if err == state.ErrStateNotSynchronized {
			log.Errorf("State not synchronized, retrying in %s", s.cfg.WaitInterval.Duration.String())
			return nil
		}
		log.Errorf("Error getting batch %d: %s", latestBatchNum, err.Error())
		return err
	}

	if len(batches) == 0 {
		return nil
	}

	var currentGER = batches[0].GlobalExitRoot
	log.Infof("Processing batches [%d, %d]", batches[0].BatchNumber, batches[len(batches)-1].BatchNumber)
	l2BlocksTemp, err := stateDB.GetDSL2Blocks(ctx, batches[0].BatchNumber, batches[len(batches)-1].BatchNumber, nil)
	if err != nil {
		log.Errorf("Error getting L2 blocks for batches starting at %d: %s", batches[0].BatchNumber, err.Error())
		return err
	}

	l2Blocks := make([]*state.DSL2Block, 0)
	if latestBlockNum > 0 {
		for _, l2block := range l2BlocksTemp {
			if l2block.L2BlockNumber <= latestBlockNum {
				continue
			}
			l2Blocks = append(l2Blocks, l2block)
		}
	} else {
		l2Blocks = l2BlocksTemp
	}

	l2Txs := make([]*state.DSL2Transaction, 0)
	if len(l2Blocks) > 0 {
		log.Infof("Processing old blocks [%d,%d], new blocks [%d,%d]", l2BlocksTemp[0].L2BlockNumber, l2BlocksTemp[len(l2BlocksTemp)-1].L2BlockNumber, l2Blocks[0].L2BlockNumber, l2Blocks[len(l2Blocks)-1].L2BlockNumber)
		l2Txs, err = stateDB.GetDSL2Transactions(ctx, l2Blocks[0].L2BlockNumber, l2Blocks[len(l2Blocks)-1].L2BlockNumber, nil)
		if err != nil {
			log.Errorf("Error getting L2 transactions for blocks starting at %d: %s", l2Blocks[0].L2BlockNumber, err.Error())
			return err
		}
	}

	fullBatches := state.ComputeFullBatches(batches, l2Blocks, l2Txs)
	for _, fullBatch := range fullBatches {
		if len(fullBatch.L2Blocks) == 0 {
			ger, err := s.handleGER(streamServer, fullBatch, currentGER)
			if err != nil {
				return err
			}
			if ger != (common.Hash{}) {
				currentGER = ger
			}
			continue
		}

		currentGER, err = s.handleBlocks(ctx, streamServer, stateDB, fullBatch)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *DataStreamer) getLatestBatchAndBlock(streamServer *datastreamer.StreamServer) (uint64, uint64, error) {
	header := streamServer.GetHeader()
	if header.TotalEntries == 0 {
		return 0, 0, errors.New("no entries in data streamer file")
	}

	var currentBatchNumber uint64
	var currentL2Block uint64
	latestEntry, err := streamServer.GetEntry(header.TotalEntries - 1)
	if err != nil {
		return 0, 0, err
	}

	switch latestEntry.Type {
	case state.EntryTypeUpdateGER:
		currentBatchNumber = binary.LittleEndian.Uint64(latestEntry.Data[0:8])
	case state.EntryTypeL2BlockEnd:
		currentL2Block = binary.LittleEndian.Uint64(latestEntry.Data[0:8])

		bookMark := state.DSBookMark{
			Type:          state.BookMarkTypeL2Block,
			L2BlockNumber: currentL2Block,
		}

		firstEntry, err := streamServer.GetFirstEventAfterBookmark(bookMark.Encode())
		if err != nil {
			return 0, 0, err
		}
		currentBatchNumber = binary.LittleEndian.Uint64(firstEntry.Data[0:8])
	default:
		return 0, 0, errors.New("latest entry type is not UpdateGER or L2BlockEnd")
	}

	log.Infof("Get latest batch number: %v, L2 block number: %v, entry type:%v", currentBatchNumber, currentL2Block, latestEntry.Type)
	return currentBatchNumber, currentL2Block, nil
}

func (s *DataStreamer) handleBlocks(ctx context.Context, streamServer *datastreamer.StreamServer, stateDB state.DSState, batch *state.DSFullBatch) (common.Hash, error) {
	err := streamServer.StartAtomicOp()
	if err != nil {
		return common.Hash{}, err
	}
	var currentGER common.Hash

	for _, l2block := range batch.L2Blocks {
		blockStart := state.DSL2BlockStart{
			BatchNumber:    l2block.BatchNumber,
			L2BlockNumber:  l2block.L2BlockNumber,
			Timestamp:      l2block.Timestamp,
			GlobalExitRoot: l2block.GlobalExitRoot,
			Coinbase:       l2block.Coinbase,
			ForkID:         l2block.ForkID,
		}

		bookMark := state.DSBookMark{
			Type:          state.BookMarkTypeL2Block,
			L2BlockNumber: blockStart.L2BlockNumber,
		}

		_, err = streamServer.AddStreamBookmark(bookMark.Encode())
		if err != nil {
			return common.Hash{}, err
		}

		_, err = streamServer.AddStreamEntry(state.EntryTypeL2BlockStart, blockStart.Encode())
		if err != nil {
			return common.Hash{}, err
		}

		for _, tx := range l2block.Txs {
			// Populate intermediate state root
			position := state.GetSystemSCPosition(l2block.L2BlockNumber)
			imStateRoot, err := stateDB.GetStorageAt(ctx, common.HexToAddress(state.SystemSC), big.NewInt(0).SetBytes(position), l2block.StateRoot)
			if err != nil {
				return common.Hash{}, err
			}
			tx.StateRoot = common.BigToHash(imStateRoot)
			log.Infof("Processing add stream entry, block:%v", tx.L2BlockNumber)
			_, err = streamServer.AddStreamEntry(state.EntryTypeL2Tx, tx.Encode())
			if err != nil {
				return common.Hash{}, err
			}
		}

		blockEnd := state.DSL2BlockEnd{
			L2BlockNumber: l2block.L2BlockNumber,
			BlockHash:     l2block.BlockHash,
			StateRoot:     l2block.StateRoot,
		}

		_, err = streamServer.AddStreamEntry(state.EntryTypeL2BlockEnd, blockEnd.Encode())
		if err != nil {
			return common.Hash{}, err
		}
		currentGER = l2block.GlobalExitRoot
	}
	// Commit at the end of each batch group
	err = streamServer.CommitAtomicOp()
	if err != nil {
		return common.Hash{}, err
	}

	return currentGER, nil
}

func (s *DataStreamer) handleGER(streamServer *datastreamer.StreamServer, batch *state.DSFullBatch, ger common.Hash) (common.Hash, error) {
	if batch.StateRoot == (common.Hash{}) {
		return common.Hash{}, nil
	}
	// Check if there is a GER update
	if batch.GlobalExitRoot != ger && batch.GlobalExitRoot != (common.Hash{}) {
		updateGer := state.DSUpdateGER{
			BatchNumber:    batch.BatchNumber,
			Timestamp:      batch.Timestamp.Unix(),
			GlobalExitRoot: batch.GlobalExitRoot,
			Coinbase:       batch.Coinbase,
			ForkID:         batch.ForkID,
			StateRoot:      batch.StateRoot,
		}

		err := streamServer.StartAtomicOp()
		if err != nil {
			return common.Hash{}, err
		}

		log.Infof("Processing add stream entry, GER batch %d", batch.BatchNumber)
		_, err = streamServer.AddStreamEntry(state.EntryTypeUpdateGER, updateGer.Encode())
		if err != nil {
			return common.Hash{}, err
		}

		err = streamServer.CommitAtomicOp()
		if err != nil {
			return common.Hash{}, err
		}

		ger = batch.GlobalExitRoot
		return ger, nil
	}
	return common.Hash{}, nil
}
