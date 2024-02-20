package gasprice

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/log"
)

type BasicGasPricer struct {
	cfg  Config
	pool poolInterface
	ctx  context.Context

	lastL2BlockNumber uint64
	lastPrice         *big.Int

	cacheLock sync.RWMutex
	fetchLock sync.Mutex
	state     stateInterface
}

func (b *BasicGasPricer) calDynamicGP() {
	l2BlockNumber, err := b.state.GetLastL2BlockNumber(b.ctx, nil)
	if err != nil {
		log.Errorf("failed to get last l2 block number, err: %v", err)
	}
	b.cacheLock.RLock()
	lastL2BlockNumber, lastPrice := b.lastL2BlockNumber, new(big.Int).Set(b.lastPrice)
	b.cacheLock.RUnlock()
	if l2BlockNumber == lastL2BlockNumber {
		log.Debug("Block is still the same, no need to update the gas price at the moment, lastL2BlockNumber: ", lastL2BlockNumber)
		return
	}

	b.fetchLock.Lock()
	defer b.fetchLock.Unlock()

	var (
		sent, exp int
		number    = lastL2BlockNumber
		result    = make(chan results, b.cfg.CheckBlocks)
		quit      = make(chan struct{})
		results   []*big.Int
	)

	for sent < b.cfg.CheckBlocks && number > 0 {
		go b.getL2BlockTxsTips(number, sampleNumber, b.cfg.IgnorePrice, result, quit)
		sent++
		exp++
		number--
	}

	for exp > 0 {
		res := <-result
		if res.err != nil {
			close(quit)
			return
		}
		exp--

		if len(res.values) == 0 {
			res.values = []*big.Int{lastPrice}
		}
		results = append(results, res.values...)
	}

	price := lastPrice
	if len(results) > 0 {
		sort.Sort(bigIntArray(results))
		price = results[(len(results)-1)*b.cfg.Percentile/100]
	}

	b.cacheLock.Lock()
	b.lastPrice = price
	b.lastL2BlockNumber = l2BlockNumber
	b.cacheLock.Unlock()
}

func (b *BasicGasPricer) getL2BlockTxsTips(l2BlockNumber uint64, limit int, ignorePrice *big.Int, result chan results, quit chan struct{}) {
	txs, err := b.state.GetTxsByBlockNumber(b.ctx, l2BlockNumber, nil)
	if txs == nil {
		select {
		case result <- results{nil, err}:
		case <-quit:
		}
		return
	}
	sorter := newSorter(txs)
	sort.Sort(sorter)

	var prices []*big.Int
	var lowPrices []*big.Int
	var highPrices []*big.Int
	for _, tx := range sorter.txs {
		tip := tx.GasTipCap()
		if ignorePrice != nil && tip.Cmp(ignorePrice) == -1 {
			continue
		}
		lowPrices = append(lowPrices, tip)
		if len(lowPrices) >= limit {
			break
		}
	}

	sorter.Reverse()
	for _, tx := range sorter.txs {
		tip := tx.GasTipCap()
		if ignorePrice != nil && tip.Cmp(ignorePrice) == -1 {
			continue
		}
		highPrices = append(highPrices, tip)
		if len(highPrices) >= limit {
			break
		}
	}

	if len(highPrices) != len(lowPrices) {
		err := errors.New("len(highPrices) != len(lowPrices)")
		log.Errorf("getL2BlockTxsTips err: %v", err)
		select {
		case result <- results{nil, err}:
		case <-quit:
		}
		return
	}

	for i := 0; i < len(lowPrices); i++ {
		price := getAvgPrice(lowPrices[i], highPrices[i])
		prices = append(prices, price)
	}

	select {
	case result <- results{prices, nil}:
	case <-quit:
	}
}

func (b *BasicGasPricer) isCongested() (bool, error) {
	txCount, err := b.pool.CountPendingTransactions(b.ctx)
	if err != nil {
		return false, err
	}
	if txCount >= b.cfg.CongestionTxThreshold {
		return true, nil
	}
	return false, nil
}

func getAvgPrice(low *big.Int, high *big.Int) *big.Int {
	avg := new(big.Int).Add(low, high)
	avg = avg.Quo(avg, big.NewInt(2)) //nolint:gomnd
	return avg
}
