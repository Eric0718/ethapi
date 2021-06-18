package api

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"metamaskServer/client"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/goinggo/mapstructure"
)

func NewServer(addr, chainId, networkId, ethTo string) *Server {
	return &Server{cli: client.New(addr, ethTo), chainId: chainId, networkId: networkId}
}

func (s *Server) HandRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	reqData := make(map[string]interface{})
	if err := json.Unmarshal(body, &reqData); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	method, err := getString(reqData, "method")
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	jsonrpc, err := getString(reqData, "jsonrpc")
	if err != nil {
		log.Println("getString error:", err)
		w.Write([]byte(err.Error()))
		return
	}
	id, err := getValue(reqData, "id")
	if err != nil {
		log.Println("getValue:", err)
		w.Write([]byte(err.Error()))
		return
	}

	log.Println("request body:", string(body))
	log.Printf("method:%v\n", method)
	log.Println("jsonrpc:", jsonrpc, "id:", id)

	switch method {
	case ETH_CHAINID:
		chainId := s.eth_chainId()
		resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: chainId})
		if err != nil {
			log.Println("eth_chainId Marshal error:", err)
			w.Write([]byte(err.Error()))
		} else {
			fmt.Println("eth_chainId success res>>>", chainId)
			w.Write(resp)
		}
	case NET_VERSION:
		networkId := s.net_version()
		resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: networkId})
		if err != nil {
			log.Println("net_version error:", err)
			w.Write([]byte(err.Error()))
		} else {
			log.Println("net_version success res>>>", networkId)
			w.Write(resp)
		}

	case ETH_SENDTRANSACTION:
		hs, err := s.eth_sendTransaction(reqData)
		if err != nil {
			log.Println("eth_sendTransaction error:", err)
			w.Write([]byte(err.Error()))
		} else {
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: hs})
			if err != nil {
				log.Println("eth_sendTransaction Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_sendTransaction success res>>>", hs)
				w.Write(resp)
			}
		}
	case ETH_CALL:
		ret, err := s.eth_call(reqData)
		if ret == "" && err != nil {
			log.Println("eth_call error:", err)
			w.Write([]byte(err.Error()))
		} else if err != nil {
			var RetErr ErrorBody
			RetErr.Code = -4677
			RetErr.Message = err.Error()
			if len(ret) > 0 {
				btret := common.Hex2Bytes(ret)
				lenth := binary.BigEndian.Uint32(btret[64:68])
				data := btret[68 : lenth+68]
				errMsg := string(data)
				RetErr.Message = RetErr.Message + ": " + errMsg
			}
			RetErr.Data = "0x" + ret

			resp, err := json.Marshal(responseErr{JsonRPC: jsonrpc, Id: id, Error: &RetErr})
			if err != nil {
				log.Println("eth_call Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_call success ret>>>", ret)
				w.Write(resp)
			}
		} else {
			res := "0x" + ret
			log.Println("eth_call success res>>>", res)
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: res})
			if err != nil {
				log.Println("eth_call Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_call success res>>>", res)
				w.Write(resp)
			}
		}
	case ETH_BLOCKNUMBER:
		num, err := s.eth_blockNumber()
		if err != nil {
			log.Println("eth_blockNumber error:", err)
			w.Write([]byte(err.Error()))
		} else {
			log.Println("eth_blockNumber =", num)
			resNum := fmt.Sprintf("%X", num)
			log.Println("resNum =", resNum)

			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: ("0x" + resNum)})
			if err != nil {
				log.Println("eth_blockNumber Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_blockNumber success res>>>", "0x"+resNum)
				w.Write(resp)
			}
		}
	case ETH_GETBALANCE:
		from, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			blc, err := s.eth_getBalance(from)
			if err != nil {
				if err.Error() == "rpc error: code = Unknown desc = NotExist" {
					resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: "0x0"})
					if err != nil {
						log.Println("eth_getBalance Marshal error:", err)
						w.Write([]byte(err.Error()))
					} else {
						log.Println("eth_getBalance success res>>>", from, "NotExist")
						w.Write(resp)
					}
					break
				}
				log.Println("eth_getBalance error:", err)
				w.Write([]byte(err.Error()))
			} else {
				//metamask's decimal is 18,kto is 11,we need do blc*Pow10(7).
				bigB := new(big.Int).SetUint64(blc)
				bl := bigB.Mul(bigB, big.NewInt(10000000))

				resBalance := fmt.Sprintf("%X", bl)

				resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: ("0x" + resBalance)})
				if err != nil {
					log.Println("eth_getBalance Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getBalance success res>>>", from, "0x"+resBalance)
					w.Write(resp)
				}
			}
		}
	case ETH_GETBLOCKBYHASH:
		hash, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			blk, err := s.eth_getBlockByHash(hash)
			if err != nil {
				log.Println("eth_getBlockByHash error:", err)
				w.Write([]byte(err.Error()))
			} else {
				resp, err := json.Marshal(responseBlock{JsonRPC: jsonrpc, Id: id, Result: blk})
				if err != nil {
					log.Println("eth_getBlockByHash Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getBlockByHash success res>>>", blk.Hash)
					w.Write(resp)
				}
			}
		}
	case ETH_GETBLOCKBYNUMBER:
		strNum, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			num, err := strconv.ParseUint(strNum[2:], 16, 64)
			if err != nil {
				log.Println("ParseUint error:", err)
				w.Write([]byte(err.Error()))
			}

			blk, err := s.eth_getBlockByNumber(num)
			if err != nil {
				log.Println("eth_getBlockByNumber error:", err)
				w.Write([]byte(err.Error()))
			} else {
				resp, err := json.Marshal(responseBlock{JsonRPC: jsonrpc, Id: id, Result: blk})
				if err != nil {
					log.Println("eth_getBlockByNumber Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getBlockByNumber success res>>>", num)
					w.Write(resp)
				}
			}
		}
	case ETH_GETTRANSACTIONBYHASH:
		hash, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			tx, err := s.eth_getTransactionByHash(hash)
			if err != nil {
				log.Println("eth_getTransactionByHash error:", err)
				w.Write([]byte(err.Error()))
			} else {
				resp, err := json.Marshal(responseTransaction{JsonRPC: jsonrpc, Id: id, Result: tx})
				if err != nil {
					log.Println("eth_gasPrice Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Printf("eth_getTransactionByHash formart >>>>>>>>>>>>>>>>>: %s\n", string(resp))
					log.Println("eth_getTransactionByHash success res>>>", tx.Hash)
					w.Write(resp)
				}
			}
		}
	case ETH_GASPRICE:
		pric, err := s.eth_gasPrice()
		if err != nil {
			log.Println("eth_gasPrice error:", err)
			w.Write([]byte(err.Error()))
		} else {
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: pric})
			if err != nil {
				log.Println("eth_gasPrice Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_gasPrice success res>>>", pric)
				w.Write(resp)
			}
		}
	case EHT_GETCODE:
		addr, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			code, err := s.eth_getCode(addr)
			if err != nil {
				log.Println("eth_getCode error:", err)
				w.Write([]byte(err.Error()))
			} else {
				resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: code})
				if err != nil {
					log.Println("eth_getBalance Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getCode success res>>>", code)
					w.Write(resp)
				}
			}
		}
	case ETH_GETTRANSACTIONCOUNT:
		addr, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			count, err := s.eth_getTransactionCount(addr)
			if err != nil {
				log.Println("eth_getTransactionCount error:", err)
				w.Write([]byte(err.Error()))
			} else {
				hexCount := fmt.Sprintf("%X", count)
				resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: ("0x" + hexCount)})
				if err != nil {
					log.Println("eth_getTransactionCount Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getTransactionCount success res>>>", "addr:", addr, "nonce", count)
					w.Write(resp)
				}
			}
		}
	case ETH_ESTIMATEGAS:
		ret, err := s.eth_estimateGas(reqData)
		if ret == "" && err != nil {
			log.Println("eth_estimateGas error:", err)
			w.Write([]byte(err.Error()))
		} else if err != nil {
			var RetErr ErrorBody
			RetErr.Code = -4677
			RetErr.Message = err.Error()
			if len(ret) > 0 {
				btret := common.Hex2Bytes(ret)
				lenth := binary.BigEndian.Uint32(btret[64:68])
				data := btret[68 : lenth+68]
				errMsg := string(data)
				RetErr.Message = RetErr.Message + ": " + errMsg
			}
			RetErr.Data = "0x" + ret

			resp, err := json.Marshal(responseErr{JsonRPC: jsonrpc, Id: id, Error: &RetErr})
			if err != nil {
				log.Println("eth_estimateGas Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_estimateGas success ret>>>", ret)
				w.Write(resp)
			}
		} else {
			res := "0x" + ret
			log.Println("eth_estimateGas success res>>>", res)
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: res})
			if err != nil {
				log.Println("eth_estimateGas Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_estimateGas success res>>>", res)
				w.Write(resp)
			}
		}
	case ETH_SENDRAWTRANSACTION:
		rawTx, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			hash, err := s.eth_sendRawTransaction(rawTx)
			if err != nil {
				log.Println("eth_sendRawTransaction error:", err)
				w.Write([]byte(err.Error()))
			} else {
				resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: hash})
				if err != nil {
					log.Println("eth_sendRawTransaction Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_sendRawTransaction success res>>>", hash)
					w.Write(resp)
				}
			}
		}
	case ETH_GETTRANSACTIONRECEIPT:
		hash, err := getParam(reqData)
		if err != nil {
			log.Println("getParam error:", err)
			w.Write([]byte(err.Error()))
		} else {
			tc, err := s.eth_getTransactionReceipt(hash)
			if err != nil {
				log.Println("eth_getTransactionReceipt error:", err)
				w.Write([]byte(err.Error()))
			} else {

				resp, err := json.Marshal(responseReceipt{JsonRPC: jsonrpc, Id: id, Result: tc})
				if err != nil {
					log.Println("eth_getTransactionReceipt Marshal error:", err)
					w.Write([]byte(err.Error()))
				} else {
					log.Println("eth_getTransactionReceipt success res>>>", tc.TransactionHash, "contractAddr:", tc.ContractAddress, "status:", tc.Status, "blockNum:", tc.BlockNumber, "blockHash:", tc.BlockHash)
					for i, lg := range tc.Logs {
						log.Printf("tc.Logs>>>>>>>>>>>>>>[%v]:addr: %v,data: %v,topics: %v", i, lg.Address, hex.EncodeToString(lg.Data), lg.Topics)
					}

					log.Printf("formart>>>>>>>>>>>>>>>>>: %s\n", string(resp))
					w.Write(resp)
				}
			}
		}

	case ETH_GETLOGS:
		res, err := s.eth_getLogs(reqData)
		if err != nil {
			log.Println("eth_getLogs error:", err)
			w.Write([]byte(err.Error()))
		} else {
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: res})
			if err != nil {
				log.Println("eth_getLogs Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_getLogs success res>>>", res)
				w.Write(resp)
			}
		}
	case WEB3_CLIENTVERSION:
		res := s.web3_clientVersion()

		resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: res})
		if err != nil {
			log.Println("web3_clientVersion Marshal error:", err)
			w.Write([]byte(err.Error()))
		} else {
			fmt.Println("web3_clientVersion success res>>>", res)
			w.Write(resp)
		}

	case ETH_GETSTORAGEAT:
		res, err := s.eth_getStorageAt(reqData)
		//res := "0x00000000000000000000000000000000000000000000000000000000000004d2"
		if err != nil {
			log.Println("eth_getStorageAt error:", err)
			w.Write([]byte(err.Error()))
		} else {
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: res})
			if err != nil {
				log.Println("eth_getStorageAt Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				fmt.Println("eth_getStorageAt success res>>>", res)
				w.Write(resp)
			}
		}
	case ETH_SIGNTRANSACTION:
		signatrue, err := s.eth_signTransaction(reqData)
		if err != nil {
			log.Println("eth_signTransaction error:", err)
			w.Write([]byte(err.Error()))
		} else {
			resp, err := json.Marshal(responseBody{JsonRPC: jsonrpc, Id: id, Result: signatrue})
			if err != nil {
				log.Println("eth_signTransaction Marshal error:", err)
				w.Write([]byte(err.Error()))
			} else {
				log.Println("eth_signTransaction success res>>>", signatrue)
				w.Write(resp)
			}
		}

	default:
		log.Printf("Error unsupport method:%v\n", method)
		w.Write([]byte(fmt.Errorf("Unsupport method:%v", method).Error()))
	}
	return
}

func (s *Server) eth_chainId() string {
	return s.chainId
}

func (s *Server) net_version() string {
	return s.networkId
}

func (s *Server) eth_signTransaction(mp map[string]interface{}) (string, error) {

	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}
	if _, ok := v.([]interface{}); !ok {
		return "", errors.New("eth_signTransaction: params is wrong!")
	}

	Para := v.([]interface{})
	var para params

	err := mapstructure.Decode(Para[0].(map[string]interface{}), &para)
	if err != nil {
		return "", err
	}
	log.Printf("eth_signTransaction params: from=%v,to=%v,gas=%v,gasPrice=%v,value=%v,data=%v\n", para.From, para.To, para.Gas, para.GasPrice, para.Value, para.Data)
	return "no private key", nil
	// return "res", nil
}

func (s *Server) eth_sendTransaction(mp map[string]interface{}) (string, error) {
	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}
	if _, ok := v.([]interface{}); !ok {
		return "", errors.New("eth_sendTransaction: params is wrong!")
	}

	Para := v.([]interface{})
	var para params

	err := mapstructure.Decode(Para[0].(map[string]interface{}), &para)
	if err != nil {
		return "", err
	}
	log.Printf("eth_sendTransaction params: from=%v,to=%v,gas=%v,gasPrice=%v,value=%v,data=%v\n", para.From, para.To, para.Gas, para.GasPrice, para.Value, para.Data)

	//send common transaction
	vl, err := strconv.ParseUint(para.Value[2:], 16, 64)
	if err != nil {
		fmt.Println("ParseUint error:", err)
		return "", err
	}

	hash, err := s.cli.SendTransaction(para.From, para.To, PRI, vl)
	if err != nil {
		return "", err
	}
	return hash, nil
}

//send signed transaction
func (s *Server) eth_sendRawTransaction(rawTx string) (string, error) {
	log.Println("eth_sendRawTransaction rawTx=", rawTx)
	return s.cli.SendRawTransaction(rawTx)
}

//Executes a new message call immediately without creating a transaction on the block chain.
func (s *Server) eth_call(mp map[string]interface{}) (string, error) {
	log.Println("eth_call:", mp)
	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}

	if _, ok := v.([]interface{}); !ok {
		return "", errors.New("eth_call: params is wrong!")
	}

	Para := v.([]interface{})
	var para params

	err := mapstructure.Decode(Para[0].(map[string]interface{}), &para)
	if err != nil {
		return "", err
	}
	log.Printf("eth_call params: from=%v,to=%v,gas=%v,gasPrice=%v,value=%v,data=%v\n", para.From, para.To, para.Gas, para.GasPrice, para.Value, para.Data)

	ret, err := s.cli.ContractCall(para.From, para.To, para.Data) //para.From, para.To, PRI, para.Value, "call")
	if ret == "" {
		return "", err
	}
	return ret, err
}

func (s *Server) eth_blockNumber() (uint64, error) {
	return s.cli.GetBlockNumber()
}

func (s *Server) eth_getBalance(from string) (uint64, error) {
	log.Println("GetBalance from=", from)
	return s.cli.GetBalance(from)
}

func (s *Server) eth_getBlockByHash(hash string) (*Block, error) {
	log.Println("GetBlockBy Hash=", hash)
	b, err := s.cli.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	var block Block
	block.Hash = hex.EncodeToString(b.Hash)
	block.Miner = common.Address{}
	block.Number = "0x" + fmt.Sprintf("%X", b.Height)
	block.ParentHash = hex.EncodeToString(b.PrevHash)
	block.TimeStamp = "0x" + fmt.Sprintf("%X", b.Timestamp)

	return &block, nil
}

func (s *Server) eth_getBlockByNumber(num uint64) (*Block, error) {
	log.Println("GetBlockByNumber=", num)
	b, err := s.cli.GetBlockByNumber(num)
	if err != nil {
		return nil, err
	}
	var block Block
	block.Hash = hex.EncodeToString(b.Hash)
	block.Miner = common.Address{}
	block.Number = "0x" + fmt.Sprintf("%X", b.Height)
	block.ParentHash = hex.EncodeToString(b.PrevHash)
	block.TimeStamp = "0x" + fmt.Sprintf("%X", b.Timestamp)

	return &block, nil
}

func (s *Server) eth_getTransactionByHash(hash string) (*Transaction, error) {
	log.Println("GetTransactionByHash =", hash)
	if len(hash) > 0 {
		if hash[:2] == "0x" {
			hash = hash[2:]
		}
	}

	tx, err := s.cli.GetTransactionByHash(hash)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	b, err := s.cli.GetBlockByNumber(tx.BlockNumber)
	if err != nil {
		log.Println("GetBlockByNumber error==========:", err)
		return nil, errors.New(err.Error())
	}

	var trs Transaction

	trs.BlockHash = hex.EncodeToString(b.Hash)
	trs.BlockNumber = "0x" + fmt.Sprintf("%X", b.Height)
	trs.From = tx.EthFrom.Hex()
	trs.Gas = "0x" + fmt.Sprintf("%X", GASPRICE)
	trs.Hash = hex.EncodeToString(tx.Hash)
	trs.To = tx.EthTo.Hex()

	n, err := s.eth_getTransactionCount(trs.From)
	if err == nil {
		trs.Nonce = "0x" + fmt.Sprintf("%X", n)
	}

	if tx.EvmC != nil {
		if tx.EvmC.Operation == "create" || tx.EvmC.Operation == "Create" {
			trs.To = ""
		}
	}
	return &trs, nil
}

func (s *Server) eth_getCode(addr string) (string, error) {
	log.Println("GetCode=", addr)
	return s.cli.GetCode(addr)
}

func (s *Server) eth_getTransactionCount(addr string) (uint64, error) {
	log.Println("eth_getTransactionCount addr=", addr)
	return s.cli.GetNonce(addr)
}

func (s *Server) eth_gasPrice() (string, error) {
	return "0x" + fmt.Sprintf("%X", 21000), nil
}

func (s *Server) eth_estimateGas(mp map[string]interface{}) (string, error) {
	log.Println("eth_estimateGas:", mp)
	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}

	if _, ok := v.([]interface{}); !ok {
		return "", errors.New("eth_estimateGas: params is wrong!")
	}

	Para := v.([]interface{})
	var para params

	err := mapstructure.Decode(Para[0].(map[string]interface{}), &para)
	if err != nil {
		return "", err
	}
	log.Printf("eth_estimateGas params: from=%v,to=%v,gas=%v,gasPrice=%v,value=%v,data=%v\n", para.From, para.To, para.Gas, para.GasPrice, para.Value, para.Data)

	if len(para.To) <= 0 {
		return fmt.Sprintf("%X", GASPRICE), nil
	}

	ret, err := s.cli.ContractCall(para.From, para.To, para.Data) //para.From, para.To, PRI, para.Value, "call")

	if err == nil {
		return fmt.Sprintf("%X", GASPRICE), nil
	}

	log.Println("eth_estimateGas successfully,ret:", ret)
	return ret, err
}

func (s *Server) eth_getTransactionReceipt(hash string) (*TransactionReceipt, error) {
	if len(hash) > 0 {
		if hash[:2] == "0x" {
			hash = hash[2:]
		}
	}
	log.Println("eth_getTransactionReceipt hash=", hash)
	tx, err := s.cli.GetTransactionByHash(hash)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	b, err := s.cli.GetBlockByNumber(tx.BlockNumber)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	var trp TransactionReceipt
	trp.TransactionHash = common.BytesToHash(tx.Hash)
	trp.BlockNumber = tx.BlockNumber
	trp.GasUsed = GASPRICE
	trp.CumulativeGasUsed = trp.GasUsed
	trp.BlockHash = common.BytesToHash(b.Hash)
	trp.From = tx.EthFrom
	trp.To = tx.EthTo

	var txIndex uint
	if tx.EvmC != nil { //contract tx
		trp.ContractAddress = tx.EvmC.ContractAddr
		if tx.EvmC.Operation == "create" || tx.EvmC.Operation == "Create" {
			trp.To = common.Address{}
		}

		if !tx.EvmC.Status {
			trp.Status = 0
		} else {
			trp.Status = 1
		}

		log.Println("eth_getTransactionReceipt GetLogs hash:", hex.EncodeToString(tx.Hash), "trp.Root:", trp.Root)

		logs, err := s.cli.GetLogs(hex.EncodeToString(tx.Hash))
		if err != nil {
			log.Println("GetLogs error:", err)
		}
		for i, lo := range logs {
			var lg types.Log
			err := json.Unmarshal([]byte(lo), &lg)
			if err != nil {
				log.Println("eth_getTransactionReceipt Unmarshal error:", err)
				continue
			}

			lg.BlockNumber = tx.BlockNumber
			trp.Logs = append(trp.Logs, &lg)
			trp.TransactionIndex = lg.TxIndex
			txIndex = lg.TxIndex
			log.Printf("GetLogs[%v]:addr: %v,data: %v,topics: %v,trp.Logs length:%v\n", i, lg.Address, hex.EncodeToString(lg.Data), lg.Topics, len(trp.Logs))
		}

		//set bloom
		receipt := &types.Receipt{
			Type:              uint8(tx.Tag),
			Status:            trp.Status,
			Logs:              trp.Logs,
			TxHash:            common.BytesToHash(tx.Hash),
			ContractAddress:   trp.ContractAddress,
			GasUsed:           trp.GasUsed,
			BlockHash:         common.BytesToHash(b.Hash),
			TransactionIndex:  txIndex,
			BlockNumber:       new(big.Int).SetUint64(tx.BlockNumber),
			PostState:         trp.Root.Bytes(),
			CumulativeGasUsed: trp.GasUsed,
		}
		receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

		trp.LogsBloom = receipt.Bloom //types.Bloom{} //or receipt.Bloom,if use bloom it will filter corresponding log
	} else { //kto tx
		trp.Status = 1
	}

	log.Println("success to eth_getTransactionReceipt txhash,contractAddr,blockhash,blocknumber:", trp.TransactionHash, trp.ContractAddress, trp.BlockHash, trp.BlockNumber)
	return &trp, nil
}

func (s *Server) eth_getLogs(mp map[string]interface{}) ([]*types.Log, error) {
	log.Println("eth_getLogs:", mp)
	v, ok := mp["params"]
	if !ok {
		return nil, errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}

	if _, ok := v.([]interface{}); !ok {
		return nil, errors.New("eth_call: params is wrong!")
	}

	Para := v.([]interface{})
	var para reqGetLog

	err := mapstructure.Decode(Para[0].(map[string]interface{}), &para)
	if err != nil {
		return nil, err
	}
	log.Printf("eth_getLogs params: blockHash = %v,fromBlock=%v,toBlock=%v,address=%v,topics=%v\n", para.BlockHash, para.FromBlock, para.ToBlock, para.Address, para.Topics)

	var resLogs []*types.Log
	var fromBlock, toBlock uint64

	if len(para.FromBlock) > 0 && len(para.ToBlock) > 0 {
		fb, err := strconv.ParseUint(para.FromBlock[2:], 16, 64)
		if err != nil {
			fmt.Println("fromblock ParseUint error:", err)
			return nil, err
		}

		fromBlock = fb

		tb, err := strconv.ParseUint(para.ToBlock[2:], 16, 64)
		if err != nil {
			fmt.Println("toblock ParseUint error:", err)
			return nil, err
		}
		toBlock = tb
	}

	logs, err := s.cli.Logs(para.Address, fromBlock, toBlock, para.Topics, para.BlockHash)
	if err != nil {
		log.Println("GetLogs error:", err)
	}
	for i, lo := range logs {
		var lg types.Log
		err := json.Unmarshal([]byte(lo), &lg)
		if err != nil {
			log.Println("eth_getTransactionReceipt Unmarshal error:", err)
			continue
		}

		resLogs = append(resLogs, &lg)
		log.Printf("GetLogs[%v]:addr: %v,data: %v,topics: %v, txHash:%v\n", i, lg.Address, hex.EncodeToString(lg.Data), lg.Topics, lg.TxHash)
	}

	// var reslog resGetLogs
	// if para.FromBlock == para.ToBlock {
	// 	reslog.BlockNumber = para.FromBlock
	// }

	// if para.FromBlock != "" {
	// 	num, err := strconv.ParseUint(para.FromBlock[2:], 16, 64)
	// 	if err != nil {
	// 		fmt.Println("ParseUint error:", err)
	// 		return nil, err
	// 	}

	// 	b, err := s.cli.GetBlockByNumber(num)
	// 	if err != nil {
	// 		log.Println("GetBlockByNumber error==========:", err)
	// 		return nil, errors.New(err.Error())
	// 	}

	// 	reslog.BlockHash = hex.EncodeToString(b.Hash)
	// 	reslog.Address = para.Address
	// }

	log.Println("eth_getLogs successfully,reslog:", resLogs)
	return resLogs, nil
}

func (s *Server) web3_clientVersion() string {
	return "Mist/v0.9.3/darwin/go1.16"
}
func (s *Server) eth_getStorageAt(mp map[string]interface{}) (string, error) {
	log.Println("eth_getStorageAt:", mp)
	v, ok := mp["params"]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", "params"))
	}

	var addr, hash string
	if paras, ok := v.([]interface{}); ok {
		addr = paras[0].(string)
		hash = paras[1].(string)
	} else {
		return "", errors.New("eth_getStorageAt: params is wrong!")
	}
	return s.cli.GetStorageAt(addr, hash)
}
