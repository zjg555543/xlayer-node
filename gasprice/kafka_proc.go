package gasprice

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/0xPolygonHermez/zkevm-node/log"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"strings"
	"sync"
	"time"
)

const (
	okbcoinId = 7184
)

// aliyun root ca
var rootCA = `-----BEGIN CERTIFICATE-----
MIIFKjCCAxICCQCdkV+iL/cBTzANBgkqhkiG9w0BAQsFADBWMQswCQYDVQQGEwJD
TjEQMA4GA1UECAwHQmVpamluZzEQMA4GA1UEBwwHQmVpamluZzEQMA4GA1UECgwH
QWxpYmFiYTERMA8GA1UEAwwIQWxpS2Fma2EwIBcNMjIwNTExMTAzOTMxWhgPMjEy
MjA0MTcxMDM5MzFaMFYxCzAJBgNVBAYTAkNOMRAwDgYDVQQIDAdCZWlqaW5nMRAw
DgYDVQQHDAdCZWlqaW5nMRAwDgYDVQQKDAdBbGliYWJhMREwDwYDVQQDDAhBbGlL
YWZrYTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAL315apERcpAkDAB
SY4A2bGrRZO4CXj4nvqbwEZ50f1HlwABjzUMKXES7lWrOwrnqZjSIgm5woqu+Pr4
sWhKFHN19SSnjeKilQoL8SzMk0p22QJK2sqKRMuHtoBtL6uOT+ykV16IEg0fY2Uu
/oX/sF2LAVCIl1IGc2HVKUr56c0/mM6V6Ur5Sum7ctKk2dm6YS5gwDOXcqAaZhwd
jVzqLEW8hmsMS7n+d2/NIJMqXvTHDRQ74xhR9tN2w92keEBGOQoMG/Qw0RvS1aQi
RKpNpvCE7z543istYuFbFji646u6kRCr7I2i4RwV0qXVM1djcS+PysUsIX4mEjdP
Kq0Fptzsii3aeTFuNswOlo5GieE3psVoymIP2HWd6xmlmFaX3Z8Nd4PxA6h0uRIY
tRbLkHw8WfOAl4dXxWQFkbvNYNLRB5xZUYjm3CA+ZhYfJRtNlPa2247Psnbup6CH
k3DP+aExdLmbtyugZO/lNqi9WMZ0qLFGXZDz8astgJPGKiCjihccpP1cdzGlCzGu
iE6S25JEBuXPl4wg4GXNuCg6tcEKL2qinvbrCimrilWuFajBh7hRH0dgkhezw6xU
+3++ZCebEJOXZ8byn3v/gmyx2PDnKlBPcXCy23nadbiX/zpNvNvCqAewajm9AlWY
fXbCl5TkUnyMPsh0rwWeeRYR2kM3AgMBAAEwDQYJKoZIhvcNAQELBQADggIBADxW
YJoWh9DVtwFGp8TOrlbZ7kwflKFv8Hew4SX00K5GwKgmnn3fjdR0F8rZ2ar/BqdD
zR63sv9LGjMci9NWAqPqN5MyKB97KrFV6nHzcYLRmT+ltolqcfp5MeGCqka7ZTEL
t658xxaSXNEY9HGHYskIu7mWd41KAj0RLRJnEEOrCSZzfpzG4LdD6J0u7wpyJSYL
jGxi2xswt5C0x790LS/JmFq65c/vzfATjbmu6XSO3UvtsADpj0pH3FJFhLzoT67o
NrUeFEHrzsMc7JenYmPIYmEb4xXlfctjCzLaiNG3u8uKwXGBk/oagAwXCsI8I0pR
wtW/QedXxlFtUfATRZnI/eLqvJ5cQ6aXg/GyJtAv+ccFf004K1ER00ECe738WNXm
+6NNkhN5gPhwsfoDhq+a7Zmvj9+x/XDjSRqZ8j+XIMi9ZQjTwUAg9JmnhyR4eJXn
oQAxGc3ii98YoAspKZGRX6LoRfYbNE3TXJsSzGw73+PqS1y74xNNmMx2XX6IV/53
Is5mA8fli6BIEKkAgE6Pn0t6v5EP6haVF84vJazYRIlYflR2mi8p8dU6kohiC79C
e4seRTTZgyXU+5dgFIXqagub2A79tRtPAr+4Xi84jzY84ceUwqX2fxRwkfaUUJb8
Hh2q+P+VJeK50B83DZ4ui+WNJbAaAbcLMsn/idX3
-----END CERTIFICATE-----`

type MsgInfo struct {
	Topic string `json:"topic"`
	Data  *Body  `json:"data"`
}

type Body struct {
	Id        string   `json:"id"`
	PriceList []*Price `json:"priceList"`
}

type Price struct {
	CoinId                   int     `json:"coinId"`
	Symbol                   string  `json:"symbol"`
	FullName                 string  `json:"fullName"`
	Price                    float64 `json:"price"`
	PriceStatus              int     `json:"priceStatus"`
	MaxPrice24H              float64 `json:"maxPrice24H"`
	MinPrice24H              float64 `json:"minPrice24H"`
	MarketCap                float64 `json:"marketCap"`
	Timestamp                int64   `json:"timestamp"`
	Vol24H                   float64 `json:"vol24h"`
	CirculatingSupply        float64 `json:"circulatingSupply"`
	MaxSupply                float64 `json:"maxSupply"`
	TotalSupply              float64 `json:"totalSupply"`
	PriceChange24H           float64 `json:"priceChange24H"`
	PriceChangeRate24H       float64 `json:"priceChangeRate24H"`
	CirculatingMarketCap     float64 `json:"circulatingMarketCap"`
	PriceChange7D            float64 `json:"priceChange7D"`
	PriceChangeRate7D        float64 `json:"priceChangeRate7D"`
	PriceChange30D           float64 `json:"priceChange30D"`
	PriceChangeRate30D       float64 `json:"priceChangeRate30D"`
	PriceChangeYearStart     float64 `json:"priceChangeYearStart"`
	PriceChangeRateYearStart float64 `json:"priceChangeRateYearStart"`
	ExceptionStatus          int     `json:"exceptionStatus"`
	Source                   int     `json:"source"`
	Type                     string  `json:"type"`
	Id                       string  `json:"id"`
}

type KafkaProcessor struct {
	kreader  *kafka.Reader
	L2Price  float64
	ctx      context.Context
	rwLock   sync.RWMutex
	l2CoinId int
}

func newKafkaProcessor(cfg Config, ctx context.Context) *KafkaProcessor {
	rp := &KafkaProcessor{
		kreader:  getKafkaReader(cfg),
		L2Price:  cfg.DefaultL2CoinPrice,
		ctx:      ctx,
		l2CoinId: okbcoinId,
	}
	if cfg.L2CoinId != 0 {
		rp.l2CoinId = cfg.L2CoinId
	}

	go rp.processor()
	return rp
}

func getKafkaReader(cfg Config) *kafka.Reader {
	brokers := strings.Split(cfg.KafkaURL, ",")

	var dialer *kafka.Dialer
	if cfg.Password != "" && cfg.Username != "" {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM([]byte(rootCA)); !ok {
			panic("caCertPool.AppendCertsFromPEM")
		}
		dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: plain.Mechanism{Username: cfg.Username, Password: cfg.Password},
			TLS:           &tls.Config{RootCAs: caCertPool, InsecureSkipVerify: true},
		}
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    1,    // 1
		MaxBytes:    10e6, // 10MB
		Dialer:      dialer,
		StartOffset: kafka.LastOffset, // read data from new message
	})
}

func (rp *KafkaProcessor) processor() {
	log.Info("kafka processor start processor ")
	defer rp.kreader.Close()
	for {
		select {
		case <-rp.ctx.Done():
			return
		default:
			value, err := rp.ReadAndCalc(rp.ctx)
			if err != nil {
				log.Warn("get the destion data fail ", err)
				time.Sleep(time.Second * 10)
				continue
			}
			rp.updateL2CoinPrice(value)
		}
	}
}

func (rp *KafkaProcessor) ReadAndCalc(ctx context.Context) (float64, error) {
	m, err := rp.kreader.ReadMessage(ctx)
	if err != nil {
		return 0, err
	}
	return rp.parseL2CoinPrice(m.Value)
}

func (rp *KafkaProcessor) updateL2CoinPrice(price float64) {
	rp.rwLock.Lock()
	defer rp.rwLock.Unlock()
	rp.L2Price = price
}

func (rp *KafkaProcessor) GetL2CoinPrice() float64 {
	rp.rwLock.RLock()
	defer rp.rwLock.RUnlock()
	return rp.L2Price
}

func (rp *KafkaProcessor) parseL2CoinPrice(value []byte) (float64, error) {
	msgI := &MsgInfo{}
	err := json.Unmarshal(value, &msgI)
	if err != nil {
		return 0, err
	}
	if msgI.Data == nil || len(msgI.Data.PriceList) == 0 {
		return 0, fmt.Errorf("the data PriceList is empty")
	}
	for _, price := range msgI.Data.PriceList {
		if price.CoinId == rp.l2CoinId {
			return price.Price, nil
		}
	}
	return 0, fmt.Errorf("not find a correct coin price coinId=%v", rp.l2CoinId)
}
