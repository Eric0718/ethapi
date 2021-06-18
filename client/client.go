package client

import (
	"context"
	"errors"
	"fmt"
	kapi "kortho/api"
	"kortho/api/message"
	"kortho/bftconsensus/bftnode"
	"kortho/block"
	"kortho/transaction"
	"log"

	//"metamaskServer/api"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"google.golang.org/grpc"
)

type client struct {
	cli   message.GreeterClient
	ethTo string
}

func New(addr, ethT string) *client {
	log.Println("New client:", addr, ethT)
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil
	}
	return &client{
		cli:   message.NewGreeterClient(conn),
		ethTo: ethT,
	}
}

func (c *client) SendTransaction(from, to, priv string, amount uint64) (string, error) {
	resFrom, err := c.cli.GetKTOAddress(context.Background(), &message.ReqEthAddress{Ethaddress: from})
	if err != nil {
		return "", err
	}
	resTo, err := c.cli.GetKTOAddress(context.Background(), &message.ReqEthAddress{Ethaddress: to})
	if err != nil {
		return "", err
	}

	n, err := c.cli.GetAddressNonceAt(context.Background(), &message.ReqNonce{Address: resFrom.Ktoaddress})
	if err != nil {
		return "", err
	}
	var req message.ReqTransaction
	req.From = resFrom.Ktoaddress
	req.To = resTo.Ktoaddress
	req.Amount = amount
	req.Nonce = n.Nonce
	req.Priv = priv
	resp, err := c.cli.SendTransaction(context.Background(), &req)
	if err != nil {
		return "", err
	}
	h := resp.Hash
	{
		log.Printf("transaction hash: %s\n", h)
	}
	return h, nil
}

func (c *client) ContractCreate(createCode string, origin string) (string, error) { // , contractName string, from, to, priv string, amount uint64, option string) (string, error) {
	var req message.ReqContractTransaction
	req.Evm.CreateCode = createCode
	req.Evm.Origin = origin

	resp, err := c.cli.SendContractTransaction(context.Background(), &req)
	if err != nil {
		return "", err
	}
	return resp.Hash, nil
}

func (c *client) ContractCall(origin string, contractAddr string, callInput string) (string, error) { //, from, to, priv string, amount uint64, option string) (string, error) {
	var req message.ReqCallContract
	req.Contractaddress = contractAddr
	req.Inputcode = callInput
	req.Origin = origin

	resp, err := c.cli.CallSmartContract(context.Background(), &req)
	if err != nil {
		return "", err
	}

	if len(resp.Msg) > 0 {
		return resp.Result, errors.New(resp.Msg)
	}

	return resp.Result, nil
}

func (c *client) GetBlockNumber() (uint64, error) {
	num, err := c.cli.GetMaxBlockNumber(context.Background(), &message.ReqMaxBlockNumber{})
	if err != nil {
		return 0, err
	}

	return num.MaxNumber, nil
}

func (c *client) GetBalance(from string) (uint64, error) {
	res, err := c.cli.GetKTOAddress(context.Background(), &message.ReqEthAddress{Ethaddress: from})
	if err != nil {
		return 0, err
	}

	num, err := c.cli.GetBalance(context.Background(), &message.ReqBalance{Address: res.Ktoaddress})
	if err != nil {
		return 0, err
	}
	log.Println("GetBalance from KTOAddress", res.Ktoaddress, "balance=", num.Balnce)
	return num.Balnce, nil
}

func (c *client) GetBlockByHash(hash string) (*block.Block, error) {
	resp, err := c.cli.GetBlockByHash(context.Background(), &message.ReqBlockByHash{Hash: hash})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("GetBlockByHash error: hash=%v,Code = %v,%v", hash, resp.Code, resp.Message)
	}

	return bftnode.BlockConversion(resp.Data)
}

func (c *client) GetBlockByNumber(num uint64) (*block.Block, error) {
	resp, err := c.cli.GetBlockByNum(context.Background(), &message.ReqBlockByNumber{Height: num})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("GetBlockByNumber error: number=%v,Code = %v,%v", num, resp.Code, resp.Message)
	}

	return bftnode.BlockConversion(resp.Data)
}

func (c *client) GetTransactionByHash(hash string) (*transaction.Transaction, error) {
	resp, err := c.cli.GetTxByHash(context.Background(), &message.ReqTxByHash{Hash: hash})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("GetTransactionByHash error: hash=%v,Code = %v,%v", hash, resp.Code, resp.Message)
	}
	return kapi.MsgTxToTx(resp.Data)
}

func (c *client) GetCode(contractAddr string) (string, error) {
	resp, err := c.cli.GetCode(context.Background(), &message.ReqEvmGetcode{Addr: contractAddr})
	if err != nil {
		return "", err
	}
	return resp.Code, nil
}

func (c *client) GetNonce(addr string) (uint64, error) {
	res, err := c.cli.GetKTOAddress(context.Background(), &message.ReqEthAddress{Ethaddress: addr})
	if err != nil {
		return 0, err
	}

	log.Println("GetNonce 'from' KTOAddress=", res.Ktoaddress)
	resp, err := c.cli.GetAddressNonceAt(context.Background(), &message.ReqNonce{Address: res.Ktoaddress})
	if err != nil {
		return 0, err
	}
	return resp.Nonce, nil
}

func (c *client) SendRawTransaction(rawTx string) (string, error) {
	decTX, err := hexutil.Decode(rawTx)
	if err != nil {
		log.Fatal("hexutil Decode error:", err)
		return "", err
	}
	var tx types.Transaction
	err = rlp.DecodeBytes(decTX, &tx)
	if err != nil {
		log.Fatal("DecodeBytes error:", err)
		return "", err
	}

	signer := types.NewEIP155Signer(tx.ChainId())
	sender, err := signer.Sender(&tx)
	if err != nil {
		log.Fatal("signer.Sender error:", err)
	}
	log.Printf("tx params:{sender:%v,to:%v,amount:%v,nounce:%v,hash:%v,gas:%v,gasPrice:%v}\n", sender, tx.To(), tx.Value(), tx.Nonce(), tx.Hash(), tx.Gas(), tx.GasPrice())

	dl := len(tx.Data())
	log.Printf("data lenght====%v\n:", dl)

	//return "Ok", nil

	if dl > 0 { //合约调用,代币转账
		var req message.ReqEthSignContractTransaction
		req.EthFrom = sender.Hex()
		req.EthTo = c.ethTo
		req.EthData = rawTx

		resp, err := c.cli.SendEthSignedContractTransaction(context.Background(), &req)
		if err != nil {
			return "", fmt.Errorf("%v,sender=%v,contractAddr=%v", err, req.EthFrom, tx.To())
		}

		{
			log.Println("kto hash:", resp.Hash)
		}
		return resp.Hash, nil
	}

	//KTO 普通交易转账
	var req message.ReqEthSigntransaction
	req.EthFrom = sender.Hex()
	req.EthData = rawTx

	resp, err := c.cli.SendEthSignedTransaction(context.Background(), &req)
	if err != nil {
		return "", err
	}
	{
		log.Println("kto hash:", resp.Hash)
	}
	return resp.Hash, nil
}

func (c *client) GetTransactionReceipt(hash string) (*transaction.Transaction, error) {
	resp, err := c.cli.GetTxByHash(context.Background(), &message.ReqTxByHash{Hash: hash})
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("GetTransactionReceipt error: hash=%v,Code = %v,%v", hash, resp.Code, resp.Message)
	}
	tx := resp.Data
	if tx.Evm != nil {
		log.Println("tx.Evm:", tx.Evm.ContractAddr)
	}
	return kapi.MsgTxToTx(resp.Data)
}

func (c *client) GetLogs(hash string) ([]string, error) {
	resp, err := c.cli.GetEvmLogs(context.Background(), &message.ReqEvmGetlogs{Hash: hash})
	if err != nil {
		return nil, err
	}

	return resp.Evmlog, nil
}

func (c *client) GetStorageAt(addr, hash string) (string, error) {
	resp, err := c.cli.GetStorageAt(context.Background(), &message.ReqGetstorage{Addr: addr, Hash: hash})
	if err != nil {
		return "nil", err
	}

	return resp.Result, nil
}

func (c *client) Logs(address string, fromB, toB uint64, topics []string, blockH string) ([]string, error) {
	resp, err := c.cli.Logs(context.Background(), &message.ReqLogs{Address: address, FromBlock: fromB, ToBlock: toB, BlockHash: blockH})
	if err != nil {
		return nil, err
	}

	return resp.Evmlogs, nil
}
