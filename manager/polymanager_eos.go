package manager

import (
	"bytes"
	"context"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ontio/ontology-crypto/keypair"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/contract"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	polytypes "github.com/polynetwork/poly/core/types"

	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	eos "github.com/qqtou/eos-go"
)

const (
	ChanLen = 64
)

var (
	testLog = log.InitLogTestData(2, "./Log/", log.Stdout)
)

type PolyManagerEOS struct {
	config        *config.ServiceEOSConfig
	polySdk       *sdk.PolySdk
	currentHeight uint64
	exitChan      chan int
	db            *db.BoltDB
	eosclient     *eos.API
	senders       []*EOSSender
}

type EOSSender struct {
	acc       *tools.EOSKeyStore
	cmap      map[string]chan *EOSTxInfo
	eosClient *eos.API
	polySdk   *sdk.PolySdk
	config    *config.ServiceEOSConfig
}

// 打包成eos中的action
type EOSTxInfo struct {
	basics     *contract.Basics //调用基础参数
	txData     []byte           //上链数据
	prkey      string           //签名私钥
	polyTxHash string
}

type CrossStatus struct {
	bolckNum   uint64
	txId       string
	sendStatus bool
}

// 序列化
func (this *CrossStatus) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.txId)
	sink.WriteUint64(this.bolckNum)
	sink.WriteBool(this.sendStatus)
}

// 反序列化
func (this *CrossStatus) Deserializaion(source *common.ZeroCopySource) error {
	txId, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize txId error")
	}
	bolckNum, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Waiting deserialize bolckNum error")
	}
	sendStatus, eof := source.NextBool()
	if eof {
		return fmt.Errorf("Waiting deserialize sendStatus error")
	}
	this.bolckNum = bolckNum
	this.txId = txId
	this.sendStatus = sendStatus

	return nil
}

func NewPolyManagerEOS(servcfg *config.ServiceEOSConfig, startblockHeight uint64, polySdk *sdk.PolySdk, eosClient *eos.API, boltDB *db.BoltDB) (*PolyManagerEOS, error) {

	eosKeyStore := tools.NewEOSKeyStore(servcfg.EOSConfig)

	senders := make([]*EOSSender, len(eosKeyStore))

	for i, v := range eosKeyStore {
		a := &EOSSender{}
		a.acc = v
		a.cmap = make(map[string]chan *EOSTxInfo)
		a.polySdk = polySdk
		a.config = servcfg
		a.eosClient = eosClient
		senders[i] = a
	}

	return &PolyManagerEOS{
		exitChan:      make(chan int),
		config:        servcfg,
		polySdk:       polySdk,
		currentHeight: startblockHeight,
		db:            boltDB,
		eosclient:     eosClient,
		senders:       senders,
	}, nil
}

/*
最新高度
查询eos内存表获取Poly的当前轮次起始高度
*/
func (this *PolyManagerEOS) findLatestHeight() uint64 {

	height, err := tools.GetEOSStartHeight(this.eosclient, this.config.EOSConfig.ContractAddress, tools.CROSSCONTRACTTABLE)
	if err != nil {
		log.Errorf("findLatestHeight - GetLatestHeight failed: %s", err.Error())
		return 0
	}
	return uint64(height)
}

/*
初始化
同步poly最新区块号
*/
func (this *PolyManagerEOS) init() bool {
	if this.currentHeight > 0 {
		log.Infof("PolyManagerEOS init -start height from flag: %d", this.currentHeight)
		return true
	}
	this.currentHeight = this.db.GetPolyHeight()
	lastestHeight := this.findLatestHeight()
	if lastestHeight > this.currentHeight {
		this.currentHeight = lastestHeight
		log.Infof("PolyManagerEOS init - latest height from ECCM %d", this.currentHeight)
	}
	log.Infof("PolyManagerEOS init - lastest height from DB:%d", this.currentHeight)
	return true
}

/*
监听链
初始化：从数据库中读取上次处理的最新高度到this.currentHeight
定时读取poly当前的最新高度lastestheight
从lastestheight往前遍历匹配
*/
func (this *PolyManagerEOS) MonitorChain() {
	ret := this.init()
	if ret == false {
		log.Errorf("MonitorChain - init failed")
	}
	// 定时任务
	log.Infof("目标链----监听Poly区块----") //ToDo
	monitorTicker := time.NewTicker(config.POLY_MONITOR_INTERVAL)
	var blockHandleResult bool
	for {
		select {
		case <-monitorTicker.C:
			//获取当前块高度
			lastestheight, err := this.polySdk.GetCurrentBlockHeight()
			if err != nil {
				log.Errorf("MonitorChain - get poly chain block height error: %s", err)
				continue
			}
			lastestheight--
			if uint64(lastestheight)-this.currentHeight < config.POLY_USEFUL_BLOCK_NUM {
				continue
			}
			log.Infof("MonitorChain - poly chain current height: %d", lastestheight)
			blockHandleResult = true
			for this.currentHeight <= uint64(lastestheight)-config.POLY_USEFUL_BLOCK_NUM {
				// 每处理10次日志记录。
				if this.currentHeight%10 == 0 {
					log.Infof("handle confirmed poly Block height: %d", this.currentHeight)
				}
				// 处理区块头
				blockHandleResult = this.handleDepositEvents(this.currentHeight)
				if blockHandleResult == false {
					break
				}
				this.currentHeight++
			}
			// 异常处理的冗余机制（将处理完的当前高度存入DB）
			if err = this.db.UpdatePolyHeight(this.currentHeight - 1); err != nil {
				log.Errorf("MonitorChain - failed to save height of poly: %v", err)
			}
		case <-this.exitChan:
			return
		}
	}
}

func (this *PolyManagerEOS) IsEpoch(hdr *polytypes.Header) (bool, []byte, error) {
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(hdr.ConsensusPayload, blkInfo); err != nil {
		return false, nil, fmt.Errorf("commitHeader - unmarshal blockInfo error : %s", err)
	}
	// 解析hdr.ConsensusPayload
	if hdr.NextBookkeeper == common.ADDRESS_EMPTY || blkInfo.NewChainConfig == nil {
		return false, nil, nil
	}
	// 从目标链获取数据GetCurEpochConPubKeyBytes
	rawKeepers, err := tools.GetEOSRawKeepers(this.eosclient, this.config.EOSConfig.ContractAddress, tools.CROSSCONTRACTTABLE)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current epoch keepers: %v", err)
	}
	// 从poly header consensusPayload.newchainconfig.peers中解析出bookkeepers
	var bookkeepers []keypair.PublicKey
	for _, peer := range blkInfo.NewChainConfig.Peers {
		keystr, _ := hex.DecodeString(peer.ID)
		key, _ := keypair.DeserializePublicKey(keystr)
		bookkeepers = append(bookkeepers, key)
	}
	bookkeepers = keypair.SortPublicKeys(bookkeepers)
	publickeys := make([]byte, 0)
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(uint64(len(bookkeepers)))
	for _, key := range bookkeepers {
		raw := tools.GetNoCompresskey(key)
		publickeys = append(publickeys, raw...)
		sink.WriteVarBytes(crypto.SHA256.New().Sum(raw[3:]))
	}

	if bytes.Equal(rawKeepers, sink.Bytes()) {
		return false, nil, nil
	}

	return true, publickeys, nil
}

// 处理区块头

func (this *PolyManagerEOS) handleDepositEvents(height uint64) bool {
	lastEpoch := this.findLatestHeight()
	// height 这次的currentHeight
	hdr, err := this.polySdk.GetHeaderByHeight(uint32(height) + 1) //下一个块
	if err != nil {
		log.Errorf("handleDepositEvents - GetNodeHeader on height : %d failed", height)
		return false
	}

	//isCurr:当前轮,lastEpoch:上一次监听执行的最新高度(即上一次的lastestHeight)
	isCurr := lastEpoch < height+1
	// 当NextBookkeeper == common.ADDRESS_EMPTY空地址或blkInfo.NewChainConfig == nil的时候表示是本周期内的其他非同步块
	// isEpoch 指的是是否是本poly周期内，false表示是本周期内，true表示非本周期内,非本周期内需要更新pubkList
	isEpoch, pubkList, err := this.IsEpoch(hdr)
	log.Infof("目标链----Poly周期性验证,验证块号为:%v----", height+1) //ToDo
	if err != nil {
		log.Errorf("falied to chech isEpoch: %v", err)
		return false
	}
	var (
		anchor *polytypes.Header
		hp     string
	)
	if !isCurr { // isCurr = false 非当前轮次，即因意外遗留的未处理块，从当前轮次的第一个块，获取块头，作为下一个证明节点
		anchor, _ = this.polySdk.GetHeaderByHeight(uint32(lastEpoch) + 1)
		proof, _ := this.polySdk.GetMerkleProof(uint32(height)+1, uint32(lastEpoch)+1)
		hp = proof.AuditPath
	} else if isEpoch { //isEpoch = true 刚好是当前轮次的证明块，
		anchor, _ = this.polySdk.GetHeaderByHeight(uint32(height) + 2)
		proof, _ := this.polySdk.GetMerkleProof(uint32(height)+1, uint32(height)+2)
		hp = proof.AuditPath
	}

	cnt := 0
	events, err := this.polySdk.GetSmartContractEventByBlock(uint32(height))
	for err != nil {
		log.Errorf("handleDepositEvents - get block event at height:%d error:%s", height, err.Error())
		return false
	}
	for _, event := range events {
		for _, notify := range event.Notify {
			if notify.ContractAddress == this.config.PolyConfig.EntranceContractAddress {

				states := notify.States.([]interface{})
				method, _ := states[0].(string)
				if method != "makeProof" {
					continue
				}
				if uint64(states[2].(float64)) != this.config.EOSConfig.SideChainId {
					continue
				}

				// 从Poly获取跨链交易证明
				proof, err := this.polySdk.GetCrossStatesProof(hdr.Height-1, states[5].(string))
				if err != nil {
					log.Errorf("handleDepositEvents - failed to get proof for key %s: %v", states[5].(string), err)
					continue
				}
				// log.Infof("目标链----获取Poly的跨链交易证明----") //ToDo
				auditpath, _ := hex.DecodeString(proof.AuditPath)
				value, _, _, _ := tools.ParseAuditpath(auditpath)
				param := &common2.ToMerkleValue{}
				if err := param.Deserialization(common.NewZeroCopySource(value)); err != nil {
					log.Errorf("handleDepositEvents - failed to deserialize MakeTxParam (value: %x, err: %v)", value, err)
					continue
				}
				//ToDo
				// log.Infof("目标链----->跨链交易目标链地址(byte):%v\n,目标链地址(string):%v", param.MakeTxParam.ToContractAddress, string(param.MakeTxParam.ToContractAddress))
				// log.Infof("目标链----->跨链交易目标链ID(uint64):%v\n", param.MakeTxParam.ToChainID)
				// log.Infof("目标链----->跨链交易目标链方法(byte):%v\n", param.MakeTxParam.Method)
				// log.Infof("目标链----->跨链交易目标链属性(byte):%v\n", param.MakeTxParam.Args)
				// log.Infof("目标链----->跨链交易起始链地址(byte):%v\n", param.MakeTxParam.FromContractAddress)
				// log.Infof("目标链----->跨链交易起始链ID(byte):%v\n", param.MakeTxParam.CrossChainID)
				//
				var isTarget bool
				// 向目标合约地址发送
				if len(this.config.TargetContracts) > 0 {
					toContractStr := string(param.MakeTxParam.ToContractAddress)
					for _, v := range this.config.TargetContracts {
						toChainIdAddr, ok := v[toContractStr]
						if ok {
							if len(toChainIdAddr["inbound"]) == 0 {
								isTarget = true
								break
							}
							for _, id := range toChainIdAddr["inbound"] {
								if id == param.FromChainID {
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
				// log.Infof("目标链----筛选满足起始链范围的跨链交易----") //ToDo
				cnt++
				sender := this.selectSender()
				log.Infof("sender %v is handling poly tx ( hash: %v, height: %d) ", sender.acc.AccountName, param.TxHash, height)
				// 忽略tx错误
				if !sender.commitVerifyTx(hdr, param, hp, anchor, event.TxHash, auditpath) {
					return false
				}

			}
		}
	}
	if cnt == 0 && isEpoch && isCurr {
		sender := this.selectSender()
		return sender.commitChbook(hdr, pubkList)
	}
	return true
}

// 选取sender
func (this *PolyManagerEOS) selectSender() *EOSSender {

	seed := rand.New(rand.NewSource(time.Now().Unix()))
	num := seed.Intn(len(this.senders))

	return this.senders[num]
}

/*
发送verifyexetxe交易到目标链管理合约
*/
func (this *EOSSender) commitVerifyTx(header *polytypes.Header, param *common2.ToMerkleValue, anchorHeaderProof string, anchorHeader *polytypes.Header, polyTxHash string, headerProof []byte) bool {
	//打包数据
	log.Infof("目标链----构造跨链管理合约交易----") //ToDo
	var (
		sigs       []byte
		headerData []byte
		rawAnchor  []byte
	)
	if anchorHeader != nil && anchorHeaderProof != "" {
		for _, sig := range anchorHeader.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			//转换格式
			newsig, _ := tools.ConvertToEosCompatible(temp)
			sigs = append(sigs, newsig...)
		}
	} else {
		for _, sig := range header.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			//转换格式
			newsig, _ := tools.ConvertToEosCompatible(temp)
			sigs = append(sigs, newsig...)
		}
	}
	headerData = header.GetMessage()
	hp, _ := hex.DecodeString(anchorHeaderProof)
	if anchorHeader != nil {
		rawAnchor = anchorHeader.GetMessage()
	}

	txData := &contract.InputVerifyexetx{
		Proof:        headerProof,
		RawHeader:    headerData,
		HeaderProof:  hp,
		CurRawHeader: rawAnchor,
		HeaderSig:    sigs,
	}
	txDataByte, err := json.Marshal(txData)
	if err != nil {
		fmt.Printf("json.Marshal err:%v", err)
	}
	basicsTx := &contract.Basics{
		Caller:     eos.AccountName(this.acc.AccountName),
		Contract:   eos.AccountName(this.config.EOSConfig.ContractAddress),
		ActionName: eos.ActionName(contract.VERIFYEXETXE),
		Per:        "active",
	}

	eosTx := &EOSTxInfo{
		polyTxHash: polyTxHash,
		prkey:      this.acc.Ks.String(),
		basics:     basicsTx,
		txData:     txDataByte,
	}

	k := this.getRouter()
	c, ok := this.cmap[k]
	if !ok {
		c = make(chan *EOSTxInfo, ChanLen)
		this.cmap[k] = c
		go func() {
			for v := range c {
				if err = this.sendTxToEOS(v); err != nil {
					log.Errorf("failed to send tx to eos: error: %v, txData: %s", err, hex.EncodeToString(v.txData))
				}
			}
		}()
	}
	//
	c <- eosTx

	return true
}

/*
发送chbookkeeper交易到目标链管理合约
*/
func (this *EOSSender) commitChbook(header *polytypes.Header, pubkList []byte) bool {
	headerdata := header.GetMessage()
	headerHash := header.Hash()

	var sigs []byte

	for _, sig := range header.SigData {
		temp := make([]byte, len(sig))
		copy(temp, sig)
		newsig, _ := tools.ConvertToEosCompatible(temp)
		sigs = append(sigs, newsig...)
	}

	basicBk := &contract.Basics{
		Caller:     eos.AccountName(this.acc.AccountName),
		Contract:   eos.AccountName(this.config.EOSConfig.ContractAddress),
		ActionName: eos.ActionName(contract.CHBOOKKEEPER),
		Per:        "active",
	}

	txDataBK := &contract.InputChbookkeeper{
		RawHeader:  headerdata,
		PubKeyList: pubkList,
		SigList:    sigs,
	}
	txDataByte, err := json.Marshal(txDataBK)
	if err != nil {
		log.Errorf("commit chbook - err:" + err.Error())
		return false
	}

	info := &EOSTxInfo{
		basics:     basicBk,
		txData:     txDataByte,
		prkey:      this.acc.Ks.String(),
		polyTxHash: headerHash.ToHexString(),
	}
	this.sendTxToEOS(info)
	return true
}

func (this *EOSSender) getRouter() string {
	return strconv.FormatInt(rand.Int63n(this.config.RoutineNum), 10)
}

/*
发送交易通用：
合约部署方账号，合约方法，参数
发送交易后解析返回获取交易ID，块号
记录入库
往后轮询块比对交易ID&&当前的块号<最新的不可逆块 ：验证交易成功。
*/
func (this *EOSSender) sendTxToEOS(info *EOSTxInfo) error {
	basics := info.basics
	var ctx context.Context = context.Background()
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, info.prkey) // 导入私钥
	if err != nil {
		log.Errorf("import private key: %v", err)
		return err
	}
	this.eosClient.SetSigner(keyBag) // 设置签名

	txOpts := &eos.TxOptions{}
	// 将HeadBlockID与ChainID填充到txOpts
	if err := txOpts.FillFromChain(ctx, this.eosClient); err != nil {
		log.Errorf("filling tx opts:%v", err)
		return err
	}
	// 构建交易
	var tx *eos.Transaction
	switch basics.ActionName {
	case "chbookkeeper":
		var input contract.InputChbookkeeper
		json.Unmarshal(info.txData, &input)
		testLog.Debugf("method:chbookkeeper\nRawHeader:%v\nPubKeyList:%v\n,SigList:%v\n", input.RawHeader, input.PubKeyList, input.SigList)
		tx = eos.NewTransaction([]*eos.Action{basics.Chbookkeeper(input.RawHeader, input.PubKeyList, input.SigList)}, txOpts)
	case "verifyexetx":
		var input contract.InputVerifyexetx
		json.Unmarshal(info.txData, &input)
		// log.Infof("input crossChain Proof is:%v", input.Proof)
		// log.Infof("input crossChain RawHeader is:%v", input.RawHeader)
		// log.Infof("input crossChain HeaderProof is:%v", input.HeaderProof)
		// log.Infof("input crossChain CurRawHeader is:%v", input.CurRawHeader)
		// log.Infof("input crossChain HeaderSig is:%v", input.HeaderSig)
		testLog.Debugf("method:verifyexetx\nProof:%v\nRawHeader:%v\n,HeaderProof:%v\nCurRawHeader:%v\nHeaderSig:%v\n", input.Proof, input.RawHeader, input.HeaderProof, input.CurRawHeader, input.HeaderSig)
		tx = eos.NewTransaction([]*eos.Action{basics.Verifyexetx(input.Proof, input.RawHeader, input.HeaderProof, input.CurRawHeader, input.HeaderSig)}, txOpts)
	case "crosschain":
		var input contract.InputCrosschain
		json.Unmarshal(info.txData, &input)
		tx = eos.NewTransaction([]*eos.Action{basics.Crosschain(input.ToChainId, input.ToContract, input.Method, input.TxData)}, txOpts)
	default:
		log.Errorf("NewTransaction err,actionName:%v not found", basics.ActionName)
	}

	// 签名并打包交易
	// signedTx, packedTx, err := this.eosClient.SignTransaction(ctx, tx, txOpts.ChainID, eos.CompressionNone)
	signedTx, packedTx, err := this.eosClient.SignTransaction(ctx, tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		log.Errorf("sign transaction: %v", err)
		return err
	}

	content, err := json.MarshalIndent(signedTx, "", "  ")
	if err != nil {
		log.Errorf("json marshalling transaction: %v", err)
	}
	fmt.Printf("signedTx:%v\n", string(content)) // TODO调试输出后续删除

	// push打包后的签名交易
	response, err := this.eosClient.PushTransaction(ctx, packedTx)
	// log.Infof("目标链----发送跨链交易到跨链管理合约----") //ToDo
	if err != nil {
		log.Errorf("push transaction:%v", err)
		return err
	}
	// resStr, _ := json.Marshal(response)
	log.Infof("目标链----发送跨链交易到跨链管理合约成功----") //ToDo
	// log.Infof("PushTransaction success, txId:%d\n", hex.EncodeToString(response.Processed.ID))
	log.Infof("---->PushTransaction success: nomal:%v\n", response)
	// log.Infof("---->PushTransaction success: json:%v\n", string(resStr))

	return nil
}

func (this *PolyManagerEOS) Stop() {
	this.exitChan <- 1
	close(this.exitChan)
	log.Infof("poly chain manager exit")
}