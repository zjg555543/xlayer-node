package ethtxmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

const (
	operateTypeSeq = 1
	operateTypeAgg = 2
	projectSymbol  = 3011
	operateSymbol  = 2
	operatorAmount = 0
	sysFrom        = 3
	requestSignURI = "/priapi/v1/assetonchain/ecology/ecologyOperate"
	querySignURI   = "/priapi/v1/assetonchain/ecology/querySignDataByOrderNo"
)

type signRequest struct {
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
	DetailMessages string `json:"detailMessages"`
	Msg            string `json:"msg"`
	Status         int    `json:"status"`
	Success        bool   `json:"success"`
}

type signResultRequest struct {
	OrderID       string `json:"orderId"`
	ProjectSymbol int    `json:"projectSymbol"`
}

func (c *Client) newSignRequest(operateType int, operateAddress common.Address, otherInfo string) *signRequest {
	refOrderID := uuid.New().String()
	return &signRequest{
		OperateType:    operateType,
		OperateAddress: operateAddress,
		Symbol:         c.cfg.CustodialAssetsConfig.Symbol,
		ProjectSymbol:  projectSymbol,
		RefOrderID:     refOrderID,
		OperateSymbol:  operateSymbol,
		OperateAmount:  operatorAmount,
		SysFrom:        sysFrom,
		OtherInfo:      otherInfo,
	}
}

func (c *Client) newSignResultRequest(orderID string) *signResultRequest {
	return &signResultRequest{
		OrderID:       orderID,
		ProjectSymbol: projectSymbol,
	}
}

func (c *Client) postCustodialAssets(ctx context.Context, request *signRequest) error {
	if c == nil || !c.cfg.CustodialAssetsConfig.Enable {
		return errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(traceID, ctx.Value(traceID))

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshal request: %w", err)
	}

	reqSignURL, err := url.JoinPath(c.cfg.CustodialAssetsConfig.URL, requestSignURI)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	var signResp signResponse
	err = json.Unmarshal(body, &signResp)
	if err != nil {
		return fmt.Errorf("error unmarshal %v response body: %w", resp, err)
	}
	if signResp.Status != 200 {
		return fmt.Errorf("error response %v status: %v", signResp, signResp.Status)
	}

	return nil
}

func (c *Client) querySignResult(ctx context.Context, request *signResultRequest) (*signResponse, error) {
	if c == nil || !c.cfg.CustodialAssetsConfig.Enable {
		return nil, errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(traceID, ctx.Value(traceID))
	mLog.Infof("get sign result request: %v", request)

	querySignURL, err := url.JoinPath(c.cfg.CustodialAssetsConfig.URL, querySignURI)
	if err != nil {
		return nil, fmt.Errorf("error join url: %w", err)
	}
	response, err := http.Get(querySignURL)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
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

func (c *Client) waitResult(ctx context.Context, request *signResultRequest) error {
	return nil
}

func (c *Client) postSignRequestAndWaitResult(ctx context.Context, request *signRequest) error {
	if c == nil || !c.cfg.CustodialAssetsConfig.Enable {
		return errCustodialAssetsNotEnabled
	}
	mLog := log.WithFields(traceID, ctx.Value(traceID))
	mLog.Infof("post custodial assets request: %v", request)
	if err := c.postCustodialAssets(ctx, request); err != nil {
		return fmt.Errorf("error post custodial assets: %w", err)
	}
	mLog.Infof("post custodial assets success")
	if err := c.waitResult(ctx, c.newSignResultRequest(request.RefOrderID)); err != nil {
		return fmt.Errorf("error wait result: %w", err)
	}

	return nil
}
