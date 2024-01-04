package ethtxmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

func (c *Client) postCustodialAssets(request *signRequest) error {
	if c == nil || !c.cfg.CustodialAssetsConfig.Enable {
		return errCustodialAssetsNotEnabled
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.cfg.CustodialAssetsConfig.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

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

func (c *Client) waitResult() error {
	return nil
}
