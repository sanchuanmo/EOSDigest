package manager

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/proof"
	"github.com/polynetwork/eos_relayer/service"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	eos "github.com/qqtou/eos-go"
)

// 起始链数据聚合服务

type RawParam struct {
	txHash              []byte // 起始链交易Hash
	crossChainID        []byte // 跨链ID
	fromContractAddress []byte //
	toChainID           uint64 // 目标链ID
	toContractAddress   []byte
	method              string
	args                []byte
}

// 添加监听到的跨链事件时间
// 添加发送成功后拿到的Poly的交易Hash
// 序列化 DDC跨链唯一标识符
// DDC id args中

func (raw *RawParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(raw.txHash)
	sink.WriteVarBytes(raw.crossChainID)
	sink.WriteVarBytes(raw.fromContractAddress)
	sink.WriteUint64(raw.toChainID)
	sink.WriteVarBytes(raw.toContractAddress)
	sink.WriteVarBytes([]byte(raw.method))
	sink.WriteVarBytes(raw.args)
}

func (raw *RawParam) Deserialization(source *common.ZeroCopySource) error {
	txHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("rawParam deserialize txHash error")
	}
	crossChainID, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("rawParam deserialize crossChainID error")
	}
	fromContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("rawParam deserialize fromContractAddress error")
	}
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("rawParam deserialize toChainID error")
	}
	toContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("rawParam deserialize toContractAddress error")
	}
	method, eof := source.NextString()
	if eof {
		return fmt.Errorf("rawParam deserialize method error")
	}
	args, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("rawParam deserialize args error")
	}

	raw.txHash = txHash
	raw.crossChainID = crossChainID
	raw.fromContractAddress = fromContractAddress
	raw.toChainID = toChainID
	raw.toContractAddress = toContractAddress
	raw.method = method
	raw.args = args
	return nil
}

type ArgsParam struct {
	crossChainID uint64
	fromOwner    []byte
	ddcType      uint64
	ddcSigner    []byte
	toOwner      []byte
	ddcId        uint64
	amount       uint64
	ddcURI       []byte
	data         []byte
}

func (argsP *ArgsParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(argsP.crossChainID)
	var temp = make([]byte, 24)
	sink.WriteBytes(temp)
	sink.WriteVarBytes(argsP.fromOwner)
	sink.WriteUint8(uint8(argsP.ddcType))
	sink.WriteVarBytes(argsP.ddcSigner)
	sink.WriteVarBytes(argsP.toOwner)
	sink.WriteUint64(argsP.ddcId)
	sink.WriteBytes(temp)
	sink.WriteUint64(argsP.amount)
	sink.WriteBytes(temp)
	sink.WriteVarBytes(argsP.ddcURI)
	sink.WriteVarBytes(argsP.data)

}

func (argsP *ArgsParam) Deserialization(source *common.ZeroCopySource) error {
	crossChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("argsParam deserialize crossChainID error")
	}
	source.Skip(24)
	// _, eof = source.NextBytes(24)
	// if eof {
	// return fmt.Errorf("argsParam deserialize crossChainID others bytes error")
	// }

	fromOwner, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("argsParam deserialize fromOwner error")
	}

	ddcType, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("argsParam deserialize ddcType error")
	}

	ddcSigner, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("argsParam deserialize ddcSigner error")
	}

	toOwner, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("argsParam deserialize toOwner error")
	}

	ddcId, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("argsParam deserialize ddcId error")
	}
	source.Skip(24)
	// _, eof = source.NextBytes(24)
	// if eof {
	// return fmt.Errorf("argsParam deserialize ddcId others bytes error")
	// }
	amount, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("argsParam deserialize amount error")
	}
	source.Skip(24)
	// _, eof = source.NextBytes(24)
	// if eof {
	// return fmt.Errorf("argsParam deserialize amount others bytes error")
	// }

	ddcURI, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("argsParam deserialize ddcURI error")
	}

	data, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("argsParam deserialize data error")
	}

	argsP.crossChainID = crossChainID
	argsP.fromOwner = fromOwner
	argsP.ddcType = uint64(ddcType)
	argsP.ddcSigner = ddcSigner
	argsP.toOwner = toOwner
	argsP.ddcId = ddcId
	argsP.ddcURI = ddcURI
	argsP.data = data
	argsP.amount = amount
	return nil
}

type TxActionData struct {
	toChainId  string // wangzelong 考虑转成string
	toContract string
	caller     string
	txHash     []byte
	feeData    FeeActionData
	rawParam   []byte
	txId       []byte
	leaf       []byte
}

type FeeActionData struct {
	ddcID        uint32 // ddc唯一标识
	account      string //调用账户
	businessType uint32 //合约类型
	funcName     string //收费action
	fee          string //收费价格
	balance      string // 账户剩余
}

type CrossTransfer struct {
	txIndex     string // 交易索引	源链event.TxHash().hexString()
	txId        []byte // 交易id	源链event.TxHash().Bytes()
	value       []byte // 值		源链event.RawData()
	toChain     string // 目标链	源链event.ToChainId
	height      uint64 // 高度		源链块高度
	fee         string // 费用		源链跨链费用
	caller      string // 调用人	源链调用人
	filterTime  string // 筛选到跨链事件时间
	merkleProof []byte // 默克尔证明
}

func (cross *CrossTransfer) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(cross.txIndex)
	sink.WriteVarBytes(cross.txId)
	sink.WriteVarBytes(cross.value)
	sink.WriteString(cross.toChain)
	sink.WriteUint64(cross.height)
	sink.WriteString(cross.fee)
	sink.WriteString(cross.caller)
	sink.WriteString(cross.filterTime)
	sink.WriteVarBytes(cross.merkleProof)

}

// 反序列化
func (cross *CrossTransfer) Deserialization(source *common.ZeroCopySource) error {
	txIndex, eof := source.NextString()
	if eof {
		return fmt.Errorf("waiting deserialize txIndex error")
	}
	txId, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("waiting deserialize txId error")
	}
	value, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("waiting deserialize value error")
	}
	toChain, eof := source.NextString()
	if eof {
		return fmt.Errorf("waiting deserialize toChain error")
	}
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("waiting deserialize height error")
	}
	fee, eof := source.NextString()
	if eof {
		return fmt.Errorf("waiting deserialize fee error")
	}
	caller, eof := source.NextString()
	if eof {
		return fmt.Errorf("waiting deserialize caller error")
	}
	filterTime, eof := source.NextString()
	if eof {
		return fmt.Errorf("waiting deserialize filterTime error")
	}
	merkleProof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("waiting deserialize merkleProof error")
	}
	cross.txIndex = txIndex
	cross.txId = txId
	cross.value = value
	cross.toChain = toChain
	cross.height = height
	cross.fee = fee
	cross.caller = caller
	cross.filterTime = filterTime
	cross.merkleProof = merkleProof
	return nil
}

type EOSManager struct {
	config *config.ServiceEOSConfig //service配置

	eosClient     *eos.API
	currentHeight uint64             // 当前高度
	forceHeight   uint64             // force高度
	preBlockID    *eos.Checksum256   // 父节点ID Proof相关
	polySdk       *sdk.PolySdk       // polySDK
	polySigner    *sdk.Account       // poly注册器
	exitChan      chan int           // exit chan
	header4sync   [][]byte           // 头同步
	crosstx4sync  []*CrossTransfer   // 跨链交易同步
	db            *db.BoltDB         // blotDB
	serviceClient *service.RpcClient // 聚合服务Client
	count         uint64             // ToDo 失败次数
}

func NewEOSManager(servConfig *config.ServiceEOSConfig, startheight uint64, startforceheight uint64, polySdk *sdk.PolySdk, eosSdk *eos.API, boltDB *db.BoltDB, serviceClient *service.RpcClient) (*EOSManager, error) {

	var wallet *sdk.Wallet
	var err error
	var signer *sdk.Account

	if !common.FileExisted(servConfig.PolyConfig.WalletFile) {
		wallet, err = polySdk.CreateWallet(servConfig.PolyConfig.WalletFile)
		if err != nil {
			log.Errorf("EOS NewEOSManager - wallet create error: %s", err.Error())
		}
	} else {
		wallet, err = polySdk.OpenWallet(servConfig.PolyConfig.WalletFile)
		if err != nil {
			log.Errorf("EOS NewEOSManager - wallet open error: %s", err)
		}
	}
	signer, err = wallet.GetDefaultAccount([]byte(servConfig.PolyConfig.WalletPwd))
	if err != nil || signer == nil {
		signer, err = wallet.NewDefaultSettingAccount([]byte(servConfig.PolyConfig.WalletPwd))
		if err != nil {
			log.Errorf("EOS NewEOSManager - wallet password error")
		}

		err = wallet.Save()
		if err != nil {
			log.Errorf("EOS NewEOSManager - wallet save account error")
		}
	}

	if err != nil {
		log.Error("EOS NewEOSManager - wallet get default account error")
	}
	log.Infof("EOS NewEOSManager - poly address: %s", signer.Address.ToBase58())

	mgr := &EOSManager{
		config:        servConfig,
		exitChan:      make(chan int),
		currentHeight: startheight,
		forceHeight:   startforceheight,
		eosClient:     eosSdk,
		polySdk:       polySdk,
		polySigner:    signer,
		header4sync:   make([][]byte, 0),
		crosstx4sync:  make([]*CrossTransfer, 0),
		db:            boltDB,
		serviceClient: serviceClient,
	}

	err = mgr.init()
	if err != nil {
		return nil, err
	} else {
		return mgr, nil
	}
}

func (eosmanager *EOSManager) init() error {
	// get latestheight
	latestHeight := eosmanager.findLastestHeight()
	if latestHeight == 0 {
		return fmt.Errorf("EOS init - the genesis block has not synced")
	}
	if eosmanager.forceHeight > 0 && eosmanager.forceHeight < latestHeight {
		eosmanager.currentHeight = eosmanager.forceHeight
	} else {
		eosmanager.currentHeight = latestHeight
	}
	log.Infof("EOS init - findLastestHeight success, LastestHeight:%v", latestHeight)
	log.Infof("EOS init - start height: %d\n", eosmanager.currentHeight)
	//Proof额外代码
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], eosmanager.config.EOSConfig.SideChainId)
	preBlockID, err := tools.GetPolyStorageHeaderID(eosmanager.polySdk, latestHeight, sideChainIdBytes)
	if err != nil {
		return fmt.Errorf("EOS init - get preBlockID error:%s", err)
	}

	eosmanager.preBlockID = preBlockID
	// end
	return nil
}

// 查询Poly内存表CURRENT_HEADER_HEIGHT 存储EOS侧链最新高度
// 查询Poly内存表MAIN_CHAIN 存储EOS侧链最新高度 对应的BlockID
func (eosmanager *EOSManager) findLastestHeight() uint64 {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], eosmanager.config.EOSConfig.SideChainId)

	contractAddress := autils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), sideChainIdBytes[:]...)
	// try to get storage
	result, err := eosmanager.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil && result != nil {
		log.Errorf("findLastestHeight - GetStorage CURRENT_HEADER_HEIGHT error: %s", err)
	}

	height := binary.LittleEndian.Uint64(result)

	if err != nil {
		return 0
	}
	if result == nil || len(result) == 0 {
		return 0
	} else {
		return height
	}
}

func (eosmanager *EOSManager) MonitorEOSChain() {
	fetchBlockTicker := time.NewTicker(time.Duration(eosmanager.config.EOSConfig.MonitorInterval) * time.Second)
	var blockHandleResult bool
	for {
		select {
		case <-fetchBlockTicker.C:
			//获取节点高度height. this.currentHeiht上次当前高度
			height, err := tools.GetEOSNodeHeight(eosmanager.eosClient)
			if err != nil {
				log.Infof("EOS MonitorEOSChain - cannot get node height, err: %s", err)
				continue
			}
			// 小于阈值不进行同步
			if height-eosmanager.currentHeight <= config.EOS_USEFUL_BLOCK_NUM {
				continue
			}
			log.Infof("EOS MonitorEOSChain - eos height is %d", height)
			blockHandleResult = true

			for eosmanager.currentHeight < height-config.EOS_USEFUL_BLOCK_NUM {
				if eosmanager.currentHeight%50 == 0 {
					log.Infof("EOS MonitorEOSChain - handle confirmed EOS block Height: %d", eosmanager.currentHeight)
				}
				blockHandleResult = eosmanager.handleNewBlock(eosmanager.currentHeight + 1)
				if !blockHandleResult {
					break
				}

				eosmanager.currentHeight++
				// 如果块大于等于EOSConfig.HeadersPerBatch提交一次
				if len(eosmanager.header4sync) >= eosmanager.config.EOSConfig.HeadersPerBatch {
					if res := eosmanager.commitHeader(); res != 0 {
						log.Infof("EOS MonitorEOSChain - per HeadersPerBatch commit header,header len is:%d", len(eosmanager.header4sync))
						blockHandleResult = false
						break
					}
				}
			}

			if blockHandleResult && len(eosmanager.header4sync) > 0 {
				log.Infof("EOS MonitorEOSChain - commit at for out")
				eosmanager.commitHeader()
			}
		case <-eosmanager.exitChan:
			return
		}
	}
}

func (eosmanager *EOSManager) handleNewBlock(height uint64) bool {
	ret := eosmanager.handleBlockHeader(height)
	if !ret {
		log.Errorf("EOS handleNewBlock - handleBlockHeader on height :%d failed", height)
		return false
	}
	ret = eosmanager.fetchLockDepositEvents(height, eosmanager.eosClient)
	if !ret {
		log.Errorf("EOS handleNewBlock - fetchLockDepositEvents on height :%d failed", height)
	}
	return true
}

func (eosmanager *EOSManager) handleBlockHeader(height uint64) bool {
	// 修改previous
	hdr, err := tools.GetEOSHeaderByNum(eosmanager.eosClient, uint32(height))
	if err != nil {
		log.Errorf("EOS handleBlockHeader - GetNodeHeader on height :%d failed", height)
		return false
	}
	// Proof额外

	hdr.Previous = *eosmanager.preBlockID

	blockID, err := hdr.BlockID()
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOS GetBlockID error: %v", err)
		return false
	}

	eosmanager.preBlockID = &blockID
	blockIDBytes, err := blockID.MarshalJSON()
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOS GetBlockIDBytes error: %v", err)
		return false
	}
	rawHdr, _ := eos.MarshalBinary(hdr)
	raw, _ := eosmanager.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		append(
			append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(eosmanager.config.EOSConfig.SideChainId)...,
			), autils.GetUint64Bytes(height)...,
		))
	// raw 获取blockID
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOSMarshalBinary header error: %v", err)
	}
	if len(raw) == 0 || !bytes.Equal(raw, blockIDBytes) {
		eosmanager.header4sync = append(eosmanager.header4sync, rawHdr)
	}
	return true
}

/*
获取指定事件
判断是否已经存在poly
获取事件后组装成crossTx
PutRetry
*/
func (eosmanager *EOSManager) fetchLockDepositEvents(height uint64, eosClient *eos.API) bool {

	events, merkleTree, err := eosmanager.filterCrossChainEvent(uint32(height), eosClient)
	if err != nil {
		log.Errorf("EOS fetchLockDepositEvents - filterCrossChainEvent error :%s\n", err)
		return false
	}
	if len(events) == 0 {
		return true
	}

	for _, event := range events {
		var isTarget bool
		if len(eosmanager.config.TargetContracts) > 0 {
			toContractStr := event.toContract
			for _, v := range eosmanager.config.TargetContracts {
				toChainIdArr, ok := v[toContractStr]
				if ok {
					if len(toChainIdArr["outbound"]) == 0 {
						isTarget = true
						break
					}
					for _, id := range toChainIdArr["outbound"] {
						if strconv.FormatUint(id, 10) == event.toChainId {
							isTarget = true
							break
						}
					}
					if isTarget {
						break
					}
				}
			}
			if !isTarget {
				continue
			}
		}

		log.Infof("<----EOS Filter after CrossChainEvent event.TargetID: %v,caller: %v,toContract: %v", event.toChainId, event.caller, event.toContract)
		param := &common2.MakeTxParam{}
		_ = param.Deserialization(common.NewZeroCopySource([]byte(event.rawParam)))
		raw, _ := eosmanager.polySdk.GetStorage(autils.CrossChainManagerContractAddress.ToHexString(),
			append(append([]byte(cross_chain_manager.DONE_TX), autils.GetUint64Bytes(eosmanager.config.EOSConfig.SideChainId)...), param.CrossChainID...))
		if len(raw) != 0 {
			log.Debugf("EOS fetchLockDepositEvents - ccid %s (tx_hash: %s) already on poly",
				hex.EncodeToString(param.CrossChainID), string(event.txHash))
			continue
		}
		proofTx, err := tools.GetEOSProof(merkleTree, event.leaf)

		if err != nil {
			log.Errorf("EOS fetchLockDepositEvents - get Merkle Proof error:%s", err)
		}

		proofBytes := common.NewZeroCopySink(nil)
		proofTx.Serialization(proofBytes)

		crossTx := &CrossTransfer{
			txIndex:     hex.EncodeToString(event.txHash),
			txId:        event.txId,
			toChain:     event.toChainId,
			value:       event.rawParam,
			fee:         event.feeData.fee,
			height:      height,
			caller:      event.caller,
			filterTime:  time.Now().Format("2006-01-02 15:04:05"),
			merkleProof: proofBytes.Bytes(),
		}

		log.Infof("---->EOS get crossTransfer:%v to chain:%d, the block height is%d", crossTx.txIndex, crossTx.toChain, crossTx.height)
		sink := common.NewZeroCopySink(nil)
		crossTx.Serialization(sink)
		err = eosmanager.db.PutRetry(sink.Bytes())
		if err != nil {
			log.Errorf("EOS fetchLockDepositEvents - this.db.PutRetry error: %s", err)
		}

		log.Infof("EOS fetchLockDepositEvents -  height: %d", height)
	}
	return true
}

/*
过滤跨链合约事件
*/
func (eosmanager *EOSManager) filterCrossChainEvent(height uint32, eosClient *eos.API) ([]TxActionData, *proof.MerkleTree, error) {

	res, err := tools.GetEOSTraceBlockByNum(eosClient, height)
	if err != nil {
		log.Errorf("EOS filterCrossChainEvent - error: %s", err)
		return nil, nil, err
	}
	resBlock, err := tools.GetEOSBlockByNum(eosClient, height)
	if err != nil {
		log.Errorf("EOS filterCrossChainEvent - error: %s", err)
		return nil, nil, err
	}

	var txActions []TxActionData

	for i, transaction := range res.Transactions {
		var txActionData = new(TxActionData)
		var sig bool = false
		for _, action := range transaction.Actions {

			if action.Action == "receiptpay" && action.Account == "ddc.contract" {
				resPayData, err := tools.GetEOSDeTraceData(eosClient, action.Account, eos.Name(action.Action), hex.EncodeToString([]byte(action.Data)))
				if err != nil {
					log.Errorf("EOS filterCrossChainEvent - trace receiptpay error: %s", err)
				}
				var feeData FeeActionData

				feeData.fee = resPayData["fee"].(string)
				txActionData.feeData = feeData

			}
			if action.Action == "crosschaine" && action.Account == "ddcccmanager" {
				sig = true
				resData, err := tools.GetEOSDeTraceData(eosClient, action.Account, eos.Name(action.Action), hex.EncodeToString([]byte(action.Data)))
				if err != nil {
					log.Errorf("EOS filterCrossChainEvent - trace crosschaine error: %s", err)
				}

				switch resData["toChainId"].(type) {
				case string:
					txActionData.toChainId, _ = resData["toChainId"].(string)
				case float64:
					txActionData.toChainId = strconv.FormatUint(uint64(resData["toChainId"].(float64)), 10)
				default:
					log.Error("EOS filterCrossChainEvent - error: toChainId转码失败")
				}
				// 字节数组的hash
				txActionData.toContract = ethcommon.BytesToAddress(tools.TransInterfacesToBytes(resData["toContract"].([]interface{}))).Hex() // 以太坊

				txActionData.txHash = tools.TransInterfacesToBytes(resData["paramTxHash"].([]interface{}))
				txActionData.caller = resData["caller"].(string) //wangzelong

				txActionData.rawParam = tools.TransInterfacesToBytes(resData["rawParam"].([]interface{}))

				txActionData.txId = transaction.ID
				trsByte := proof.SerializationTrans(resBlock.Transactions[i-1])
				txActionData.leaf = trsByte
			}

		}
		if sig {
			txActions = append(txActions, *txActionData)
		}
	}
	if len(txActions) > 0 {
		tree, _ := proof.NewTree(resBlock.Transactions)
		if err != nil {
			return txActions, nil, err
		}
		return txActions, tree, nil
	} else {
		return txActions, nil, err
	}

}

/*
将区块链同步到poly
*/
func (eosmanager *EOSManager) commitHeader() int {
	// 提交同步头

	tx, err := eosmanager.polySdk.Native.Hs.SyncBlockHeader(
		eosmanager.config.EOSConfig.SideChainId,
		eosmanager.polySigner.Address,
		eosmanager.header4sync,
		eosmanager.polySigner,
	)
	if err != nil {
		errDesc := err.Error()
		if strings.Contains(errDesc, "get the parent block failed") || strings.Contains(errDesc, "missing required field") {
			log.Warnf("EOS commitHeader - send transaction to poly chain err: %s", errDesc)
			eosmanager.rollBackToCommAncestor()
			return 0
		} else {
			log.Errorf("EOS commitHeader - send transaction to poly chain err: %s", errDesc)
			return 1
		}
	}
	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	for range tick.C {
		h, _ = eosmanager.polySdk.GetBlockHeightByTxHash(tx.ToHexString())
		curr, _ := eosmanager.polySdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}
	}
	log.Infof("EOS commitHeader - commit abount %v block header %s to poly chain and confirmed on height %d", len(eosmanager.header4sync), tx.ToHexString(), h)
	eosmanager.header4sync = make([][]byte, 0) // 提交后将header4sync数据归零
	return 0
}

// 回滚
func (eosmanager *EOSManager) rollBackToCommAncestor() {
	for ; ; eosmanager.currentHeight-- {

		raw, err := eosmanager.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
			append(append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(eosmanager.config.EOSConfig.SideChainId)...), autils.GetUint64Bytes(eosmanager.currentHeight)...))
		//没有找到，继续往下找,currentHeight 继续--
		if len(raw) == 0 || err != nil {
			continue
		}
		hdr, err := tools.GetEOSHeaderByNum(eosmanager.eosClient, uint32(eosmanager.currentHeight))
		if err != nil {
			log.Errorf("EOS rollBackToCommAncestor - failed to get header by number, so we wait for one second to retry: %v", err)
			time.Sleep(time.Second)
			eosmanager.currentHeight++
		}

		// proof 额外代码
		var sideChainIdBytes [8]byte
		binary.LittleEndian.PutUint64(sideChainIdBytes[:], eosmanager.config.EOSConfig.SideChainId)
		preBlockID, err := tools.GetPolyStorageHeaderID(eosmanager.polySdk, eosmanager.currentHeight-1, sideChainIdBytes)
		if err != nil {
			log.Errorf("EOS rollBackToCommAncestor - get preBlockID error:%s\n", err)
		}
		hdr.Previous = *preBlockID
		// end

		blockID, err := hdr.BlockID()
		if err != nil {
			log.Errorf("EOS rollBackToCommAncestor - EOS GetBlockID error: %v", err)
		}
		blockIDBytes, err := blockID.MarshalJSON()
		if err != nil {
			log.Errorf("EOS rollBackToCommAncestor - EOS GetBlockIDBytes error: %v", err)
		}

		if bytes.Equal(blockIDBytes, raw) {
			log.Infof("EOS rollBackToCommAncestor - find the common ancestor: %s(number: %d)", blockIDBytes, eosmanager.currentHeight)
			break
		}
	}
	eosmanager.header4sync = make([][]byte, 0)
}

// 监控Deposit
func (eosmanager *EOSManager) MonitorDeposit() {
	monitorTicker := time.NewTicker(time.Duration(eosmanager.config.EOSConfig.MonitorInterval) * time.Second)
	for {
		select {
		case <-monitorTicker.C:
			height, err := tools.GetEOSNodeHeight(eosmanager.eosClient)
			if err != nil {
				log.Errorf("EOS MonitorDeposit - cannot get eos node height, err: %s", err)
				continue
			}
			// 同步高度为poly侧的最新高度
			snycheight := eosmanager.findLastestHeight()
			log.Infof("EOS MonitorDeposit from eos - snyced eos height: %d,eos height: %d diff: %d", snycheight, height, height-snycheight)
			// 处理指定高度的跨链事件信息,监听同步高度，当高度更新，处理retry中内容
			eosmanager.handleLockDepositEvents(snycheight)
		case <-eosmanager.exitChan:
			return
		}
	}
}

func (eosmanager *EOSManager) handleLockDepositEvents(snycheight uint64) error {
	retryList, err := eosmanager.db.GetAllRetry()
	if err != nil {
		return fmt.Errorf("EOS handleLockDepositEvents - this.db.GetAllRetry error:%s", err)
	}
	for _, v := range retryList {
		time.Sleep(time.Second * 1)
		crosstx := new(CrossTransfer)
		err := crosstx.Deserialization(common.NewZeroCopySource(v))
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - retry.Deserialization error: %s", err)
			continue
		}

		if snycheight <= crosstx.height+eosmanager.config.EOSConfig.BlockConfig {
			continue
		}

		//1. commit proof to poly
		txHash, err := eosmanager.commitProof(uint32(crosstx.height), crosstx.merkleProof, crosstx.value, crosstx.txId)
		if err != nil {
			//异常错误
			if strings.Contains(err.Error(), "chooseUtxos, current utxo is not enough") {
				log.Infof("EOS handleLockDepositEvents - invokeNativeContract error: %s", err)
				continue
			} else {
				if err := eosmanager.db.DeleteRetry(v); err != nil {
					log.Errorf("EOS handleLockDepositEvents - this.db.DeleteRetry error: %s", err)
				}
				if strings.Contains(err.Error(), "tx already done") {
					log.Debugf("EOS handleLockDepositEvents - eos_tx %s already on poly", hex.EncodeToString(crosstx.txId))
				} else {
					log.Errorf("EOS handleLockDepositEvents - invokeNativeContract error for eos_tx %s: %s", hex.EncodeToString(crosstx.txId), err)
				}
				continue
			}
		}
		log.Infof("起始链----提交跨链交易（默克尔证明）----交易ID为:%v,Poly交易哈希为:%v", crosstx.txId, txHash)
		// wangzelong 构造起始链DDC跨链信息
		crossChainInfo, err := eosmanager.collectCrossChainInfo(crosstx.txIndex, crosstx.value, crosstx.filterTime, crosstx.fee, crosstx.caller)
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents collectCrossChainInfo error:%s", err)
		}

		//2.put crossInfo to CrossInfoSend db for checking
		var sink = new(common.ZeroCopySink)
		crossChainInfo.Serialization(sink)
		err = eosmanager.db.PutCrossInfoSend(sink.Bytes())
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - this.db.PutCrossInfoSend error: %s", err)
		}

		//3. put to check db for checking
		err = eosmanager.db.PutCheck(txHash, v)
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - this.db.PutCheck error: %s", err)
		}
		err = eosmanager.db.DeleteRetry(v)
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - this.db.PutCheck error: %s", err)
		}
		log.Infof("---->EOS handleLockDepositEvents - syncProof txHash is %s", txHash)
	}
	return nil
}

func (eosmanager *EOSManager) commitProof(height uint32, proof []byte, value []byte, txhash []byte) (string, error) {

	tx, err := eosmanager.polySdk.Native.Ccm.ImportOuterTransfer(
		eosmanager.config.EOSConfig.SideChainId,
		value,
		height,
		proof,
		tools.Hex2Bytes(eosmanager.polySigner.Address.ToHexString()),
		[]byte{},
		eosmanager.polySigner)

	if err != nil {
		eosmanager.count++
		return "", err
	} else {
		log.Infof("---->EOS commitProof - send transaction to poly chain: ( poly_txhash: %s, eos_txhash: %s, height: %d )",
			tx.ToHexString(), hex.EncodeToString(txhash), height)
		return tx.ToHexString(), nil
	}
}

func (eosmanager *EOSManager) CheckDeposit() {
	checkTicker := time.NewTicker(time.Duration(eosmanager.config.EOSConfig.MonitorInterval) * time.Second)
	for {
		select {
		case <-checkTicker.C:
			// try to check deposit
			eosmanager.checkLockDepositEvents()
		case <-eosmanager.exitChan:
			return
		}
	}
}

func (eosmanager *EOSManager) checkLockDepositEvents() error {
	checkMap, err := eosmanager.db.GetAllCheck()
	if err != nil {
		return fmt.Errorf("EOS checkLockDepositEvents - this.db.GetAllCheck error: %s", err)
	}
	for k, v := range checkMap {
		// k txHash
		event, err := eosmanager.polySdk.GetSmartContractEvent(k)
		if err != nil {
			log.Errorf("EOS checkLockDepositEvents - this.aliaSdk.GetSmartContractEvent error: %s", err)
			continue
		}
		if event == nil {
			continue
		}
		if event.State != 1 {
			log.Infof("EOS checkLockDepositEvents - state of poly tx %s is not success", k)
			err := eosmanager.db.PutRetry(v)
			if err != nil {
				log.Errorf("EOS checkLockDepositEvents - this.db.PutRetry error:%s", err)
			}
		}
		log.Infof("起始链----Poly出块,跨链交易成功,Poly交易Hash为:%v----", event.TxHash)
		err = eosmanager.db.DeleteCheck(k)
		if err != nil {
			log.Errorf("EOS checkLockDepositEvents - this.db.DeleteRetry error:%s", err)
		}
	}
	return nil
}

func (eosmanager *EOSManager) collectCrossChainInfo(txHash string, rawParam []byte, filterTime, fee, caller string) (*service.CrossChainInfo, error) {

	txParam := new(RawParam)
	if err := txParam.Deserialization(common.NewZeroCopySource(rawParam)); err != nil {
		return nil, fmt.Errorf("EOS collectCrossChainInfo deserilization RawParam error: %s", err)
	}

	argParam := new(ArgsParam)
	if err := argParam.Deserialization(common.NewZeroCopySource(txParam.args)); err != nil {
		return nil, fmt.Errorf("EOS collectCrossChainInfo deserilization ArgsParam error: %s", err)
	}

	crossInfo := service.NewCrossChainInfo()

	CrossChainID := new(big.Int)
	CrossChainID.SetString(strings.TrimLeft(hex.EncodeToString(txParam.crossChainID), "0"), 16)
	feeInt, err := tools.FeeStrToInt(fee)
	if err != nil {
		return nil, fmt.Errorf("collectCrossChainInfo fee Trans to int error :%v", err)
	}

	crossInfo.CrossChain_id = CrossChainID.Uint64()
	crossInfo.DDC_amount = uint32(argParam.amount)
	crossInfo.DDC_id = strconv.FormatUint(argParam.ddcId, 10)
	crossInfo.DDC_type = uint32(argParam.ddcType - 1)
	crossInfo.DDC_uri = hex.EncodeToString(argParam.ddcURI)
	crossInfo.From_address = hex.EncodeToString(argParam.fromOwner)
	crossInfo.From_cc_addr = hex.EncodeToString(txParam.fromContractAddress)
	crossInfo.From_chainid = strconv.FormatUint(eosmanager.config.EOSConfig.SideChainId, 10)
	crossInfo.From_tx = hex.EncodeToString(txParam.txHash)
	crossInfo.Poly_tx = txHash
	crossInfo.Sender = caller
	crossInfo.To_cc_addr = hex.EncodeToString(txParam.toContractAddress)
	crossInfo.To_address = hex.EncodeToString(argParam.toOwner)
	crossInfo.To_chainId = strconv.FormatUint(txParam.toChainID, 10)
	crossInfo.Tx_createtime = filterTime
	crossInfo.Cross_chain_fee = feeInt
	crossInfo.Tx_time = filterTime //起始链发送捕获交易时间

	return crossInfo, nil
}

func (eosmanager *EOSManager) SendDepositCrossInfo() {
	checkTicker := time.NewTicker(time.Duration(eosmanager.config.CollectInfoConfig.MonitorInterval) * time.Second)
	for {
		select {
		case <-checkTicker.C:
			//send CrossInfo
			log.Info("---->EOS SendDepositCrossInfo sendCrossChainMessage")
			err := eosmanager.sendCrossChainMessage()
			if err != nil {
				log.Errorf("SendDepositCrossInfo error :%s", err)
			}
		case <-eosmanager.exitChan:
			return
		}
	}
}

func (eosmanager *EOSManager) sendCrossChainMessage() error {
	sendList, err := eosmanager.db.GetAllCrossInfoSend()
	if err != nil {
		return fmt.Errorf("EOS sendCrossChainMessage - this.db.GetAllCrossInfoSend error: %s", err)
	}
	for _, v := range sendList {
		var crossChainInfo = new(service.CrossChainInfo)
		err := crossChainInfo.Deserialization(common.NewZeroCopySource(v))
		if err != nil {
			log.Errorf("error deserialization crossChainInfo:%v", crossChainInfo)
			log.Errorf("EOS sendCrossChainMessage - deserialization crossChainInfo error: %s", err)
			err = eosmanager.db.DeleteCrossInfoSend(v) //序列化失败，删掉该数据，日志记录
			if err != nil {
				log.Errorf("EOS sendCrossChainMessage - this.db.DeleteCrossInfoSend error: %s", err)
			}
			continue
		}
		res, msg, err := eosmanager.serviceClient.SendCrossChainInfo(crossChainInfo)
		log.Infof("<<<<<<<the service response is%v", res)
		if err != nil {
			log.Errorf("EOS sendCrossChainMessage - sendCrossChainInfo error: %s", err)
			continue
		}
		if res == "9002" {
			log.Errorf("EOS sendCrossChainMessage - Data content does not comform to the format, the crossChainId is:%v, the response is:%v", crossChainInfo.CrossChain_id, msg)
			err := eosmanager.db.DeleteCrossInfoSend(v) // 数据格式不对，删除数据
			if err != nil {
				log.Errorf("EOS sendCrossChainMessage - this.db.DeleteCrossInfoSend error: %s", err)
			}
			continue
		}
		if res == "9001" {
			// 服务器无响应，重试两次
			log.Infof("EOS sendCrossChainMessage - sendCrossChainMessage error msg is:%v, the crossChainId is:%v", msg, crossChainInfo.CrossChain_id)
			reRes := eosmanager.reSendCrossChainMessage(crossChainInfo)
			if reRes == "9001" { //重试两次失败，调用定时发送
				// 记录放入CrossInfoRetry表
				err := eosmanager.db.PutCrossInfoRetry(v)
				if err != nil {
					log.Errorf("EOS sendCrossChainMessage - this.db.PutCrossInfoRetry error: %s", err)
				}
				// 删除CrossInfoSend表中记录
				err = eosmanager.db.DeleteCrossInfoSend(v)
				if err != nil {
					log.Errorf("EOS sendCrossChainMessage - this.db.DeleteCrossInfoSend error: %s", err)
				}
				log.Info("重试发送两次失败，放入Retry表中，等待下次定时发送") //ToDo

			}
			continue
		}
		if res == "0000" {
			log.Infof("EOS sendCrossChainMessage - sendCrossChainInfo success, crossChainID is:%s", crossChainInfo.CrossChain_id)
			err = eosmanager.db.DeleteCrossInfoSend(v)
			if err != nil {
				log.Errorf("EOS sendCrossChainMessage - this.db.DeleteCrossInfoSend error: %s", err)
			}
		}
	}
	return nil
}

func (eosmanager *EOSManager) reSendCrossChainMessage(crossChainInfo *service.CrossChainInfo) string {
	var res string
	var err error
	for i := 0; i < eosmanager.config.CollectInfoConfig.ReSendNum; i++ {
		res, _, err = eosmanager.serviceClient.SendCrossChainInfo(crossChainInfo)
		if err != nil {
			log.Errorf("EOS sendCrossChainMessage - reSendCrossChainInfo index %d error: %s", i, err)
			continue
		}
		if res == "0000" {
			return res
		}

	}
	return res
}

func (eosmanager *EOSManager) RetryDepositCrossInfo() {
	retryTicker := time.NewTicker(time.Duration(eosmanager.config.CollectInfoConfig.RetryMonitorInterval) * time.Minute)
	for {
		select {
		case <-retryTicker.C:
			//try to send CrossInfo
			log.Info("每十分钟循环一次，发送重试跨链信息")
			err := eosmanager.retrySendCrossChainMessage()
			if err != nil {
				log.Errorf("RetrySendCrossChainMessage error:%s", err)
			}
		case <-eosmanager.exitChan:
			return
		}
	}
}

func (eosmanager *EOSManager) retrySendCrossChainMessage() error {
	retryList, err := eosmanager.db.GetAllCrossInfoRetry()
	if err != nil {
		return fmt.Errorf("EOS retrySendCrossChainMessage - this.db.GetAllCrossInfoRetry error: %s", err)
	}
	for _, v := range retryList {
		var crossChainInfo = new(service.CrossChainInfo)
		err := crossChainInfo.Deserialization(common.NewZeroCopySource(v))
		if err != nil {
			log.Errorf("EOS retrySendCrossChainMessage - deserialization crossChainInfo error: %s", err)
			err = eosmanager.db.DeleteCrossInfoRetry(v)
			if err != nil {
				log.Errorf("EOS retrySendCrossChainMessage - this.db.DeleteCrossInfoRetry error: %s", err)
			}
			continue
		}
		res, msg, err := eosmanager.serviceClient.SendCrossChainInfo(crossChainInfo)
		log.Infof("间隔十分钟Retry发送结果为:%v", res)
		if err != nil {
			log.Errorf("EOS retrySendCrossChainMessage - sendCrossChainInfo error: %s", err)
			err = eosmanager.db.DeleteCrossInfoRetry(v)
			if err != nil {
				log.Errorf("EOS retrySendCrossChainMessage - this.db.DeleteCrossInfoRetry error: %s", err)
			}
			continue
		}
		if res == "9002" {
			log.Errorf("EOS retrySendCrossChainMessage - Data content does not comform to the format,the crossChainId is:%v, response info is:%v ", crossChainInfo.CrossChain_id, msg)
			err = eosmanager.db.DeleteCrossInfoRetry(v)
			if err != nil {
				log.Errorf("EOS retrySendCrossChainMessage - this.db.DeleteCrossInfoRetry error: %s", err)
			}
			continue
		}
		if res == "9001" {
			log.Errorf("EOS retrySendCrossChainMessage -Service not response,the crossChainId is:%v ,response info is:%v", crossChainInfo.CrossChain_id, msg)
			continue
		}
		if res == "0000" {
			log.Infof("EOS retrySendCrossChainMessage - sendCrossChainInfo success ,crossChainId is:%v ", crossChainInfo.CrossChain_id)
			err = eosmanager.db.DeleteCrossInfoRetry(v)
			if err != nil {
				log.Errorf("EOS retrySendCrossChainMessage - this.db.DeleteCrossInfoRetry error: %s", err)
			}
		}

	}
	return nil
}
