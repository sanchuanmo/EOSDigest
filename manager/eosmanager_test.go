package manager

import (
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"

	eos "github.com/qqtou/eos-go"
)

var ConfigPath string = "../config_eos.json"
var LogDir string = "../Log/"
var StartHeight uint64 = 15776234    // >=起始链提交的同步创世节点高度
var PolyStartHeight uint64 = 3960000 // >=Poly侧同步周期节点高度
var StartForceHeight uint64 = 15776234
var traceHeight uint32 = 18167351

// var eosUrl = "http://0.0.0.0:8888"
var eosCSheight = 8651667 // eos包含跨链事件的指定高度	暂未定，待测试完善

/*获取EOSServer*/
func getEOSServer() *eos.API {
	// read config
	servConfig := getEOSServerConfig()
	// 注册api
	chainApi := eos.New(servConfig.EOSConfig.RestURL)
	return chainApi
}

/*获取polyServer*/
func getPolyServer() (*sdk.PolySdk, error) {
	polySdk := sdk.NewPolySdk()
	servConfig := getEOSServerConfig()
	err := setUpPoly(polySdk, servConfig.PolyConfig.RestURL)
	if err != nil {
		log.Errorf("startServer - failed to setup Poly sdk %v", err)
	}
	return polySdk, err
}

/*获取EOSSerrverConfig*/
func getEOSServerConfig() config.ServiceEOSConfig {
	servConfig := config.NewServiceEOSConfig(ConfigPath)
	return *servConfig
}

/*获取BoltDB*/
func getBoltDB() (*db.BoltDB, error) {
	var boltDB *db.BoltDB
	var err error
	servConfig := getEOSServerConfig()
	if servConfig.BoltDbPath == "" {
		boltDB, err = db.NewBoltDB("boltdb")
	} else {
		boltDB, err = db.NewBoltDB(servConfig.BoltDbPath)
	}
	if err != nil {
		log.Fatalf("db.NewWaitingDB error:%s", err)
	}
	return boltDB, err
}

/*
设置poly
将chainID设置到poly
*/
func setUpPoly(poly *sdk.PolySdk, RpcAddr string) error {
	poly.NewRpcClient().SetAddress(RpcAddr)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	fmt.Printf("hdr.ChainId:%v", hdr.ChainID)
	return nil
}

/*
1、测试NewEOSManager
*/
func Test_NewEOSManager(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s", err)
	}

	fmt.Printf("eosmanager.eosClient is %s,forceHeight is%d", EOSManager.eosClient.BaseURL, EOSManager.forceHeight)
}

/*
获取EOS产出的最新不可逆块号：LastIrreversibleBlockNum
*/
func Test_GetEOSNodeHeight(t *testing.T) {
	eosSdk := getEOSServer()

	fmt.Printf("eosmanager new success")
	height, err := tools.GetEOSNodeHeight(eosSdk)
	if err != nil {
		fmt.Printf("eos get height error ,%s", err)
	} else {
		fmt.Printf("eos get height success, %d", height)
	}

}

/*
筛选EOS跨链事件: filterCrossChainEvent
*/
func Test_filterCrossChainEvent(t *testing.T) {
	eosSdk := getEOSServer()
	fmt.Printf("eosmamager new success \n")
	txActions, _, err := filterCrossChainEvent(uint32(eosCSheight), eosSdk)
	if err != nil {
		fmt.Printf("eos get crossChainEvent error,%s\n", err)
	} else {
		fmt.Printf("txActions %v\n", txActions)
	}
}

/*
测试fetchLockDepositEvents
从指定高度中获取事件并筛选跨链事件，打包Poly交易后存入DB数据库
eosCSheight	eos中包含跨链事件指定高度
*/
func Test_FetchLockDepositEvents(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s", err)
	}
	fmt.Printf("eosmamager new success \n")
	bool := EOSManager.fetchLockDepositEvents(uint64(eosCSheight), eosSdk)
	fmt.Printf("fetchLockDepositEvents : %v\n", bool)
	fmt.Print("Test_FetchLockDepositEvents end ")
}

/*
测试CommitHeader
*/
func Test_CommitHeader(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Get New EOS Manager error,%s", err)
	}
	sign := EOSManager.handleBlockHeader(uint64(eosCSheight))
	if !sign {
		fmt.Printf("Get handleBlockHeader error")
	}
	EOSManager.commitHeader()
}

/*
测试rollBackToCommAncestor
提交同步头失败后的回滚
回滚触发条件：commitHeader()失败，且错误为其中之一:["get the parent block failed","missing required field"]
*/
func Test_RollBackToCommAncestor(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("New EOS Manager error,%s", err)
	}
	sign := EOSManager.handleBlockHeader(uint64(eosCSheight))
	if !sign {
		fmt.Printf("Get handleBlockHeader error")
	}
	// 会报错执行回滚RollBackToCommAncestor
	EOSManager.commitHeader()
}

/*
需先测试Test_FetchLockDepositEvents，将跨链事件存入DB中
测试checkLockDepositEvents
*/
func Test_CheckLockDepositEvents(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s", err)
	}
	fmt.Printf("eosmamager new success \n")
	EOSManager.checkLockDepositEvents()
	fmt.Print("Test_CheckLockDepositEvents end ")
}

func Test_MonitorDeposit(t *testing.T) {
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb)
	if err != nil {
		fmt.Printf("Test New EOS Manager error,%s", err)
	}
	fmt.Printf("eosmamager new success \n")
	EOSManager.MonitorDeposit()
	fmt.Print("Test_MonitorDeposit end ")
}

/*
测试GetBlockTraceByNum接口，ABIBinToJSON接口
测试ABIBinToJSON获取数据的数据类型，并测试如何转换
*/
func Test_GetTraceBlock(t *testing.T) {
	eosSdk := getEOSServer()
	res, err := tools.GetEOSTraceBlockByNum(eosSdk, traceHeight)
	if err != nil {
		panic(fmt.Errorf("GetEOSTraceBlockByNum err:%s", err))
	}
	trans := res.Transactions

	// var events []string

	for i, tran := range trans {
		actions := tran.Actions
		for j, action := range actions {
			if action.Action == "crosschaine" && action.Account == "ddcccmanager" {
				// events = append(events, action.Data.(string))
				fmt.Printf("the %d transaction %d action.Data: %v\n", i, j, action.Data.(string))
				transData, err := tools.GetEOSDeTraceData(eosSdk, action.Account, eos.Name(action.Action), action.Data.(string))
				if err != nil {
					panic(fmt.Errorf("GetEOSDeTraceData error:%s", err))
				}
				fmt.Printf("transData: %v\n", transData)
				fmt.Printf("transData[\"caller\"].(string): %v\n", transData["caller"].(string))
				fmt.Printf("transData[\"toContract\"].(string): %v\n", transData["toContract"].(string))
				fmt.Printf("transData[\"paramTxHash\"].([]interface{}): %v\n", transData["paramTxHash"].([]interface{}))
				fmt.Printf("transData[\"toChainId\"].(float64): %v\n", transData["toChainId"].(float64))
				fmt.Printf("transData[\"rawParam\"].([]interface{}): %v\n", transData["rawParam"].([]interface{}))
				txHashTransBefore := transData["paramTxHash"].([]interface{})
				txHash := tools.TransInterfacesToBytes(txHashTransBefore)
				fmt.Printf("txHash: %v\n", txHash)
				rawParamTransBefore := transData["rawParam"].([]interface{})
				rawParam := tools.TransInterfacesToBytes(rawParamTransBefore)
				fmt.Printf("rawParam: %v\n", rawParam)

			}
		}
	}

}

// func TestTargetContract(t *testing.T) {
// 	servConfig := getEOSServerConfig()
// 	toContract := 90
// 	if len(servConfig.TargetContracts) > 0 {
// 		fmt.Printf("len(servConfig.TargetContracts): %v\n", len(servConfig.TargetContracts))

// 	}
// }
