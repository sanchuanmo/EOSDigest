package toolbox

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/log"
	sdk "github.com/polynetwork/poly-go-sdk"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	"github.com/qqtou/eos-go"
)

var ConfigPath string = "../config_eos.json"
var LogDir string = "../Log/"
var chainNum uint64 = 93
var chainName string = "testChain093"
var eosHeight uint32 = 21128173 // 91 最新高度
var Epoch uint32 = 60000

var polyHeight uint32 = 4140000 //同步Poly的共识区块高度，为60000的整数倍// 当前高度4087694 ,周期为4080000

/*
获取配置类
*/
func getEOSServerConfig() *config.ServiceEOSConfig {
	servConfig := config.NewServiceEOSConfig(ConfigPath)
	return servConfig
}

/*
获取EOS SDK
*/
func getEOSServer() *eos.API {
	// read config
	servConfig := getEOSServerConfig()
	// 注册api
	chainApi := eos.New(servConfig.EOSConfig.RestURL)
	return chainApi
}

/*
获取Poly SDK
*/
func getPolyServer() (*sdk.PolySdk, error) {
	polySdk := sdk.NewPolySdk()
	servConfig := getEOSServerConfig()
	err := setUpPoly(polySdk, servConfig.PolyConfig.RestURL)
	if err != nil {
		log.Errorf("startServer - failed to setup Poly sdk %v", err)
	}
	return polySdk, err
}

/*
设置Poly SDK 参数
*/
func setUpPoly(poly *sdk.PolySdk, RpcAddr string) error {
	poly.NewRpcClient().SetAddress(RpcAddr)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	return nil
}

/*
整合获取RegisterEOS对象
*/
func GetRegisterEOS() *RegisterEOS {
	serv := getEOSServerConfig()
	polySdk, _ := getPolyServer()
	eosSdk := getEOSServer()
	register := NewRegisterEOS(serv, polySdk, eosSdk)
	return register
}

/*
注册侧链(申请注册侧链与同意注册侧链)
*/
func TestRegisterSideChain(t *testing.T) {
	register := GetRegisterEOS()
	err := register.RegisterSideChain(register.config.EOSConfig.SideChainId, uint64(register.config.RoutineNum), chainNum, chainName, register.config.EOSConfig.ContractAddress)
	if err != nil {
		fmt.Printf("TestRegisterSideChain: RegisterSideChain RPC faild error: %v", err)
	}
}

/*
注销侧链（申请注销侧链与同意注销侧链）
*/
func TestQuitSideChain(t *testing.T) {
	register := GetRegisterEOS()
	err := register.QuitSideChain(register.config.EOSConfig.SideChainId)
	if err != nil {
		fmt.Println(err)
	}
}

/*
构造并提交同步Poly创世节点的交易到目标链跨链管理合约
*/
func TestSyncPolyHdrToEOS(t *testing.T) {
	register := GetRegisterEOS()
	err := register.SyncPolyHdrToEOS(polyHeight)
	if err != nil {
		fmt.Printf("TestSyncPolyHdrToEOS: SyncPolyHdrToEOS faild error: %v", err)
	}
}

/*
同步起始链创世区块头到Poly
*/
func TestSyncGenesisHeaderToPoly(t *testing.T) {
	register := GetRegisterEOS()
	err := register.SyncGenesisHeaderToPoly(register.config.EOSConfig.SideChainId, eosHeight, register.eosclient)
	if err != nil {
		fmt.Printf("TestSyncPolyHdrToEOS: SyncGenesisHeaderToPoly faild error: %v", err)
	}
}

/*
测试获取Poly GenesisHeader接口存内存表的值
*/
func TestGetGenesisHeadrStorage(t *testing.T) {

	register := GetRegisterEOS()
	// CURRENT_HEADER_HEIGHT
	contractAddress := autils.HeaderSyncContractAddress
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], register.config.EOSConfig.SideChainId)
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), sideChainIdBytes[:]...)
	result, err := register.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		panic("poly: GetStorage CURRENT_HEADER_HEIGHT error" + err.Error())
	}

	fmt.Printf("result  GENESIS_HEADER is:%v\n", result)

	chainHeight := binary.LittleEndian.Uint64(result)

	fmt.Printf("poly: GetStorage get current height is success height: %d\n", chainHeight)

	// GENESIS_HEADER
	key2 := append([]byte(scom.GENESIS_HEADER), sideChainIdBytes[:]...)
	result2, err := register.polySdk.GetStorage(contractAddress.ToHexString(), key2)

	fmt.Printf("result2  GENESIS_HEADER is:%v\n", result2)

	if err != nil {
		panic("poly: GetStorage GENESIS_HEADER error" + err.Error())
	}
	var blockHeader *eos.SignedBlockHeader
	err = eos.UnmarshalBinary(result2, &blockHeader)
	if err != nil {
		panic("json: GENESIS_HEADER unmarshal eos.SignedBlockHeader error" + err.Error())
	}

	fmt.Printf("poly: GetStorage GENESIS_HEADER is success header: %v\n", blockHeader)

	// MAIN_CHAIN	存BlockIDBytes
	key3 := append(append([]byte(scom.MAIN_CHAIN), sideChainIdBytes[:]...), autils.GetUint64Bytes(chainHeight)...)

	result3, err := register.polySdk.GetStorage(contractAddress.ToHexString(), key3)
	if err != nil && result3 != nil {
		panic("poly: GetStorage MAIN_CHAIN error" + err.Error())
	}
	fmt.Printf("result3  MAIN_CHAIN is:%v\n", result3)

	var blockID2 eos.Checksum256
	err = blockID2.UnmarshalJSON(result3)
	if err != nil {
		panic("json: unmarshal eos blockID error" + err.Error())
	}
	fmt.Printf("poly: GetStorage MAIN_CHAIN is success blockID: %v\n", blockID2)

	// HEADER_INDEX	//依据ID查header
	blockID, _ := blockHeader.BlockID()
	blockIDBytes, _ := blockID.MarshalJSON()
	fmt.Printf("blockIDBytes :%v", blockIDBytes)
	key4 := append(append([]byte(scom.HEADER_INDEX), sideChainIdBytes[:]...), blockIDBytes...)
	result4, err := register.polySdk.GetStorage(contractAddress.ToHexString(), key4)
	if err != nil {
		panic("poly: GetStorage HEADER_INDEX error" + err.Error())
	}

	fmt.Printf("result4  HEADER_INDEX is:%v\n", result4)

	var blockHeader2 *eos.SignedBlockHeader
	err = eos.UnmarshalBinary(result4, &blockHeader2)
	header2ID, _ := blockHeader2.BlockID()
	header2IDByte, _ := header2ID.MarshalJSON()

	fmt.Printf("poly:header_index store the headerID is:%v\n", header2ID)
	fmt.Printf("poly:header_index store the headerIDByte is: %v\n", header2IDByte)
	if err != nil {
		panic("json: HEADER_INDEX unmarshal eos.SignedBlockHeader error" + err.Error())
	}
	fmt.Printf("poly: GetStorage HEADER_INDEX is success header: %v\n", blockHeader2)

}

/*
测试更换TransactionMRoot后序列化是否支持
*/
func TestENDEcode(t *testing.T) {
	register := GetRegisterEOS()
	eosSdk := register.eosclient
	var ctx context.Context = context.Background()
	blockResp, err := eosSdk.GetBlockByNum(ctx, eosHeight)
	if err != nil {
		panic("get block header error")
	}
	hdr := blockResp.SignedBlockHeader

	fmt.Printf("原hdr: %v", hdr)
	hdrByte, err := eos.MarshalBinary(hdr)
	if err != nil {
		panic("from header to byte error" + err.Error())
	}

	// 构成序列化结果
	var newHdr *eos.SignedBlockHeader
	err = eos.UnmarshalBinary(hdrByte, &newHdr)
	fmt.Printf("后hdr: %v", newHdr)
	if err != nil {
		panic("from byte to header error" + err.Error())
	}
}

/*
获取Poly最新同步周期高度,和最新高度
*/
func TestGetPolyEpochHeight(t *testing.T) {
	register := GetRegisterEOS()
	polySdk := register.polySdk
	lastHeight, _ := polySdk.GetCurrentBlockHeight()
	epochHeight := uint32(lastHeight/Epoch) * Epoch

	fmt.Printf("current height is %d\nepochHeight: %v\ndiff is: %d", lastHeight, epochHeight, lastHeight-epochHeight)
}

/*
注册Relayer(申请注册与同意注册)【已弃用】

func TestRegisterRelayer(t *testing.T) {
	register := GetRegisterEOS()
	addresses := []string{}
	acc, _ := register.getAccount()
	addresses = append(addresses, acc.Address.ToBase58())
	// addresses = append(addresses, register.config.PolyConfig.EntranceContractAddress)
	id, err := register.RegisterRelayer(addresses)
	if err != nil {
		fmt.Printf("TestRegisterRelayer: RegisterRelayer RPC faild error: %v", err)
	}
	err = register.ApproveRegisterRelayer(id)
	if err != nil {
		fmt.Printf("TestRegisterRelayer: ApproveRegisterRelayer RPC faild error: %v", err)
	}
}
*/
