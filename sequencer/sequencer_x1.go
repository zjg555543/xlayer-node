package sequencer

import (
	"context"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/log"
)

var countinterval = 10

func (s *Sequencer) countPendingTx(ctx context.Context) {
	for {
		<-time.After(time.Second * time.Duration(countinterval))
		transactions, err := s.pool.CountPendingTransactions(ctx)
		if err != nil {
			log.Errorf("load pending tx from pool: %v", err)
			continue
		}

		pmetrics.PendingTxCount(int(transactions))
	}
}
