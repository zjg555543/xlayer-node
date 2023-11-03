package sequencesender

import (
	"context"
	"github.com/0xPolygon/cdk-data-availability/config"
	cfgTypes "github.com/0xPolygonHermez/zkevm-node/config/types"
	"github.com/0xPolygonHermez/zkevm-node/event"
	"github.com/0xPolygonHermez/zkevm-node/event/nileventstorage"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	state_interface    = new(stateMock)
	etherman_interface = new(ethermanMock)
	ethtxman_interface = new(ethTxManagerMock)
	ctx                context.Context
	cfg                = Config{
		WaitPeriodSendSequence: cfgTypes.Duration{
			Duration: 5,
		},
		LastBatchVirtualizationTimeMaxWaitPeriod: cfgTypes.Duration{
			Duration: 5,
		},
		MaxTxSizeForL1:  10,
		MaxBatchesForL1: 10,
		PrivateKey: cfgTypes.KeystoreFileConfig{
			Path:     "./test.keystore",
			Password: "testonly",
		},
		UseValidium: true,
	}
)

func TestSequenceSender_getSequencesToSend(t *testing.T) {
	eventStorage, err := nileventstorage.NewNilEventStorage()
	require.NoError(t, err)
	eventLog := event.NewEventLog(event.Config{}, eventStorage)

	priv, err := config.NewKeyFromKeystore(cfg.PrivateKey)

	sequenceSender, err := New(cfg, state_interface, etherman_interface, ethtxman_interface, eventLog, priv)
	require.NoError(t, err)
	ctx = context.Background()

	state_interface.On("GetLastVirtualBatchNum", ctx, nil).Return(uint64(9), nil)

	sequence, address, err := sequenceSender.getSequencesToSend(ctx)
	require.NoError(t, err)
	t.Log("sequence", sequence)
	t.Log("address", address)
}
