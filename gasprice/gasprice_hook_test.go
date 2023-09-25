package gasprice

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCalculateRate(t *testing.T) {
	testcases := []struct {
		l1CoinId int
		l2CoinId int
		msg      string
		check    func(rate float64, err error)
	}{
		{
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\"}"),
			check: func(rate float64, err error) {
				require.Error(t, err)
			},
		},
		{
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.02}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId),
			check: func(rate float64, err error) {
				require.Error(t, err)
			},
		},
		{
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.02}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", okbcoinId),
			check: func(rate float64, err error) {
				require.Error(t, err)
			},
		},
		{
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.00000000000001}, {\"coinId\":%d,\"price\":0.002}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId, okbcoinId),
			check: func(rate float64, err error) {
				require.Error(t, err)
			},
		},
		{
			// correct
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.02}, {\"coinId\":%d,\"price\":0.002}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId, okbcoinId),
			check: func(rate float64, err error) {
				require.Equal(t, rate, 0.1)
				require.NoError(t, err)
			},
		},
		{
			// correct
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.002}, {\"coinId\":%d,\"price\":0.02}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId, okbcoinId),
			check: func(rate float64, err error) {
				require.Equal(t, rate, float64(10))
				require.NoError(t, err)
			},
		},
		{
			// correct
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.04}, {\"coinId\":%d,\"price\":0.002}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId, okbcoinId),
			check: func(rate float64, err error) {
				require.Equal(t, rate, 0.05)
				require.NoError(t, err)
			},
		},
		{
			// correct
			l1CoinId: ethcoinId,
			l2CoinId: okbcoinId,
			msg:      fmt.Sprintf("{\"topic\":\"middle_coinPrice_push\",\"source\":null,\"type\":null,\"data\":{\"priceList\":[{\"coinId\":%d,\"price\":0.04}, {\"coinId\":%d,\"price\":0.002}, {\"coinId\":123,\"price\":0.005}],\"id\":\"98a797ce-f61b-4e90-87ac-445e77ad3599\"}}", ethcoinId, okbcoinId),
			check: func(rate float64, err error) {
				require.Equal(t, rate, 0.05)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		rp := newRateProcessor(Config{Topic: "middle_coinPrice_push", L1CoinId: tc.l1CoinId, L2CoinId: tc.l2CoinId}, context.Background())
		rt, err := rp.calculateRate([]byte(tc.msg))
		tc.check(rt, err)
	}
}
