package gasprice

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygonHermez/zkevm-node/log"
)

// DefaultGasPricer gas price from config is set.
type DefaultGasPricer struct {
	BasicGasPricer

	l1GasPrice uint64

	apolloConfig Apollo
}

// newDefaultGasPriceSuggester init default gas price suggester.
func newDefaultGasPriceSuggester(ctx context.Context, cfg Config, state stateInterface, pool poolInterface, fetch Apollo) *DefaultGasPricer {

	gpe := &DefaultGasPricer{
		BasicGasPricer: BasicGasPricer{
			cfg:       cfg,
			pool:      pool, // nolint:gomnd
			ctx:       ctx,
			lastPrice: new(big.Int).SetUint64(cfg.DefaultGasPriceWei),
			state:     state,
		},
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
		isCongested, err := d.isCongested()
		if err != nil {
			log.Errorf("failed to count pool txs by status pending while judging if the pool is congested: ", err)
		}
		if isCongested {
			log.Warn("there is congestion for L2")
			d.calDynamicGP()
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

	// Apply factor to calculate l1 gasPrice
	factorAsPercentage := int64(d.cfg.Factor * 100) // nolint:gomnd
	factor := big.NewInt(factorAsPercentage)
	gasPriceDivByFactor := new(big.Int).Div(result, factor)

	d.l1GasPrice = new(big.Int).Mul(gasPriceDivByFactor, big.NewInt(100)).Uint64() // nolint:gomnd
	err := d.pool.SetGasPrices(d.ctx, result.Uint64(), d.l1GasPrice)
	if err != nil {
		panic(fmt.Errorf("failed to set default gas price, err: %v", err))
	}
}
