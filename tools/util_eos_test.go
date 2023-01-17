package tools

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/poly/common"
	ccmcommon "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	"github.com/qqtou/eos-go"
)

var (
	height      = 18561780
	chainHeight = 19047110 //19044074
)

func TestGetBookkeeper(t *testing.T) {
	eosSdk := getEOSServer()
	data, err := GetEOSRawKeepers(eosSdk, "ddcccmanager", "polyglobal")
	if err != nil {
		panic("get Raw Keepers error:" + err.Error())
	}
	fmt.Printf("bookkeepers is:%v\n", data)
}

func TestGetTableRowMap(t *testing.T) {
	eosSdk := getEOSServer()
	var request = eos.GetTableRowsRequest{
		JSON:    true,
		Code:    "ddcccmanager",
		Scope:   "ddcccmanager",
		Table:   "polyglobal",
		Reverse: false,
	}
	data, err := GetTableRowsMap(eosSdk, request)
	if err != nil {
		panic("get table Rows Map error" + err.Error())
	}
	fmt.Printf("data: %v\n", data)
	rawKeepers := data[0]
	fmt.Printf("rawKeepers: %v\n", rawKeepers)
	rawKeeper := rawKeepers["conKeepersPkBytes"]
	fmt.Printf("rawKeeper: %v\n", rawKeeper)
}

func TestHeightChange(t *testing.T) {
	eosSdk := getEOSServer()
	for i := 0; i < 10; i++ {
		fmt.Println()
		block, err := GetEOSBlockByNum(eosSdk, uint32(height+i))
		if err != nil {
			fmt.Printf("Get Block error : %s", err)
		}
		fmt.Printf("%d: block.BlockNum: %v\n", i, block.BlockNum)
	}
}

func TestGetPolySideChainProof(t *testing.T) {
	polySdk, _ := getPolyServer()
	servConfig := getEOSServerConfig()
	// 根据高度获取blockID
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], servConfig.EOSConfig.SideChainId)
	contractAddress := autils.HeaderSyncContractAddress

	key3 := append(append([]byte(scom.MAIN_CHAIN), sideChainIdBytes[:]...), autils.GetUint64Bytes(uint64(chainHeight))...)

	result3, err := polySdk.GetStorage(contractAddress.ToHexString(), key3)
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
	//依据ID查header
	blockIDBytes, _ := blockID2.MarshalJSON()
	key4 := append(append([]byte(scom.HEADER_INDEX), sideChainIdBytes[:]...), blockIDBytes...)
	result4, err := polySdk.GetStorage(contractAddress.ToHexString(), key4)
	if err != nil {
		panic("poly: GetStorage HEADER_INDEX error" + err.Error())
	}

	fmt.Printf("result4  HEADER_INDEX is:%v\n", result4)
	var blockHeader2 *eos.SignedBlockHeader
	err = eos.UnmarshalBinary(result4, &blockHeader2)
	if err != nil {
		panic("eos.UnmarshalBinary" + err.Error())
	}
	fmt.Printf("blockHeader2.TransactionMRoot: %v\n", blockHeader2.TransactionMRoot)
}

func TxDataDeserialization(source *common.ZeroCopySource) (*ccmcommon.MakeTxParam, error) {
	var data *ccmcommon.MakeTxParam
	txHash, eof := source.NextVarBytes()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize txHash error")
	}
	crossChainID, eof := source.NextVarBytes()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize crossChainID error")
	}
	fromContractAddress, eof := source.NextVarBytes()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize toChainID error")
	}
	toContractAddress, eof := source.NextVarBytes()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize toContractAddress error")
	}
	method, eof := source.NextString()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize method error")
	}
	args, eof := source.NextVarBytes()
	if eof {
		return nil, fmt.Errorf("MakeTxParam deserialize args error")
	}

	data.TxHash = txHash
	data.CrossChainID = crossChainID
	data.FromContractAddress = fromContractAddress
	data.ToChainID = toChainID
	data.ToContractAddress = toContractAddress
	data.Method = method
	data.Args = args
	return data, nil
}

func TestDeSerlizeTxData(t *testing.T) {
	eosSdk := getEOSServer()
	res, err := GetEOSTraceBlockByNum(eosSdk, uint32(height))
	if err != nil {
		log.Error("EOS filterCrossChainEvent - error: %s", err)
	}
	var resData []byte
	for i, transaction := range res.Transactions {
		for _, action := range transaction.Actions {
			if action.Action != "onblock" {
				log.Infof("---->the block height %d, transaction [%d] action is:%s account is:%s", height, i, action.Action, action.Account)
			}
			if action.Action == "crosschaine" && action.Account == "ddcccmanager" {

				resDatas, err := GetEOSDeTraceData(eosSdk, action.Account, eos.Name(action.Action), action.Data.(string))
				if err != nil {
					log.Error("EOS filterCrossChainEvent - error: %s", err)
				}
				resData = TransInterfacesToBytes(resDatas["rawParam"].([]interface{}))
			} else {
				continue
			}
		}
	}
	data := common.NewZeroCopySource(resData)
	var deTxData *ccmcommon.MakeTxParam
	deTxData, err = TxDataDeserialization(data)
	if err != nil {
		log.Error("deTxData Deserialization - error: %s", err)
	} else {
		log.Infof("deTxData Deserialization success. deTxData: %v", deTxData)
	}

}
