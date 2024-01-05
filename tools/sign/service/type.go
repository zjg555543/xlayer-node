package service

import "encoding/json"

const OperateTypeSeq = 1
const OperateTypeAgg = 2

const CodeSuccess = 0
const CodeFail = 1

type Request struct {
	OperateType    int             `json:"operateType"`
	OperateAddress string          `json:"operateAddress"`
	Symbol         int             `json:"symbol"`
	ProjectSymbol  int             `json:"projectSymbol"`
	RefOrderId     string          `json:"refOrderId"`
	OperateSymbol  int             `json:"operateSymbol"`
	OperateAmount  int             `json:"operateAmount"`
	SysFrom        int             `json:"sysFrom"`
	OtherInfo      json.RawMessage `json:"otherInfo"`
}

type SeqData struct {
	Batches            []Batch `json:"batches"`
	SignaturesAndAddrs string  `json:"signaturesAndAddrs"`
	L2Coinbase         string  `json:"l2Coinbase"`
}

type Batch struct {
	GlobalExitRoot     string `json:"globalExitRoot"`
	MinForcedTimestamp int64  `json:"minForcedTimestamp"`
	Timestamp          int64  `json:"timestamp"`
	Transactions       string `json:"transactions"`
	TransactionsHash   string `json:"transactionsHash"`
}

type AggData struct {
	NewLocalExitRoot string   `json:"newLocalExitRoot"`
	NewStateRoot     string   `json:"newStateRoot"`
	FinalNewBatch    uint64   `json:"finalNewBatch"`
	Proof            []string `json:"proof"`
	InitNumBatch     uint64   `json:"initNumBatch"`
	PendingStateNum  int      `json:"pendingStateNum"`
}

type Response struct {
	Code      int    `json:"code"`
	Data      string `json:"data"`
	DetailMsg string `json:"detailMsg"`
	Msg       string `json:"msg"`
	Status    int    `json:"status"`
	Success   bool   `json:"success"`
}
