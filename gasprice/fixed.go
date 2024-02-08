package gasprice

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/0xPolygonHermez/zkevm-node/encoding"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"sync"
	"sort"
	"errors"
)

const (
	// OKBWei OKB wei
	OKBWei       = 1e18
	minCoinPrice = 1e-18
)

// FixedGasPrice struct
type FixedGasPrice struct {
	cfg     Config
	pool    poolInterface
	ctx     context.Context
	eth     ethermanInterface
	ratePrc *KafkaProcessor

	lastL2BlockNumber uint64
	lastPrice         *big.Int

	cacheLock sync.RWMutex
	fetchLock sync.Mutex
	state     stateInterface

	apolloConfig Apollo
}

// newFixedGasPriceSuggester inits l2 fixed price suggester.
func newFixedGasPriceSuggester(ctx context.Context, cfg Config, state stateInterface, pool poolInterface, ethMan ethermanInterface, fetch Apollo) *FixedGasPrice {
	gps := &FixedGasPrice{
		cfg:       cfg,
		pool:      pool,
		ctx:       ctx,
		eth:       ethMan,
		ratePrc:   newKafkaProcessor(cfg, ctx),
		state:     state,
		lastPrice: new(big.Int).SetUint64(cfg.DefaultGasPriceWei),

		apolloConfig: fetch,
	}
	gps.UpdateGasPriceAvg()
	return gps
}

// UpdateGasPriceAvg updates the gas price.
func (f *FixedGasPrice) UpdateGasPriceAvg() {
	if f.apolloConfig != nil {
		f.apolloConfig.FetchL2GasPricerConfig(&f.cfg)
	}
	ctx := context.Background()
	// Get L1 gasprice
	l1GasPrice := f.eth.GetL1GasPrice(f.ctx)
	if big.NewInt(0).Cmp(l1GasPrice) == 0 {
		log.Warn("gas price 0 received. Skipping update...")
		return
	}

	l2CoinPrice := f.ratePrc.GetL2CoinPrice()
	if l2CoinPrice < minCoinPrice {
		log.Warn("the L2 native coin price too small...")
		return
	}
	res := new(big.Float).Mul(big.NewFloat(0).SetFloat64(f.cfg.GasPriceUsdt/l2CoinPrice), big.NewFloat(0).SetFloat64(OKBWei))
	// Store l2 gasPrice calculated
	result := new(big.Int)
	res.Int(result)
	minGasPrice := big.NewInt(0).SetUint64(f.cfg.DefaultGasPriceWei)
	if minGasPrice.Cmp(result) == 1 { // minGasPrice > result
		log.Warn("setting DefaultGasPriceWei for L2")
		result = minGasPrice
	}
	maxGasPrice := new(big.Int).SetUint64(f.cfg.MaxGasPriceWei)
	if f.cfg.MaxGasPriceWei > 0 && result.Cmp(maxGasPrice) == 1 { // result > maxGasPrice
		log.Warn("setting MaxGasPriceWei for L2")
		result = maxGasPrice
	}

	if f.cfg.EnableDynamicFixed {
		//todo: judge if there is congestion
		log.Debug("enable dynamic fixed strategy")
		f.calDynamicGPFromLastNBatches()
		if result.Cmp(f.lastPrice) < 0 {
			result = new(big.Int).Set(f.lastPrice)
		}
	}

	var truncateValue *big.Int
	log.Debug("Full L2 gas price value: ", result, ". Length: ", len(result.String()), ". L1 gas price value: ", l1GasPrice)

	numLength := len(result.String())
	if numLength > 3 { //nolint:gomnd
		aux := "%0" + strconv.Itoa(numLength-3) + "d" //nolint:gomnd
		var ok bool
		value := result.String()[:3] + fmt.Sprintf(aux, 0)
		truncateValue, ok = new(big.Int).SetString(value, encoding.Base10)
		if !ok {
			log.Error("error converting: ", truncateValue)
		}
	} else {
		truncateValue = result
	}
	log.Debugf("Storing truncated L2 gas price: %v, L2 native coin price: %v", truncateValue, l2CoinPrice)
	if truncateValue != nil {
		log.Infof("Set gas prices, L1: %v, L2: %v", l1GasPrice.Uint64(), truncateValue.Uint64())
		err := f.pool.SetGasPrices(ctx, truncateValue.Uint64(), l1GasPrice.Uint64())
		if err != nil {
			log.Errorf("failed to update gas price in poolDB, err: %v", err)
		}
	} else {
		log.Error("nil value detected. Skipping...")
	}
}

func (f *FixedGasPrice) calDynamicGPFromLastNBatches() {
	l2BlockNumber, err := f.state.GetLastL2BlockNumber(f.ctx, nil)
	if err != nil {
		log.Errorf("failed to get last l2 block number, err: %v", err)
	}
	f.cacheLock.RLock()
	lastL2BlockNumber, lastPrice := f.lastL2BlockNumber, f.lastPrice
	f.cacheLock.RUnlock()
	if l2BlockNumber == lastL2BlockNumber {
		log.Debug("Block is still the same, no need to update the gas price at the moment, lastL2BlockNumber: ", lastL2BlockNumber)
		return
	}

	f.fetchLock.Lock()
	defer f.fetchLock.Unlock()

	var (
		sent, exp int
		number    = lastL2BlockNumber
		result    = make(chan results, f.cfg.CheckBlocks)
		quit      = make(chan struct{})
		results   []*big.Int
	)

	for sent < f.cfg.CheckBlocks && number > 0 {
		go f.getL2BlockTxsTips(f.ctx, number, sampleNumber, f.cfg.IgnorePrice, result, quit)
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
		price = results[(len(results)-1)*f.cfg.Percentile/100]
	}
	if price.Cmp(f.cfg.MaxPrice) > 0 {
		price = f.cfg.MaxPrice
	}

	f.cacheLock.Lock()
	f.lastPrice = price
	f.lastL2BlockNumber = l2BlockNumber
	f.cacheLock.Unlock()

	log.Debug("MaxPrice: ", f.cfg.MaxPrice)
	log.Debug("IgnorePrice: ", f.cfg.IgnorePrice)
	log.Debug("Factor: ", f.cfg.Factor)
	log.Debug("lastPrice: ", f.lastPrice)
}

func (f *FixedGasPrice) getL2BlockTxsTips(ctx context.Context, l2BlockNumber uint64, limit int, ignorePrice *big.Int, result chan results, quit chan struct{}) {
	txs, err := f.state.GetTxsByBlockNumber(ctx, l2BlockNumber, nil)
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

func getAvgPrice(low *big.Int, high *big.Int) *big.Int {
	avg := new(big.Int).Add(low, high)
	avg = avg.Quo(avg, big.NewInt(2))
	return avg
}
