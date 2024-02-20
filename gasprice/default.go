package gasprice

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/0xPolygonHermez/zkevm-node/log"
)

// DefaultGasPricer gas price from config is set.
type DefaultGasPricer struct {
	cfg        Config
	pool       poolInterface
	ctx        context.Context
	l1GasPrice uint64

	lastL2BlockNumber uint64
	lastPrice         *big.Int

	cacheLock sync.RWMutex
	fetchLock sync.Mutex
	state     stateInterface

	apolloConfig Apollo
}

// newDefaultGasPriceSuggester init default gas price suggester.
func newDefaultGasPriceSuggester(ctx context.Context, cfg Config, state stateInterface, pool poolInterface, fetch Apollo) *DefaultGasPricer {
	// Apply factor to calculate l1 gasPrice
	factorAsPercentage := int64(cfg.Factor * 100) // nolint:gomnd
	factor := big.NewInt(factorAsPercentage)
	defaultGasPriceDivByFactor := new(big.Int).Div(new(big.Int).SetUint64(cfg.DefaultGasPriceWei), factor)

	gpe := &DefaultGasPricer{
		ctx:        ctx,
		cfg:        cfg,
		pool:       pool,
		l1GasPrice: new(big.Int).Mul(defaultGasPriceDivByFactor, big.NewInt(100)).Uint64(), // nolint:gomnd
		state:      state,
		lastPrice:  new(big.Int).SetUint64(cfg.DefaultGasPriceWei),

		apolloConfig: fetch,
	}
	gpe.UpdateGasPriceAvg()
	return gpe
}

// UpdateGasPriceAvg not needed for default strategy.
func (d *DefaultGasPricer) UpdateGasPriceAvg() {
	if d.apolloConfig != nil {
		d.apolloConfig.FetchL2GasPricerConfig(&d.cfg)
	}

	result := new(big.Int).SetUint64(d.cfg.DefaultGasPriceWei)
	if d.cfg.EnableDynamicGP {

		log.Debug("enable dynamic gas price")
		// judge if there is congestion
		isCongested, err := isCongested(d.ctx, d.cfg, d.pool)
		if err != nil {
			log.Errorf("failed to count pool txs by status pending while judging if the pool is congested: ", err)
		}
		if isCongested {
			log.Warn("there is congestion for L2")
			calDynamicGP(d.ctx, d.cfg, d.state, &d.lastL2BlockNumber, d.lastPrice, &d.cacheLock, &d.fetchLock)
			if result.Cmp(d.lastPrice) < 0 {
				result = new(big.Int).Set(d.lastPrice)
			}
		}
		minGasPrice := big.NewInt(0).SetUint64(d.cfg.DefaultGasPriceWei)
		if minGasPrice.Cmp(result) == 1 { // minGasPrice > result
			log.Warn("setting DefaultGasPriceWei for L2")
			result = minGasPrice
		}
		maxGasPrice := new(big.Int).SetUint64(d.cfg.MaxGasPriceWei)
		if d.cfg.MaxGasPriceWei > 0 && result.Cmp(maxGasPrice) == 1 { // result > maxGasPrice
			log.Warn("setting MaxGasPriceWei for L2")
			result = maxGasPrice
		}
	}

	err := d.pool.SetGasPrices(d.ctx, result.Uint64(), d.l1GasPrice)
	if err != nil {
		panic(fmt.Errorf("failed to set default gas price, err: %v", err))
	}
}

func (d *DefaultGasPricer) setDefaultGasPrice() {
	err := d.pool.SetGasPrices(d.ctx, d.cfg.DefaultGasPriceWei, d.l1GasPrice)
	if err != nil {
		panic(fmt.Errorf("failed to set default gas price, err: %v", err))
	}
}
