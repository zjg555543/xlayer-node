package datastreamer

import (
	"context"
	"encoding/binary"
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

	s.updateDataStreamerFile(ctx, s.streamServer)

	s.loopSendDataStreamer(ctx)
}

func (s *DataStreamer) updateDataStreamerFile(ctx context.Context, streamServer *datastreamer.StreamServer) {
	err := state.GenerateDataStreamerFile(ctx, streamServer, s.state, true, nil)
	if err != nil {
		log.Fatalf("failed to generate data streamer file, err: %v", err)
	}
	log.Info("Data streamer file updated")
}

func (s *DataStreamer) loopSendDataStreamer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("context done, exiting")
			return
		default:
			time.Sleep(s.cfg.WaitPeriodReadDB.Duration)
			err := s.trySendDataStreamer(ctx, s.streamServer, s.state, true, nil)
			if err != nil {
				log.Fatalf("Error sending data to streamer: %s", err.Error())
				break
			}
		}
	}
}

// GenerateDataStreamerFile generates or resumes a data stream file
func (s *DataStreamer) trySendDataStreamer(ctx context.Context, streamServer *datastreamer.StreamServer, stateDB state.DSState, readWIPBatch bool, imStateRoots *map[uint64][]byte) error {
	header := streamServer.GetHeader()

	var currentBatchNumber uint64 = 0
	var currentL2Block uint64 = 0

	if header.TotalEntries == 0 {
		// Get Genesis block
		genesisL2Block, err := stateDB.GetDSGenesisBlock(ctx, nil)
		if err != nil {
			return err
		}

		err = streamServer.StartAtomicOp()
		if err != nil {
			return err
		}

		bookMark := state.DSBookMark{
			Type:          state.BookMarkTypeL2Block,
			L2BlockNumber: genesisL2Block.L2BlockNumber,
		}

		_, err = streamServer.AddStreamBookmark(bookMark.Encode())
		if err != nil {
			return err
		}

		genesisBlock := state.DSL2BlockStart{
			BatchNumber:    genesisL2Block.BatchNumber,
			L2BlockNumber:  genesisL2Block.L2BlockNumber,
			Timestamp:      genesisL2Block.Timestamp,
			GlobalExitRoot: genesisL2Block.GlobalExitRoot,
			Coinbase:       genesisL2Block.Coinbase,
			ForkID:         genesisL2Block.ForkID,
		}

		log.Infof("Genesis block: %+v", genesisBlock)

		_, err = streamServer.AddStreamEntry(1, genesisBlock.Encode())
		if err != nil {
			return err
		}

		genesisBlockEnd := state.DSL2BlockEnd{
			L2BlockNumber: genesisL2Block.L2BlockNumber,
			BlockHash:     genesisL2Block.BlockHash,
			StateRoot:     genesisL2Block.StateRoot,
		}

		_, err = streamServer.AddStreamEntry(state.EntryTypeL2BlockEnd, genesisBlockEnd.Encode())
		if err != nil {
			return err
		}

		err = streamServer.CommitAtomicOp()
		if err != nil {
			return err
		}
	} else {
		latestEntry, err := streamServer.GetEntry(header.TotalEntries - 1)
		if err != nil {
			return err
		}

		log.Infof("Latest entry: %+v", latestEntry)

		switch latestEntry.Type {
		case state.EntryTypeUpdateGER:
			log.Info("Latest entry type is UpdateGER")
			currentBatchNumber = binary.LittleEndian.Uint64(latestEntry.Data[0:8])
		case state.EntryTypeL2BlockEnd:
			log.Info("Latest entry type is L2BlockEnd")
			currentL2Block = binary.LittleEndian.Uint64(latestEntry.Data[0:8])

			bookMark := state.DSBookMark{
				Type:          state.BookMarkTypeL2Block,
				L2BlockNumber: currentL2Block,
			}

			firstEntry, err := streamServer.GetFirstEventAfterBookmark(bookMark.Encode())
			if err != nil {
				return err
			}
			currentBatchNumber = binary.LittleEndian.Uint64(firstEntry.Data[0:8])
		}
	}

	log.Infof("Current Batch number: %d", currentBatchNumber)
	log.Infof("Current L2 block number: %d", currentL2Block)

	var entry uint64 = header.TotalEntries
	var currentGER = common.Hash{}

	if entry > 0 {
		entry--
	}

	// Start on the current batch number + 1
	currentBatchNumber++

	var err error

	const limit = 10000

	for err == nil {
		log.Debugf("Current entry number: %d", entry)
		log.Debugf("Current batch number: %d", currentBatchNumber)
		// Get Next Batch
		batches, err := stateDB.GetDSBatches(ctx, currentBatchNumber, currentBatchNumber+limit, readWIPBatch, nil)
		if err != nil {
			if err == state.ErrStateNotSynchronized {
				break
			}
			log.Errorf("Error getting batch %d: %s", currentBatchNumber, err.Error())
			return err
		}

		// Finished?
		if len(batches) == 0 {
			break
		}

		l2Blocks, err := stateDB.GetDSL2Blocks(ctx, batches[0].BatchNumber, batches[len(batches)-1].BatchNumber, nil)
		if err != nil {
			log.Errorf("Error getting L2 blocks for batches starting at %d: %s", batches[0].BatchNumber, err.Error())
			return err
		}

		l2Txs := make([]*state.DSL2Transaction, 0)
		if len(l2Blocks) > 0 {
			l2Txs, err = stateDB.GetDSL2Transactions(ctx, l2Blocks[0].L2BlockNumber, l2Blocks[len(l2Blocks)-1].L2BlockNumber, nil)
			if err != nil {
				log.Errorf("Error getting L2 transactions for blocks starting at %d: %s", l2Blocks[0].L2BlockNumber, err.Error())
				return err
			}
		}

		// Gererate full batches
		fullBatches := state.ComputeFullBatches(batches, l2Blocks, l2Txs)
		currentBatchNumber += limit

		for _, batch := range fullBatches {
			if len(batch.L2Blocks) == 0 {
				// Empty batch
				// Is WIP Batch?
				if batch.StateRoot == (common.Hash{}) {
					continue
				}
				// Check if there is a GER update
				if batch.GlobalExitRoot != currentGER && batch.GlobalExitRoot != (common.Hash{}) {
					updateGer := state.DSUpdateGER{
						BatchNumber:    batch.BatchNumber,
						Timestamp:      batch.Timestamp.Unix(),
						GlobalExitRoot: batch.GlobalExitRoot,
						Coinbase:       batch.Coinbase,
						ForkID:         batch.ForkID,
						StateRoot:      batch.StateRoot,
					}

					err = streamServer.StartAtomicOp()
					if err != nil {
						return err
					}

					entry, err = streamServer.AddStreamEntry(state.EntryTypeUpdateGER, updateGer.Encode())
					if err != nil {
						return err
					}

					err = streamServer.CommitAtomicOp()
					if err != nil {
						return err
					}

					currentGER = batch.GlobalExitRoot
				}
				continue
			}

			err = streamServer.StartAtomicOp()
			if err != nil {
				return err
			}

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
					return err
				}

				_, err = streamServer.AddStreamEntry(state.EntryTypeL2BlockStart, blockStart.Encode())
				if err != nil {
					return err
				}

				for _, tx := range l2block.Txs {
					// Populate intermediate state root
					if imStateRoots == nil || (*imStateRoots)[blockStart.L2BlockNumber] == nil {
						position := state.GetSystemSCPosition(l2block.L2BlockNumber)
						imStateRoot, err := stateDB.GetStorageAt(ctx, common.HexToAddress(state.SystemSC), big.NewInt(0).SetBytes(position), l2block.StateRoot)
						if err != nil {
							return err
						}
						tx.StateRoot = common.BigToHash(imStateRoot)
					} else {
						tx.StateRoot = common.BytesToHash((*imStateRoots)[blockStart.L2BlockNumber])
					}

					entry, err = streamServer.AddStreamEntry(state.EntryTypeL2Tx, tx.Encode())
					if err != nil {
						return err
					}
				}

				blockEnd := state.DSL2BlockEnd{
					L2BlockNumber: l2block.L2BlockNumber,
					BlockHash:     l2block.BlockHash,
					StateRoot:     l2block.StateRoot,
				}

				_, err = streamServer.AddStreamEntry(state.EntryTypeL2BlockEnd, blockEnd.Encode())
				if err != nil {
					return err
				}
				currentGER = l2block.GlobalExitRoot
			}
			// Commit at the end of each batch group
			err = streamServer.CommitAtomicOp()
			if err != nil {
				return err
			}
		}
	}

	return err
}
