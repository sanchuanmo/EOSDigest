package manager

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/proof"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"

	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"

	eos "github.com/qqtou/eos-go"
)

type CrossTransfer struct {
	txIndex     string // 交易索引
	txId        []byte // 交易id	源链event.TxHash().Bytes()
	value       []byte // 值		源链event.RawData()
	toChain     uint64 // 目标链	源链event.ToChainId
	height      uint64 // 高度		源链块高度
	merkleProof []byte // 默克尔证明
}

type TxActionData struct {
	toChainId  uint64
	toContract string
	caller     string
	txHash     []byte
	rawParam   []byte
	txId       []byte
	leaf       []byte
}

func (this *CrossTransfer) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.txIndex)
	sink.WriteVarBytes(this.txId)
	sink.WriteVarBytes(this.value)
	sink.WriteUint64(this.toChain)
	sink.WriteUint64(this.height)
	sink.WriteVarBytes(this.merkleProof)

}

// 反序列化
func (this *CrossTransfer) Deserialization(source *common.ZeroCopySource) error {
	txIndex, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize txIndex error")
	}
	txId, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize txId error")
	}
	value, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize value error")
	}
	toChain, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Waiting deserialize toChain error")
	}
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Waiting deserialize height error")
	}
	merkleProof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize merkleProof error")
	}
	this.txIndex = txIndex
	this.txId = txId
	this.value = value
	this.toChain = toChain
	this.height = height
	this.merkleProof = merkleProof
	return nil
}

type EOSManager struct {
	config *config.ServiceEOSConfig //service配置

	eosClient     *eos.API
	currentHeight uint64           // 当前高度
	forceHeight   uint64           // force高度
	preBlockID    *eos.Checksum256 // 父节点ID Proof相关
	polySdk       *sdk.PolySdk     // polySDK
	polySigner    *sdk.Account     // poly注册器
	exitChan      chan int         // exit chan
	header4sync   [][]byte         // 头同步
	crosstx4sync  []*CrossTransfer // 跨链交易同步
	db            *db.BoltDB       // blotDB
}

func NewEOSManager(servConfig *config.ServiceEOSConfig, startheight uint64, startforceheight uint64, polySdk *sdk.PolySdk, eosSdk *eos.API, boltDB *db.BoltDB) (*EOSManager, error) {

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
			log.Errorf("EOS NewEOSManager - wallet open error: %s", err.Error())
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
	}

	err = mgr.init()
	if err != nil {
		return nil, err
	} else {
		return mgr, nil
	}
}

func (this *EOSManager) init() error {
	// get latestheight
	latestHeight := this.findLastestHeight()
	if latestHeight == 0 {
		return fmt.Errorf("EOS init - the genesis block has not synced!")
	}
	if this.forceHeight > 0 && this.forceHeight < latestHeight {
		this.currentHeight = this.forceHeight
	} else {
		this.currentHeight = latestHeight
	}
	log.Infof("EOS init - findLastestHeight success, LastestHeight:%v", latestHeight)
	log.Infof("EOS init - start height: %d\n", this.currentHeight)
	//Proof额外代码
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], this.config.EOSConfig.SideChainId)
	preBlockID, err := tools.GetPolyStorageHeaderID(this.polySdk, latestHeight, sideChainIdBytes)
	if err != nil {
		return fmt.Errorf("EOS init - get preBlockID error:%s", err)
	}

	this.preBlockID = preBlockID
	// end
	return nil
}

// 查询Poly内存表CURRENT_HEADER_HEIGHT 存储EOS侧链最新高度
// 查询Poly内存表MAIN_CHAIN 存储EOS侧链最新高度 对应的BlockID
func (this *EOSManager) findLastestHeight() uint64 {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], this.config.EOSConfig.SideChainId)

	contractAddress := autils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), sideChainIdBytes[:]...)
	// try to get storage
	result, err := this.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil && result != nil {
		log.Errorf("findLastestHeight - GetStorage CURRENT_HEADER_HEIGHT error" + err.Error())
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

func (this *EOSManager) MonitorEOSChain() {
	fetchBlockTicker := time.NewTicker(time.Duration(this.config.EOSConfig.MonitorInterval) * time.Second)
	var blockHandleResult bool
	log.Infof("起始链----监听起始链区块----") //ToDo
	for {
		select {
		case <-fetchBlockTicker.C:
			//获取节点高度height. this.currentHeiht上次当前高度
			height, err := tools.GetEOSNodeHeight(this.eosClient)
			if err != nil {
				log.Infof("EOS MonitorEOSChain - cannot get node height, err: %s", err)
				continue
			}
			// 小于阈值不进行同步
			if height-this.currentHeight <= config.EOS_USEFUL_BLOCK_NUM {
				continue
			}
			log.Infof("EOS MonitorEOSChain - eos height is %d", height)
			blockHandleResult = true

			for this.currentHeight < height-config.EOS_USEFUL_BLOCK_NUM {
				if this.currentHeight%50 == 0 {
					log.Infof("EOS MonitorEOSChain - handle confirmed EOS block Height: %d", this.currentHeight)
				}
				blockHandleResult = this.handleNewBlock(this.currentHeight + 1)
				if blockHandleResult == false {
					break
				}

				this.currentHeight++
				// 如果块大于等于EOSConfig.HeadersPerBatch提交一次
				if len(this.header4sync) >= this.config.EOSConfig.HeadersPerBatch {
					if res := this.commitHeader(); res != 0 {
						log.Infof("EOS MonitorEOSChain - per HeadersPerBatch commit header,header len is:%d", len(this.header4sync))
						blockHandleResult = false
						break
					}
				}
			}

			if blockHandleResult && len(this.header4sync) > 0 {
				log.Infof("EOS MonitorEOSChain - commit at for out")
				this.commitHeader()
			}
		case <-this.exitChan:
			return
		}
	}
}

func (this *EOSManager) handleNewBlock(height uint64) bool {
	ret := this.handleBlockHeader(height)
	if !ret {
		log.Errorf("EOS handleNewBlock - handleBlockHeader on height :%d failed", height)
		return false
	}
	ret = this.fetchLockDepositEvents(height, this.eosClient)
	if !ret {
		log.Errorf("EOS handleNewBlock - fetchLockDepositEvents on height :%d failed", height)
	}
	return true
}

func (this *EOSManager) handleBlockHeader(height uint64) bool {
	// 修改previous
	hdr, err := tools.GetEOSHeaderByNum(this.eosClient, uint32(height))
	if err != nil {
		log.Errorf("EOS handleBlockHeader - GetNodeHeader on height :%d failed", height)
		return false
	}
	// Proof额外

	hdr.Previous = *this.preBlockID

	blockID, err := hdr.BlockID()
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOS GetBlockID error: %v", err)
		return false
	}

	this.preBlockID = &blockID
	blockIDBytes, err := blockID.MarshalJSON()
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOS GetBlockIDBytes error: %v", err)
		return false
	}
	rawHdr, _ := eos.MarshalBinary(hdr)
	raw, _ := this.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		append(
			append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(this.config.EOSConfig.SideChainId)...,
			), autils.GetUint64Bytes(height)...,
		))
	// raw 获取blockID
	if err != nil {
		log.Errorf("EOS handleBlockHeader - EOSMarshalBinary header error: %v", err)
	}
	if len(raw) == 0 || !bytes.Equal(raw, blockIDBytes) {
		this.header4sync = append(this.header4sync, rawHdr)
	}
	return true
}

/*
获取指定事件
判断是否已经存在poly
获取事件后组装成crossTx
PutRetry
*/
func (this *EOSManager) fetchLockDepositEvents(height uint64, eosClient *eos.API) bool {

	events, merkleTree, err := filterCrossChainEvent(uint32(height), eosClient)
	if err != nil {
		log.Errorf("EOS fetchLockDepositEvents - filterCrossChainEvent error :%s\n", err)
		return false
	}
	if len(events) == 0 {
		return true
	}

	for _, event := range events {
		var isTarget bool
		if len(this.config.TargetContracts) > 0 {
			toContractStr := event.toContract
			for _, v := range this.config.TargetContracts {
				toChainIdArr, ok := v[toContractStr]
				if ok {
					if len(toChainIdArr["outbound"]) == 0 {
						isTarget = true
						break
					}
					for _, id := range toChainIdArr["outbound"] {
						if id == event.toChainId {
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
		log.Info("起始链----筛选跨链事件的目标地址,目标链ID:%v,目标链合约地址:%v----", event.toContract, event.toContract)
		log.Infof("<----EOS Filter after CrossChainEvent event.TargetID: %v,caller: %v,toContract: %v", event.toChainId, event.caller, event.toContract)
		param := &common2.MakeTxParam{}
		_ = param.Deserialization(common.NewZeroCopySource([]byte(event.rawParam)))
		raw, _ := this.polySdk.GetStorage(autils.CrossChainManagerContractAddress.ToHexString(),
			append(append([]byte(cross_chain_manager.DONE_TX), autils.GetUint64Bytes(this.config.EOSConfig.SideChainId)...), param.CrossChainID...))
		if len(raw) != 0 {
			log.Debugf("EOS fetchLockDepositEvents - ccid %s (tx_hash: %s) already on poly",
				hex.EncodeToString(param.CrossChainID), string(event.txHash))
			continue
		}
		proof, err := tools.GetEOSProof(merkleTree, event.leaf)
		if err != nil {
			log.Errorf("EOS fetchLockDepositEvents - get Merkle Proof error:%s", err)
		}
		proofBytes := common.NewZeroCopySink(nil)
		proof.Serialization(proofBytes)

		crossTx := &CrossTransfer{
			txIndex:     string(event.txHash),
			txId:        event.txId,
			toChain:     event.toChainId,
			value:       event.rawParam,
			height:      height,
			merkleProof: proofBytes.Bytes(),
		}
		log.Infof("起始链----筛选跨链事件:构建Poly跨链交易,跨链交易ID:%v----", crossTx.txId) //ToDo
		log.Infof("---->EOS get crossTransfer %s to chain:%d, the block height is%d", crossTx.txIndex, crossTx.toChain, crossTx.height)
		sink := common.NewZeroCopySink(nil)
		crossTx.Serialization(sink)
		err = this.db.PutRetry(sink.Bytes())
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
func filterCrossChainEvent(height uint32, eosClient *eos.API) ([]TxActionData, *proof.MerkleTree, error) {

	res, err := tools.GetEOSTraceBlockByNum(eosClient, height)
	if err != nil {
		log.Error("EOS filterCrossChainEvent - error: %s", err)
		return nil, nil, err
	}
	resBlock, err := tools.GetEOSBlockByNum(eosClient, height)
	if err != nil {
		log.Error("EOS filterCrossChainEvent - error: %s", err)
		return nil, nil, err
	}

	var txActions []TxActionData
	for i, transaction := range res.Transactions {
		for _, action := range transaction.Actions {
			if action.Action != "onblock" {
				log.Infof("---->the block height %d, transaction [%d] action is:%s account is:%s", height, i, action.Action, action.Account)
			}
			if action.Action == "crosschaine" && action.Account == "ddcccmanager" {
				log.Infof("起始链----筛选跨链事件:监听到发起跨链----") //ToDo
				resData, err := tools.GetEOSDeTraceData(eosClient, action.Account, eos.Name(action.Action), action.Data.(string))
				if err != nil {
					log.Error("EOS filterCrossChainEvent - error: %s", err)
				}
				var txActionData TxActionData
				txActionData.toChainId = uint64(resData["toChainId"].(float64))
				txActionData.toContract = resData["toContract"].(string)
				txActionData.txHash = tools.TransInterfacesToBytes(resData["paramTxHash"].([]interface{}))
				txActionData.caller = resData["caller"].(string)
				txActionData.rawParam = tools.TransInterfacesToBytes(resData["rawParam"].([]interface{}))
				txActionData.txId = transaction.ID
				trsByte := proof.SerializationTrans(resBlock.Transactions[i-1])
				txActionData.leaf = trsByte
				txActions = append(txActions, txActionData)
			} else {
				continue
			}
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
func (this *EOSManager) commitHeader() int {
	// 提交同步头

	//测试日志
	var startHdr, endHdr *eos.SignedBlockHeader
	err := eos.UnmarshalBinary(this.header4sync[0], &startHdr)
	if err != nil {
		log.Errorf("EOS commitHeader get the header4sync [0] error: %s", err)
	}
	err = eos.UnmarshalBinary(this.header4sync[len(this.header4sync)-1], &endHdr)
	if err != nil {
		log.Errorf("EOS commitHeader get the header4sync [-1] error: %s", err)
	}
	log.Infof("EOS commitHeader the header4sync len is:%v,startheight is:%v,endheight is:%v", len(this.header4sync), startHdr.BlockNumber(), endHdr.BlockNumber())

	// 测试日志结束

	tx, err := this.polySdk.Native.Hs.SyncBlockHeader(
		this.config.EOSConfig.SideChainId,
		this.polySigner.Address,
		this.header4sync,
		this.polySigner,
	)
	if err != nil {
		errDesc := err.Error()
		if strings.Contains(errDesc, "get the parent block failed") || strings.Contains(errDesc, "missing required field") {
			log.Warnf("EOS commitHeader - send transaction to poly chain err: %s", errDesc)
			this.rollBackToCommAncestor()
			return 0
		} else {
			log.Errorf("EOS commitHeader - send transaction to poly chain err: %s", errDesc)
			return 1
		}
	}
	log.Infof("起始链----发起提交同步区块头----,提交区块头数量:%d,交易Hash为%s", len(this.header4sync), tx.ToHexString()) //ToDo
	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	for range tick.C {
		h, _ = this.polySdk.GetBlockHeightByTxHash(tx.ToHexString())
		curr, _ := this.polySdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}
	}
	log.Infof("EOS commitHeader - send transaction %s to poly chain and confirmed on height %d", tx.ToHexString(), h)
	log.Infof("起始链----提交同步区块头成功----,可在Poly链%d高度证明", h) //ToDo
	this.header4sync = make([][]byte, 0)               // 提交后将header4sync数据归零
	return 0
}

// 回滚
func (this *EOSManager) rollBackToCommAncestor() {
	for ; ; this.currentHeight-- {

		raw, err := this.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
			append(append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(this.config.EOSConfig.SideChainId)...), autils.GetUint64Bytes(this.currentHeight)...))
		//没有找到，继续往下找,currentHeight 继续--
		if len(raw) == 0 || err != nil {
			continue
		}
		hdr, err := tools.GetEOSHeaderByNum(this.eosClient, uint32(this.currentHeight))
		if err != nil {
			log.Errorf("EOS rollBackToCommAncestor - failed to get header by number, so we wait for one second to retry: %v", err)
			time.Sleep(time.Second)
			this.currentHeight++
		}

		// proof 额外代码
		var sideChainIdBytes [8]byte
		binary.LittleEndian.PutUint64(sideChainIdBytes[:], this.config.EOSConfig.SideChainId)
		preBlockID, err := tools.GetPolyStorageHeaderID(this.polySdk, this.currentHeight-1, sideChainIdBytes)
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
			log.Infof("EOS rollBackToCommAncestor - find the common ancestor: %s(number: %d)", blockIDBytes, this.currentHeight)
			break
		}
	}
	this.header4sync = make([][]byte, 0)
}

// 监控Deposit
func (this *EOSManager) MonitorDeposit() {
	monitorTicker := time.NewTicker(time.Duration(this.config.EOSConfig.MonitorInterval) * time.Second)
	for {
		select {
		case <-monitorTicker.C:
			height, err := tools.GetEOSNodeHeight(this.eosClient)
			if err != nil {
				log.Errorf("EOS MonitorDeposit - cannot get eos node height, err: %s", err)
				continue
			}
			// 同步高度为poly侧的最新高度
			snycheight := this.findLastestHeight()
			log.Infof("EOS MonitorDeposit from eos - snyced eos height: %d,eos height: %d diff: %d", snycheight, height, height-snycheight)
			// 处理指定高度的跨链事件信息,监听同步高度，当高度更新，处理retry中内容
			this.handleLockDepositEvents(snycheight)
		case <-this.exitChan:
			return
		}
	}
}

func (this *EOSManager) handleLockDepositEvents(snycheight uint64) error {
	retryList, err := this.db.GetAllRetry()
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

		if snycheight <= crosstx.height+this.config.EOSConfig.BlockConfig {
			continue
		}
		// 测试日志
		log.Infof("---->EOS handleLockDepositEvents retryProof %v,height %v", crosstx.merkleProof, crosstx.height)

		//1. commit proof to poly
		txHash, err := this.commitProof(uint32(crosstx.height), crosstx.merkleProof, crosstx.value, crosstx.txId)
		if err != nil {
			//异常错误
			if strings.Contains(err.Error(), "chooseUtxos, current utxo is not enough") {
				log.Infof("EOS handleLockDepositEvents - invokeNativeContract error: %s", err)
				continue
			} else {
				if err := this.db.DeleteRetry(v); err != nil {
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
		log.Infof("起始链----提交跨链交易（默克尔证明）----交易ID为:%v,Poly交易哈希为:%v", crosstx.txId, txHash) //ToDo
		//2. put to check db for checking
		err = this.db.PutCheck(txHash, v)
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - this.db.PutCheck error: %s", err)
		}
		err = this.db.DeleteRetry(v)
		if err != nil {
			log.Errorf("EOS handleLockDepositEvents - this.db.PutCheck error: %s", err)
		}
		log.Infof("---->EOS handleLockDepositEvents - syncProof txHash is %s", txHash)
	}
	return nil
}

func (this *EOSManager) commitProof(height uint32, proof []byte, value []byte, txhash []byte) (string, error) {
	log.Infof("EOS commitProof -  height: %d, proof: %v, value: %s, txhash: %s", height, proof, hex.EncodeToString(value), hex.EncodeToString(txhash))

	tx, err := this.polySdk.Native.Ccm.ImportOuterTransfer(
		this.config.EOSConfig.SideChainId,
		value,
		height,
		proof,
		tools.Hex2Bytes(this.polySigner.Address.ToHexString()),
		[]byte{},
		this.polySigner)
	if err != nil {
		return "", err
	} else {
		log.Infof("---->EOS commitProof - send transaction to poly chain: ( poly_txhash: %s, eos_txhash: %s, height: %d )",
			tx.ToHexString(), hex.EncodeToString(txhash), height)
		return tx.ToHexString(), nil
	}
}

func (this *EOSManager) CheckDeposit() {
	checkTicker := time.NewTicker(time.Duration(this.config.EOSConfig.MonitorInterval) * time.Second)
	for {
		select {
		case <-checkTicker.C:
			// try to check deposit
			this.checkLockDepositEvents()
		case <-this.exitChan:
			return
		}
	}
}

func (this *EOSManager) checkLockDepositEvents() error {
	checkMap, err := this.db.GetAllCheck()
	if err != nil {
		return fmt.Errorf("EOS checkLockDepositEvents - this.db.GetAllCheck error: %s", err)
	}
	for k, v := range checkMap {
		// k txHash
		event, err := this.polySdk.GetSmartContractEvent(k)
		if err != nil {
			log.Errorf("EOS checkLockDepositEvents - this.aliaSdk.GetSmartContractEvent error: %s", err)
			continue
		}
		if event == nil {
			continue
		}
		if event.State != 1 {
			log.Infof("EOS checkLockDepositEvents - state of poly tx %s is not success", k)
			err := this.db.PutRetry(v)
			if err != nil {
				log.Errorf("EOS checkLockDepositEvents - this.db.PutRetry error:%s", err)
			}
		}
		log.Infof("起始链----Poly出块,跨链交易成功,Poly交易Hash为:%v----", event.TxHash) //ToDo
		err = this.db.DeleteCheck(k)
		if err != nil {
			log.Errorf("EOS checkLockDepositEvents - this.db.DeleteRetry error:%s", err)
		}
	}
	return nil
}
