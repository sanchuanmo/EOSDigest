package manager

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/contract"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/tools"
	"github.com/polynetwork/poly/common"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	eos "github.com/qqtou/eos-go"
)

var (
	polyEventHeight = 4351101
)

type EOSAccounts struct {
	chainId  string
	accounts []Account
}

type Account struct {
	accountResp *eos.AccountResp
	accountName string
	publicKey   string
	privateKey  string
}

func TestGetChainId(t *testing.T) {
	eosClient := getEOSServer()
	chainInfo, _ := eosClient.GetInfo(context.Background())
	chainId := chainInfo.ChainID.String()
	fmt.Printf("get the chainId is:%s\n", chainId)
}

func TestGetAccount(t *testing.T) {
	servConfig := getEOSServerConfig()
	storeAccounts := servConfig.EOSConfig.StoreAccounts
	for i, account := range storeAccounts {
		fmt.Printf("account %d info:\n accountName:%s\n privateKey:%s\n publicKey:%s\n", i, account["accountName"], account["privateKey"], account["publicKey"])
	}
}

func TestGetConfigAccount(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	chain_id := "demo"

	if len(servConfig.EOSConfig.StoreAccounts) == 0 {
		log.Fatal("relayer has no account")
		panic(fmt.Errorf("relayer has no account"))
	}

	service := &EOSAccounts{}
	service.chainId = chain_id
	for _, account := range servConfig.EOSConfig.StoreAccounts {
		fmt.Printf("account info:\n accountName:%s\n privateKey:%s\n publicKey:%s\n", account["accountName"], account["privateKey"], account["publicKey"])

		accountResp, err := eosSdk.GetAccount(context.Background(), eos.AccountName(account["accountName"]))
		if err != nil {
			log.Fatal("relayer config account info is error")
			panic(fmt.Errorf("relayer config account info is error"))
		}
		newAccount := &Account{
			accountResp: accountResp,
			accountName: account["accountName"],
			privateKey:  account["privateKey"],
		}

		service.accounts = append(service.accounts, *newAccount)
	}

	fmt.Printf("the struct EOSAccounts is %v\n", service)
}

func TestNewPolyManager(t *testing.T) {
	sigConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	polyManager, err := NewPolyManagerEOS(&sigConfig, 0, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Println("err", err)
	} else {
		fmt.Println(polyManager)
	}
}

func Test_SendTxToEOS(t *testing.T) {
	var info EOSTxInfo
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	PolyManagerEOS, err := NewPolyManagerEOS(&servConfig, 0, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s", err)
	}

	storeAccounts := servConfig.EOSConfig.StoreAccounts
	accountName := storeAccounts[0]["accountName"]
	prk := storeAccounts[0]["privateKey"]

	var input = new(contract.InputCrosschain)
	input.ToChainId = float64(88)
	input.ToContract = "test contract"
	input.Method = "method test"
	input.TxData = "txdata test"
	b, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("json.Marshal err:%v", err)
	}
	fmt.Printf("accountName:%v", accountName)
	info.basics = &contract.Basics{
		Caller:     eos.AccountName(accountName),
		Contract:   eos.AccountName("crosstest"),
		ActionName: eos.ActionName("crosschain"),
		Per:        "active",
	}
	// info.basics.Caller = eos.AccountName(accountName)
	info.prkey = prk
	// info.basics.Contract = eos.AccountName("crosstest")
	// info.basics.ActionName = eos.ActionName("crosschain")
	// info.basics.Per = "active"
	info.txData = b
	sender := PolyManagerEOS.selectSender()
	sender.sendTxToEOS(&info)
}

func TestGetPolyEvents(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	polyManager, err := NewPolyManagerEOS(&servConfig, 0, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s\n", err)
	}
	events, err := polyManager.polySdk.GetSmartContractEventByBlock(uint32(polyEventHeight))
	if err != nil {
		fmt.Printf("get events from poly error: %v\n", err)
	}
	for _, event := range events {
		for _, notify := range event.Notify {
			if notify.ContractAddress == polyManager.config.PolyConfig.EntranceContractAddress {

				states := notify.States.([]interface{})
				proof, err := polyManager.polySdk.GetCrossStatesProof(uint32(polyEventHeight), states[5].(string))
				if err != nil {
					log.Errorf("handleDepositEvents - failed to get proof for key %s: %v", states[5].(string), err)
					continue
				}
				log.Infof("目标链----获取Poly的跨链交易证明----") //ToDo
				auditpath, _ := hex.DecodeString(proof.AuditPath)
				value, _, _, _ := tools.ParseAuditpath(auditpath)
				param := &common2.ToMerkleValue{}
				if err := param.Deserialization(common.NewZeroCopySource(value)); err != nil {
					log.Errorf("handleDepositEvents - failed to deserialize MakeTxParam (value: %x, err: %v)", value, err)
					continue
				}

				//ToDo
				log.Infof("目标链----->跨链交易目标链地址(byte):%v\n,目标链地址(string):%v", param.MakeTxParam.ToContractAddress, string(param.MakeTxParam.ToContractAddress))
				log.Infof("目标链----->跨链交易目标链ID(uint64):%v\n", param.MakeTxParam.ToChainID)
				log.Infof("目标链----->跨链交易目标链方法(byte):%v\n", param.MakeTxParam.Method)
				log.Infof("目标链----->跨链交易目标链属性(byte):%v\n", param.MakeTxParam.Args)
				log.Infof("目标链----->跨链交易起始链地址(byte):%v\n", param.MakeTxParam.FromContractAddress)
				log.Infof("目标链----->跨链交易起始链ID(byte):%v\n", param.MakeTxParam.CrossChainID)
				//

			}
		}
	}

	fmt.Printf("events is :%v\n", events)
}
