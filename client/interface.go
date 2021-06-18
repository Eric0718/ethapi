package client

import (
	"kortho/block"
	"kortho/transaction"
)

type Client interface {
	SendTransaction(string, string, string, uint64) (string, error)
	//ContractCreate(createCode string, origin string, contractName string, from, to, priv string, amount uint64, option string) (string, error)
	ContractCreate(createCode string, origin string) (string, error)
	//ContractCall(origin string, contractAddr string, callInput string, from, to, priv string, amount uint64, option string) (string, error)
	ContractCall(origin string, contractAddr string, callInput string) (string, error)
	GetBlockNumber() (uint64, error)
	GetBalance(from string) (uint64, error)
	GetBlockByHash(hash string) (*block.Block, error)
	GetBlockByNumber(num uint64) (*block.Block, error)
	GetCode(contractAddr string) (string, error)
	GetNonce(addr string) (uint64, error)
	GetTransactionByHash(hash string) (*transaction.Transaction, error)
	SendRawTransaction(rawTx string) (string, error)
	GetTransactionReceipt(hash string) (*transaction.Transaction, error)
	GetLogs(hash string) ([]string, error)
	GetStorageAt(addr, hash string) (string, error)
	Logs(address string, fromB, toB uint64, topics []string, blockH string) ([]string, error)
}
