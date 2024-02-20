package gasprice

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/log"
)

func calDynamicGP(ctx context.Context, cfg Config, state stateInterface, lastL2BlockNum *uint64, lastGP *big.Int, cacheLock *sync.RWMutex, fetchLock *sync.Mutex) {
	l2BlockNumber, err := state.GetLastL2BlockNumber(ctx, nil)
	if err != nil {
		log.Errorf("failed to get last l2 block number, err: %v", err)
	}
	cacheLock.RLock()
	lastL2BlockNumber, lastPrice := *lastL2BlockNum, new(big.Int).Set(lastGP)
	cacheLock.RUnlock()
	if l2BlockNumber == lastL2BlockNumber {
		log.Debug("Block is still the same, no need to update the gas price at the moment, lastL2BlockNumber: ", lastL2BlockNumber)
		return
	}

	fetchLock.Lock()
	defer fetchLock.Unlock()

	var (
		sent, exp int
		number    = lastL2BlockNumber
		result    = make(chan results, cfg.CheckBlocks)
		quit      = make(chan struct{})
		results   []*big.Int
	)

	for sent < cfg.CheckBlocks && number > 0 {
		go getL2BlockTxsTips(ctx, state, number, sampleNumber, cfg.IgnorePrice, result, quit)
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
		price = results[(len(results)-1)*cfg.Percentile/100]
	}

	cacheLock.Lock()
	lastGP = price
	lastL2BlockNum = &l2BlockNumber
	cacheLock.Unlock()
}

func getL2BlockTxsTips(ctx context.Context, state stateInterface, l2BlockNumber uint64, limit int, ignorePrice *big.Int, result chan results, quit chan struct{}) {
	txs, err := state.GetTxsByBlockNumber(ctx, l2BlockNumber, nil)
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

func isCongested(ctx context.Context, cfg Config, pool poolInterface) (bool, error) {
	txCount, err := pool.CountPendingTransactions(ctx)
	if err != nil {
		return false, err
	}
	if txCount >= cfg.CongestionTxThreshold {
		return true, nil
	}
	return false, nil
}

func getAvgPrice(low *big.Int, high *big.Int) *big.Int {
	var divisor int64 = 2
	avg := new(big.Int).Add(low, high)
	avg = avg.Quo(avg, big.NewInt(divisor))
	return avg
}
