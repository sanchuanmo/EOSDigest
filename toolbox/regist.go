package toolbox

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/contract"
	"github.com/polynetwork/eos_relayer/tools"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	polytypes "github.com/polynetwork/poly/core/types"
	"github.com/qqtou/eos-go"
)

type RegisterEOS struct {
	config    *config.ServiceEOSConfig
	polySdk   *sdk.PolySdk
	eosclient *eos.API
	senders   []*EOSSender
}

type EOSSender struct {
	acc       *tools.EOSKeyStore
	eosTxInfo *EOSTxInfo
	eosClient *eos.API
	polySdk   *sdk.PolySdk
	config    *config.ServiceEOSConfig
}

// 打包成eos中的action,
type EOSTxInfo struct {
	basics     *contract.Basics //调用基础参数
	txData     []byte           //上链数据
	prkey      string           //签名私钥
	polyTxHash string
}

type RspBlockHeader struct {
	Timestamp        time.Time `json:"timestamp"`
	Producer         string    `json:"producer"`
	Confirmed        uint16    `json:"confirmed"`
	Previous         []byte    `json:"previous"`
	TransactionMRoot []byte    `json:"transaction_mroot"`
	ActionMRoot      []byte    `json:"action_mroot"`
	ScheduleVersion  uint32    `json:"schedule_version"`

	// EOSIO 1.x
	NewProducersV1 *eos.ProducerSchedule `json:"new_producers,omitempty" eos:"optional"`

	HeaderExtensions []*eos.Extension `json:"header_extensions"`
}

var WalletPath string = "../wallet2.dat"
var WalletPwd = "4cUYqGj2yib718E7ZmGQc"
var WalletAcc1 = "../wallet1.dat"
var WalletAcc2 = "../wallet2.dat"
var WalletAcc3 = "../wallet3.dat"
var WalletAcc4 = "../wallet4.dat"

func NewRegisterEOS(servcfg *config.ServiceEOSConfig, polySdk *sdk.PolySdk, eosSdk *eos.API) *RegisterEOS {

	eosKeyStore := tools.NewEOSKeyStore(servcfg.EOSConfig)

	senders := make([]*EOSSender, len(eosKeyStore))

	for i, v := range eosKeyStore {
		a := &EOSSender{}
		a.acc = v
		a.eosTxInfo = new(EOSTxInfo)
		a.polySdk = polySdk
		a.config = servcfg
		a.eosClient = eosSdk
		senders[i] = a
	}

	return &RegisterEOS{
		config:    servcfg,
		polySdk:   polySdk,
		eosclient: eosSdk,
		senders:   senders,
	}
}

func (this *RegisterEOS) getAccount() (*sdk.Account, error) {
	polyWallet, err := this.polySdk.OpenWallet(WalletPath)
	if err != nil {
		return nil, err
	}
	acc, err := polyWallet.GetDefaultAccount([]byte(WalletPwd))
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (this *RegisterEOS) getAccounts() ([]*sdk.Account, error) {

	var polyWallet []*sdk.Wallet
	var accs []*sdk.Account
	polyWallet1, _ := this.polySdk.OpenWallet(WalletAcc1)
	polyWallet2, _ := this.polySdk.OpenWallet(WalletAcc2)
	polyWallet3, _ := this.polySdk.OpenWallet(WalletAcc3)
	polyWallet4, _ := this.polySdk.OpenWallet(WalletAcc4)
	polyWallet = append(polyWallet, polyWallet1, polyWallet2, polyWallet3, polyWallet4)
	for _, wallet := range polyWallet {

		acc, err := wallet.GetDefaultAccount([]byte(WalletPwd))
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}

	return accs, nil
}

/*
序列化EOS块头
*/
// func SerializationTrans(hdr *eos.SignedBlockHeader) []byte {

// }

/*
注册Relayer
参数:地址列表
*/
func (this *RegisterEOS) RegisterRelayer(addresses []string) (uint64, error) {
	var err error
	adds := make([]common.Address, len(addresses))
	for i, v := range addresses {
		adds[i], err = common.AddressFromBase58(v)
		if err != nil {
			return 0, fmt.Errorf("no%d address decode failed: %v", i, err)
		}
	}
	acc, err := this.getAccount()
	if err != nil {
		return 0, fmt.Errorf("get defaultAccount error: %v", err)
	}
	txHash, err := this.polySdk.Native.Rm.RegisterRelayer(adds, acc)
	if err != nil {
		return 0, fmt.Errorf("register Relayer error: %v", err)
	}
	WaitPolyTx(txHash, this.polySdk)
	event, err := this.polySdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		return 0, fmt.Errorf("get register relayer smart contract event error: %v", err)
	}
	var id uint64
	for _, e := range event.Notify {
		states := e.States.([]interface{})
		if states[0].(string) == "putRelayerApply" {
			id = uint64(states[1].(float64))
		}
	}
	fmt.Printf("successful to register%v, and id is %d: txhash: %s\n", addresses, id, txHash.ToHexString())
	return id, nil
}

/*
同意注册Relayer
参数:注册Relayer事件的id号
*/
func (this *RegisterEOS) ApproveRegisterRelayer(id uint64) error {
	acc, err := this.getAccount()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}
	txHash, err := this.polySdk.Native.Rm.ApproveRegisterRelayer(id, acc)
	if err != nil {
		return fmt.Errorf("approve register relayer error: %v", err)
	}
	WaitPolyTx(txHash, this.polySdk)
	fmt.Printf("successful to approve registration id %d: txhash: %s\n", id, txHash.ToHexString())
	return nil
}

/*
注册侧链
参数:

	chainId 侧链Id
	router	侧链Router
	name	侧链名
	num		块等待时间
	CMCC	合约地址
*/
func (this *RegisterEOS) RegisterSideChain(chainId, router, num uint64, name, cmcc string) error {
	cmcc = strings.TrimPrefix(cmcc, "0x")
	var cmccAddr []byte
	var err error
	if cmcc == "" {
		cmccAddr = []byte{}
	} else {
		cmccAddr = []byte(cmcc)
	}

	acc, err := this.getAccount()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}

	txHash, err := this.polySdk.Native.Scm.RegisterSideChain(acc.Address, chainId, router, name, num, cmccAddr, acc)
	if err != nil {
		return fmt.Errorf("registerSideChain error:%v", txHash)
	}
	WaitPolyTx(txHash, this.polySdk)
	fmt.Printf("successful to register side chain: txhash: %s\n", txHash.ToHexString())

	return this.ApproveRegisterSideChain(chainId)
}

/*
同意注册侧链
参数：

	chainId 侧链Id
	signer	poly同步节点账户
*/
func (this *RegisterEOS) ApproveRegisterSideChain(chainId uint64) error {
	accs, err := this.getAccounts()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}
	for i := 0; i < len(accs)-1; i++ {
		txHash, err := this.polySdk.Native.Scm.ApproveRegisterSideChain(chainId, accs[i])
		if err != nil {
			return fmt.Errorf("ApproveRegisterSideChain failed: %v", err)
		}
		WaitPolyTx(txHash, this.polySdk)
		fmt.Printf("第%d次, successful to approve: ( acc: %s, txhash: %s, chain-id: %d )\n", i,
			accs[i].Address.ToHexString(), txHash.ToHexString(), chainId)
	}

	return nil
}

/*
退出侧链
*/
func (this *RegisterEOS) QuitSideChain(chainID uint64) error {
	acc, err := this.getAccount()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}
	fmt.Printf("acc.Address.ToBase58(): %v\n", acc.Address.ToBase58())
	fmt.Printf("acc.Address.ToHexString(): %v\n", acc.Address.ToHexString())
	polySdk := this.polySdk

	txHash, err := polySdk.Native.Scm.QuitSideChain(chainID, acc)
	if err != nil {
		return fmt.Errorf("QuitSideChain failed: %v", err)
	}
	fmt.Printf("successful to quit chain: ( acc: %s, txhash: %s, chain-id: %d )\n",
		acc.Address.ToBase58(), txHash.ToHexString(), chainID)
	WaitPolyTx(txHash, this.polySdk)
	return this.ApproveQuitSideChain(chainID)
}

/*
同意注册侧链
*/
func (this *RegisterEOS) ApproveQuitSideChain(chainID uint64) error {

	accs, err := this.getAccounts()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}
	polySdk := this.polySdk

	for i := 0; i < len(accs)-1; i++ {
		// acc 第二个钱包
		txhash, err := polySdk.Native.Scm.ApproveQuitSideChain(chainID, accs[i])
		if err != nil {
			return fmt.Errorf("ApproveQuitSideChain failed: %v", err)
		}
		fmt.Printf("index %d, successful to approve quit chain: ( acc: %s, txhash: %s, chain-id: %d )\n",
			i, accs[i].Address.ToHexString(), txhash.ToHexString(), chainID)
	}
	return nil
}

/*
提交侧链创世区块头到Poly
参数:

	chainId	侧链Id
	eos_hdr_height	EOS侧链创世区块头高度
	eosSdk	eos_go_Sdk API
*/
func (this *RegisterEOS) SyncGenesisHeaderToPoly(chainID uint64, eos_hdr_height uint32, eosSdk *eos.API) error {
	blockResp, err := tools.GetEOSBlockByNum(eosSdk, eos_hdr_height)
	if err != nil {
		return fmt.Errorf("eos:GetBlockNum error: %v", err)
	}
	hdr := blockResp.SignedBlockHeader
	raw, err := eos.MarshalBinary(hdr)
	if err != nil {
		return fmt.Errorf("eos MarshalBinary error: %v", err)
	}
	accs, err := this.getAccounts()
	if err != nil {
		return fmt.Errorf("get defaultAccount error: %v", err)
	}
	// res, err := this.polySdk.Native.Hs.SyncGenesisHeader(chainID, raw, accs)
	res, err := this.polySdk.Native.Hs.SyncGenesisHeader(chainID, raw, accs)
	if err != nil {
		return fmt.Errorf("poly:faild send to Poly SyncGenesisHeader error is: %v", err)
	}
	fmt.Printf("SyncGenesisHeader success, the return is: %v\n", res)
	return nil
}

/*
提交Poly创世区块头到跨链管理合约
参数:

	height	poly链高度
*/
func (this *RegisterEOS) SyncPolyHdrToEOS(height uint32) error {

	hdr, err := this.polySdk.GetHeaderByHeight(height)
	if err != nil {
		return fmt.Errorf("poly:GetHeaderByHeight error: %v", err)
	}

	// fmt.Printf("polySdk:GetHeaderByHeight the hdr is:%v\n", hdr)

	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(hdr.ConsensusPayload, blkInfo); err != nil {
		return fmt.Errorf("commitHeader - unmarshal blockInfo error : %s", err)
	}
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
	}

	sender := this.selectSender()
	if sender.commitInitgenblock(hdr, publickeys) {
		hash := hdr.Hash()
		fmt.Printf("successful to sync poly header (hash: %s, height: %d) to EOS: \n", hash.ToHexString(), hdr.Height)
	}

	return nil
}

func (this *RegisterEOS) selectSender() *EOSSender {

	fmt.Printf("sender :%v", this.senders)
	seed := rand.New(rand.NewSource(time.Now().Unix()))
	num := seed.Intn(len(this.senders))
	fmt.Printf("seed num is:%d", num)

	return this.senders[num]
}

func (this *EOSSender) commitInitgenblock(header *polytypes.Header, pubkList []byte) bool {
	headerdata := header.GetMessage()
	headerHash := header.Hash()

	basicBk := &contract.Basics{
		Caller:     eos.AccountName(this.acc.AccountName),
		Contract:   eos.AccountName(this.config.EOSConfig.ContractAddress),
		ActionName: eos.ActionName(contract.INITGENBLOCK),
		Per:        "active",
	}
	txDataBK := &contract.InputInitgenblock{
		RawHeader:  headerdata,
		PubKeyList: pubkList,
	}
	txDataByte, err := json.Marshal(txDataBK)
	if err != nil {
		fmt.Printf("commit init - err:" + err.Error())
		return false
	}

	info := &EOSTxInfo{
		basics:     basicBk,
		txData:     txDataByte,
		prkey:      this.acc.Ks.String(),
		polyTxHash: headerHash.ToHexString(),
	}
	err = this.sendTxToEOS(info)
	if err != nil {
		fmt.Printf("send Tx to EOS error :%v", err)
	}
	return true
}

func (this *EOSSender) sendTxToEOS(info *EOSTxInfo) error {
	basics := info.basics
	var ctx context.Context = context.Background()
	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, info.prkey) // 导入私钥
	if err != nil {
		return fmt.Errorf("import private key: %v", err)
	}
	this.eosClient.SetSigner(keyBag) // 设置签名

	txOpts := &eos.TxOptions{}
	// 将HeadBlockID与ChainID填充到txOpts
	if err := txOpts.FillFromChain(ctx, this.eosClient); err != nil {
		return fmt.Errorf("filling tx opts:%v", err)
	}
	// 构建交易
	var tx *eos.Transaction

	var input contract.InputInitgenblock
	json.Unmarshal(info.txData, &input)
	// fmt.Printf("json Unmarshal input, input is:%v", input)
	// fmt.Printf("method:Initgenblock\nRawHeader:%v\nPubKeyList:%v\n", input.RawHeader, input.PubKeyList)
	tx = eos.NewTransaction([]*eos.Action{basics.Initgenblock(input.RawHeader, input.PubKeyList)}, txOpts)

	// 签名并打包交易
	signedTx, packedTx, err := this.eosClient.SignTransaction(ctx, tx, txOpts.ChainID, eos.CompressionNone)
	if err != nil {
		return fmt.Errorf("sign transaction: %v", err)
	}
	content, err := json.MarshalIndent(signedTx, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshalling transaction: %v", err)
	}
	fmt.Printf("signedTx:%v\n", string(content)) // TODO调试输出后续删除
	// push打包后的签名交易
	response, err := this.eosClient.PushTransaction(ctx, packedTx)
	if err != nil {
		return fmt.Errorf("push transaction:%v", err)
	}
	fmt.Printf("PushTransaction success, txId:%s\n", hex.EncodeToString(response.Processed.ID))
	fmt.Printf("PushTransaction success, transaction ID :%s\n", response.TransactionID)
	return nil
}
