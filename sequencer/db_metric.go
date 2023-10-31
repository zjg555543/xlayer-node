package sequencer

import (
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/0xPolygonHermez/zkevm-node/sequencer/metrics"
	"time"
)

func (d *dbManager) countPendingTx() {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			transactions, err := d.txPool.CountPendingTransactions(d.ctx)
			if err != nil {
				log.Errorf("load pending tx from pool: %v", err)
				continue
			}
			metrics.PendingTxCount(int(transactions))
		}
	}
}
