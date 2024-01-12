package ethtxmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/hex"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

type signRequest struct {
	UserID         int            `json:"userId"`
	OperateType    int            `json:"operateType"` // 1 seq 2 agg
	OperateAddress common.Address `json:"operateAddress"`
	Symbol         int            `json:"symbol"`        // devnet 2882 mainnet 2
	ProjectSymbol  int            `json:"projectSymbol"` // 3011
	RefOrderID     string         `json:"refOrderId"`
	OperateSymbol  int            `json:"operateSymbol"` // 2
	OperateAmount  int            `json:"operateAmount"` // 0
	SysFrom        int            `json:"sysFrom"`       // 3
	OtherInfo      string         `json:"otherInfo"`     // ""
}

type signResponse struct {
	Code           int    `json:"code"`
	Data           string `json:"data"`
	DetailMessages string `json:"detailMsg"`
	Msg            string `json:"msg"`
	Status         int    `json:"status"`
	Success        bool   `json:"success"`
}

type signResultRequest struct {
	UserID        int    `json:"userId"`
	OrderID       string `json:"orderId"`
	ProjectSymbol int    `json:"projectSymbol"`
}

func (c *Client) newSignRequest(operateType int, operateAddress common.Address, otherInfo string) *signRequest {
	refOrderID := uuid.New().String()
	return &signRequest{
		UserID:         c.cfg.CustodialAssets.UserID,
		OperateType:    operateType,
		OperateAddress: operateAddress,
		Symbol:         c.cfg.CustodialAssets.Symbol,
		ProjectSymbol:  c.cfg.CustodialAssets.ProjectSymbol,
		RefOrderID:     refOrderID,
		OperateSymbol:  c.cfg.CustodialAssets.OperateSymbol,
		OperateAmount:  c.cfg.CustodialAssets.OperateAmount,
		SysFrom:        c.cfg.CustodialAssets.SysFrom,
		OtherInfo:      otherInfo,
	}
}

func (c *Client) newSignResultRequest(orderID string) *signResultRequest {
	return &signResultRequest{
		UserID:        c.cfg.CustodialAssets.UserID,
		OrderID:       orderID,
		ProjectSymbol: c.cfg.CustodialAssets.ProjectSymbol,
	}
}

func (c *Client) postCustodialAssets(ctx context.Context, request *signRequest) error {
	if c == nil || !c.cfg.CustodialAssets.Enable {
		return errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(getTraceID(ctx))

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshal request: %w", err)
	}

	reqSignURL, err := url.JoinPath(c.cfg.CustodialAssets.URL, c.cfg.CustodialAssets.RequestSignURI)
	if err != nil {
		return fmt.Errorf("error join url: %w", err)
	}

	req, err := http.NewRequest("POST", reqSignURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	mLog.Infof("post custodial assets request: %v", string(payload))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	var signResp signResponse
	err = json.Unmarshal(body, &signResp)
	if err != nil {
		return fmt.Errorf("error unmarshal %v response body: %w", resp, err)
	}
	if signResp.Status != 200 || !signResp.Success {
		return fmt.Errorf("error response %v status: %v", signResp, signResp.Status)
	}

	return nil
}

func (c *Client) querySignResult(ctx context.Context, request *signResultRequest) (*signResponse, error) {
	if c == nil || !c.cfg.CustodialAssets.Enable {
		return nil, errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(getTraceID(ctx))
	mLog.Infof("get sign result request: %v", request)

	querySignURL, err := url.JoinPath(c.cfg.CustodialAssets.URL, c.cfg.CustodialAssets.QuerySignURI)
	if err != nil {
		return nil, fmt.Errorf("error join url: %w", err)
	}
	params := url.Values{}
	params.Add("orderId", request.OrderID)
	params.Add("projectSymbol", fmt.Sprintf("%d", request.ProjectSymbol))
	fullQuerySignURL := fmt.Sprintf("%s?%s", querySignURL, params.Encode())

	req, err := http.NewRequest("GET", fullQuerySignURL, nil)
	// response, err := http.Get(fullQuerySignURL)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var signResp signResponse
	err = json.Unmarshal(body, &signResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal %v response body: %w", response, err)
	}

	if signResp.Status != 200 || len(signResp.Data) == 0 {
		return nil, fmt.Errorf("error response %v status: %v", signResp, signResp.Status)
	}

	return &signResp, nil
}

func (c *Client) waitResult(parentCtx context.Context, request *signResultRequest) (*signResponse, error) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()
	ctx, cancel := context.WithTimeout(parentCtx, c.cfg.CustodialAssets.WaitResultTimeout.Duration)
	defer cancel()

	mLog := log.WithFields(getTraceID(ctx))
	for {
		result, err := c.querySignResult(ctx, request)
		if err == nil {
			mLog.Infof("query sign result success: %v", result)
			return result, nil
		}
		mLog.Infof("query sign result failed: %v", err)

		// Wait for the next round.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

func (c *Client) postSignRequestAndWaitResult(ctx context.Context, mTx monitoredTx, request *signRequest) (*types.Transaction, error) {
	if c == nil || !c.cfg.CustodialAssets.Enable {
		return nil, errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(getTraceID(ctx))
	mLog.Infof("post custodial assets request: %v", request)
	if err := c.postCustodialAssets(ctx, request); err != nil {
		return nil, fmt.Errorf("error post custodial assets: %w", err)
	}
	mLog.Infof("post custodial assets success")
	result, err := c.waitResult(ctx, c.newSignResultRequest(request.RefOrderID))
	if err != nil {
		return nil, fmt.Errorf("error wait result: %w", err)
	}
	mLog.Infof("wait result success: %v", result)
	data, err := hex.DecodeHex(result.Data)
	if err != nil {
		return nil, fmt.Errorf("error decode hex: %w", err)
	}
	transaction := &types.Transaction{}
	err = transaction.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal binary: %w", err)
	}
	mLog.Infof("unmarshal transaction success: %v", transaction.Hash())

	err = c.checkSignedTransaction(ctx, mTx, transaction, request)
	if err != nil {
		return nil, fmt.Errorf("error check signed transaction: %w", err)
	}

	return transaction, nil
}

func (c *Client) checkSignedTransaction(ctx context.Context, mTx monitoredTx, transaction *types.Transaction, request *signRequest) error {
	if c == nil || !c.cfg.CustodialAssets.Enable {
		return errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(getTraceID(ctx))
	mLog.Infof("check signed transaction: %v", transaction.Hash())

	contractAddress, err := c.etherman.GetZkEVMAddress()
	if err != nil {
		return fmt.Errorf("failed to get zkEVM address: %v", err)
	}
	var signedRequest string
	switch request.OperateType {
	case c.cfg.CustodialAssets.OperateTypeSeq:
		args, err := c.unpackSequenceBatchesTx(transaction)
		if err != nil {
			return fmt.Errorf("error unpack sequence batches tx: %w", err)
		}
		signedRequest, err = args.marshal(contractAddress, mTx)
		if err != nil {
			return fmt.Errorf("error marshal sequence batches tx: %w", err)
		}
	case c.cfg.CustodialAssets.OperateTypeAgg:
		args, err := c.unpackVerifyBatchesTrustedAggregatorTx(transaction)
		if err != nil {
			return fmt.Errorf("error unpack sequence batches tx: %w", err)
		}
		signedRequest, err = args.marshal(contractAddress, mTx)
		if err != nil {
			return fmt.Errorf("error marshal sequence batches tx: %w", err)
		}
	default:
		return fmt.Errorf("error operate type: %v", request.OperateType)
	}
	mLog.Infof("signed transaction nonce: %v to: %v gas limit: %v gas price: %v", transaction.Nonce(), transaction.To(), transaction.Gas(), transaction.GasPrice())
	mLog.Infof("mTx    transaction nonce: %v to: %v gas limit: %v gas price: %v", mTx.nonce, mTx.to.String(), mTx.gas+mTx.gasOffset, mTx.gasPrice.String())
	if signedRequest != request.OtherInfo {
		return fmt.Errorf("signed transaction not equal with other info: %v, %v", signedRequest, request.OtherInfo)
	}
	if transaction.Nonce() != mTx.nonce {
		return fmt.Errorf("signed transaction nonce not equal with mTx: %v, %v", transaction.Nonce(), mTx.nonce)
	}
	if transaction.To().String() != mTx.to.String() {
		return fmt.Errorf("signed transaction to not equal with mTx: %v, %v", transaction.To(), mTx.to)
	}
	from, err := types.Sender(types.LatestSignerForChainID(transaction.ChainId()), transaction)
	if err != nil {
		return fmt.Errorf("error get sender: %w", err)
	}
	if from.String() != mTx.from.String() {
		return fmt.Errorf("signed transaction from not equal with mTx: %v, %v", from, mTx.from)
	}
	if transaction.Gas() != mTx.gas+mTx.gasOffset {
		return fmt.Errorf("signed transaction gas limit not equal with mTx: %v, %v", transaction.Gas(), mTx.gas)
	}
	if transaction.GasPrice().Cmp(mTx.gasPrice) != 0 {
		return fmt.Errorf("signed transaction gas price not equal with mTx: %v, %v", transaction.GasPrice().String(), mTx.gasPrice.String())
	}

	return nil
}
