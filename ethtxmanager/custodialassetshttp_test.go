package ethtxmanager

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func TestClient_postSignRequestAndWaitResult(t *testing.T) {
	client := &Client{
		cfg: Config{
			CustodialAssetsConfig: CustodialAssetsConfig{
				Enable:            false,
				URL:               "http://asset-onchain.base-defi.svc.test.local:7001",
				Symbol:            2882,
				SequencerAddr:     common.HexToAddress("1a13bddcc02d363366e04d4aa588d3c125b0ff6f"),
				AggregatorAddr:    common.HexToAddress("66e39a1e507af777e8c385e2d91559e20e306303"),
				WaitResultTimeout: 2 * time.Minute,
			},
		},
	}
	ctx := context.WithValue(context.Background(), traceID, uuid.New().String())
	txInput, _ := hex.DecodeHex("0x2b0006fa0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001876f0000000000000000000000000000000000000000000000000000000000018783f0341a0c4e4c69cc9fd2e64b6e8d84ae351184d0990a4798b5553e0697627cc8c1ae9298860d9a0fb9f7d0b0f811071b044c38923238875481b35a0a137ca53c300423c13c9bceac761575d31361d57a9984cfca6b6b7ed32701b19d080a4f452b744315d3ecbaf535f1c44639f397cbf0138b85f74dafa69b645d44ab0a0b001ceae6a20f736343e95c1f0dfdf471855e241a47ac04cd3f4466a5aca1d9703f13a20fa5c363efaf7c0785ea5a3e902ebfe3c95ba993618de67211a6697e7d340034b4b2f3513ae869839e4c3856aa7d07f7bfccfd8cdb24f3a97e70b045cc6d03156b4ee318f0a5b166a27783823442c8687c4372f7535049e152318762dc8809cd7d034ddb8f08b5bae2f7c02aedc7a0e92d23ef8096ec24d25c0b353ae05912ffe1bf1c399cd127cb8a8df070e9c60ba21176fca766bb76e47cfec736b84407a627f417641549b3eddea0036a814fbb6ca97749b7ac1973f024cb4c01f455025b970fb95faefc731cb83ebc0543aff9bf1396d1c4603dbdad7fbe14f8bb36001f93d39d06842a55b659c923b1f9c3952809ec578f219627e741586729eda6230c964a726afb3184d1c2211b50448a8647e674e90dbdc812c8fd7eda855d6220f9135eaf56139166ab5c1d5856ec2b299bfec2dc7e633d97e382e8524720552cf1bb9f34fc16c3800ea66829f80496d667d2f9d65c316e2e456f7c30b05fb41b33e1b04173abea059d06a49e3a19054b07afa5f9916fb1dde5e3535f67564a161cf679e60fc04d14383b196a6ffd57d4c9ba1a7de54236b7d562ce007c83e718f5b12bab4ca956efa20c2fb9fb163db43009ee18e77a54e6daeaca5a253c9e2b4aff57fc0f58f2871c15bf8cc67cc4eaaeeecbadd2f8d41ee7f941319b14020957753f325753a6bd8c0d0f389825f4ec9b1724acc6c332b5c1a0d1f9086eaf1e9d242a9ae740c884bdf96358e3eebac1cc4a6d1e2204ed28368a1588e7bc8f054f5261798152d468bc72d3f3b3653bb3744f3bc4121c8bc1ef165c2cccacaf2007007ec3ad96fd544eaf9d0b1793756809138dad1ef80afa086ce285ca9e861988cfe96051b8cf1a73442fcad68d599c44a13b34ac70e70236592b1120762b170f0bb47059297a12c41dfbd2ebe688597377b9b95bc57ac73fc11b932fa9d7")

	tx := types.NewTransaction(0, common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87"), big.NewInt(10), 50000, big.NewInt(10), txInput)

	seqReq, _ := client.unpackVerifyBatchesTrustedAggregatorTx(tx)
	ret, _ := seqReq.marshal()

	req := client.newSignRequest(2, client.cfg.CustodialAssetsConfig.AggregatorAddr, ret)

	mTx := monitoredTx{
		from: common.HexToAddress("d6dda5aa7749142b7fda3fe4662c9f346101b8a6"),
	}
	_, err := client.postSignRequestAndWaitResult(ctx, mTx, req)
	if err != nil {
		t.Log(err)
	}
}
