package api

import (
	"errors"
	"fmt"
	"metamaskServer/client"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var GASPRICE uint64 = 500000

// Server struct
type Server struct {
	//r   *fasthttprouter.Router
	cli       client.Client
	chainId   string
	networkId string
}

type params struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
}

type responseBody struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result"`
}

type responseErr struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Error   *ErrorBody  `json:"error"`
}

type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type Transaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"S"`
}

type responseTransaction struct {
	JsonRPC string       `json:"jsonrpc"`
	Id      interface{}  `json:"id"`
	Result  *Transaction `json:"result"`
}

type TransactionReceipt struct {
	BlockHash         common.Hash    `json:"blockHash"`
	BlockNumber       uint64         `json:"blockNumber"`
	ContractAddress   common.Address `json:"contractAddress"`
	CumulativeGasUsed uint64         `json:"cumulativeGasUsed"`
	From              common.Address `json:"from"`
	GasUsed           uint64         `json:"gasUsed"`
	Logs              []*types.Log   `json:"logs"` //[]*types.Log
	LogsBloom         types.Bloom    `json:"logsBloom"`
	Status            uint64         `json:"status"`
	To                common.Address `json:"to"`

	TransactionHash  common.Hash `json:"transactionHash"`
	TransactionIndex uint        `json:"transactionIndex"`

	Root common.Hash `json:"root"`
}

type responseReceipt struct {
	JsonRPC string              `json:"jsonrpc"`
	Id      interface{}         `json:"id"`
	Result  *TransactionReceipt `json:"result"`
}

type Block struct {
	Difficulty string         `json:"difficulty"`
	ExtraData  string         `json:"extraData"`
	GasLimit   string         `json:"gasLimit"`
	GasUsed    string         `json:"gasUsed"`
	Hash       string         `json:"hash"`
	LogsBloom  string         `json:"logsBloom"`
	Miner      common.Address `json:"miner"`
	MixHash    string         `json:"mixHash"`
	Nonce      string         `json:"nonce"`
	Number     string         `json:"number"`
	ParentHash string         `json:"parentHash"`
	TimeStamp  string         `json:"timestamp"`
}

type responseBlock struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  *Block      `json:"result"`
}

type reqGetLog struct {
	FromBlock string   `json:"fromBlock"`
	ToBlock   string   `json:"toBlock"`
	Address   string   `json:"address"`
	Topics    []string `json:"topics"`
	BlockHash string   `json:"blockhash"`
}

type resGetLogs struct {
	LogIndex         string   `json:"logIndex"`
	BlockNumber      string   `json:"blockNumber"`
	BlockHash        string   `json:"blockHash"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

var (
	PRI                       string = "3DqKDKADdhWgdvpNgx1JJpFBx8fkvxEKNFeTkihJ3Xgwue3qRjq6R5JcPTdvZ5PEPjC9JaRLgPhhGPAQkoCYreV"
	ETH_CHAINID               string = "eth_chainId"
	NET_VERSION               string = "net_version"
	ETH_SENDTRANSACTION       string = "eth_sendTransaction"
	ETH_CALL                  string = "eth_call"
	ETH_BLOCKNUMBER           string = "eth_blockNumber"
	ETH_GETBALANCE            string = "eth_getBalance"
	ETH_GETBLOCKBYHASH        string = "eth_getBlockByHash"
	ETH_GETBLOCKBYNUMBER      string = "eth_getBlockByNumber"
	ETH_GETTRANSACTIONBYHASH  string = "eth_getTransactionByHash"
	ETH_GASPRICE              string = "eth_gasPrice"
	EHT_GETCODE               string = "eth_getCode"
	ETH_GETTRANSACTIONCOUNT   string = "eth_getTransactionCount"
	ETH_ESTIMATEGAS           string = "eth_estimateGas"
	ETH_SENDRAWTRANSACTION    string = "eth_sendRawTransaction"
	ETH_GETTRANSACTIONRECEIPT string = "eth_getTransactionReceipt"
	ETH_GETLOGS               string = "eth_getLogs"
	ETH_GETSTORAGEAT          string = "eth_getStorageAt"
	ETH_SIGNTRANSACTION       string = "eth_signTransaction"

	WEB3_CLIENTVERSION string = "web3_clientVersion"
)

func getString(mp map[string]interface{}, k string) (string, error) {
	v, ok := mp[k]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", k))
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", errors.New(fmt.Sprintf("'%s' not string", k))
}

func getValue(mp map[string]interface{}, k string) (interface{}, error) {
	v, ok := mp[k]
	if !ok {
		return 0, errors.New(fmt.Sprintf("'%s' not exist", k))
	}
	//log.Printf("value type %T,value:%v\n", v, v)
	return v, nil
}

func getParam(mp map[string]interface{}) (string, error) {
	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}

	if para, ok := v.([]interface{}); ok {
		return para[0].(string), nil
	}
	return "", errors.New("get params failed.")
}
