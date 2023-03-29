package manager

import (
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/service"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	eos "github.com/qqtou/eos-go"
)

var ConfigPath string = "../config_eos.json"
var LogDir string = "../Log/"
var dbDir string = "./db"
var StartHeight uint64 = 15776234    // >=起始链提交的同步创世节点高度
var PolyStartHeight uint64 = 3960000 // >=Poly侧同步周期节点高度
var StartForceHeight uint64 = 15776234
var traceHeight uint32 = 18167351

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

	boltDB, err = db.NewBoltDB(dbDir)

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
	fmt.Printf("hdr.ChainId:%v\n", hdr.ChainID)
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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
	servConfig := getEOSServerConfig()
	eosSdk := getEOSServer()
	polySdk, _ := getPolyServer()
	boltDb, _ := getBoltDB()

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
	fmt.Printf("eosmamager new success \n")
	txActions, _, err := EOSManager.filterCrossChainEvent(uint32(eosCSheight), eosSdk)
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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

	EOSManager, err := NewEOSManager(&servConfig, StartHeight, StartForceHeight, polySdk, eosSdk, boltDb, service.NewRpcClient())
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

func TestArgs(t *testing.T) {
	var demoArgs = []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 100, 100, 99, 46, 99, 111, 110, 49, 1, 12, 100, 100, 99, 99, 99, 109, 97, 110, 97, 103, 101, 114, 8, 100, 100, 99, 46, 99, 111, 110, 50, 38, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 37, 104, 116, 116, 112, 115, 58, 47, 47, 103, 105, 116, 104, 117, 98, 46, 99, 111, 109, 49, 54, 55, 57, 51, 48, 51, 54, 51, 49, 54, 56, 51, 53, 56, 56, 52, 48, 48, 8, 116, 101, 115, 116, 100, 97, 116, 97}

	var strArgs = new(ArgsParam)

	strArgs.Deserialization(common.NewZeroCopySource(demoArgs))

	var strArgSer = new(ArgsParam)

	strArgSer.crossChainID = 1
	strArgSer.amount = 1
	strArgSer.ddcType = 1
	strArgSer.data = []byte("testdata")
	strArgSer.ddcSigner = []byte("ddcccmanager")
	// strArgSer. = []byte{105, 100, 220, 128, 71, 189, 213, 106, 75, 138, 204, 162, 128, 62, 107, 46, 2, 109, 49, 191}
	strArgSer.toOwner = []byte("ddc.con2")
	strArgSer.fromOwner = []byte("ddc.con1")
	strArgSer.ddcId = 37
	strArgSer.ddcURI = []byte("https://github.com1679303347626329500")

	fmt.Printf("strArgSer.fromOwner: %v\n", strArgSer.fromOwner)

	var newSink = common.NewZeroCopySink(nil)
	strArgSer.Serialization(newSink)

	fmt.Printf("newSink: %v\n", newSink)

	fmt.Printf("反序列化:\n")

	var otherArgs = new(ArgsParam)

	otherArgs.Deserialization(common.NewZeroCopySource(newSink.Bytes()))

	fmt.Printf("otherArgs: %v\n", otherArgs)

	fmt.Printf("otherArgs.crossChainID: %v\n", otherArgs.crossChainID)
	fmt.Printf("otherArgs.amount: %v\n", otherArgs.amount)
	fmt.Printf("otherArgs.ddcType: %v\n", otherArgs.ddcType)
	fmt.Printf("string(otherArgs.data): %v\n", string(otherArgs.data))
	fmt.Printf("string(otherArgs.ddcSigner): %v\n", string(otherArgs.ddcSigner))
	fmt.Printf("string(otherArgs.toOwner): %v\n", string(otherArgs.toOwner))
	fmt.Printf("string(otherArgs.fromOwner): %v\n", string(otherArgs.fromOwner))
	fmt.Printf("string(otherArgs.ddcURI): %v\n", string(otherArgs.ddcURI))
	fmt.Printf("otherArgs.ddcId: %v\n", otherArgs.ddcId)

}
