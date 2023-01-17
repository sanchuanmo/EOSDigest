#### 1、

    sha256工具类实现

    实现细节：将结构体转化为bute[:]切片，然后进行Sha256加密Hash，最后生成加密后的Byte[:]



#### 2、
    配置文件测试账户
测试账户钱包信息
```sh
accountName janifer
publicKey EOS4wD4ystokQxSs681QLJ3sNjbojeUMbff2FVjx8Sjr8nmAeCgiA
privateKey 5JvypYCUAKnKpe2zDdtNrMcMR2XXPiHPUNhEEhT7oW8kNabd13G

accountName jarry
publicKey EOS4wD4ystokQxSs681QLJ3sNjbojeUMbff2FVjx8Sjr8nmAeCgiA
privateKey 5JvypYCUAKnKpe2zDdtNrMcMR2XXPiHPUNhEEhT7oW8kNabd13G
```
创建测试账户
```sh
    root@f8f822c1c047:/test_relayer# cleos system newaccount eosio janifer EOS4wD4ystokQxSs681QLJ3sNjbojeUMbff2FVjx8Sjr8nmAeCgiA --stake-net "100 SYS" --stake-cpu "100 SYS" --buy-ram "100 SYS" -p eosio@active
executed transaction: 55286b4ace3bb550561323715d7d1b0a4e836aa6adbffb8a6f77c2a7f076e2a8  344 bytes  410 us
#         eosio <= eosio::newaccount            {"creator":"eosio","name":"janifer","owner":{"threshold":1,"keys":[{"key":"EOS4wD4ystokQxSs681QLJ3sN...
#         eosio <= eosio::buyram                {"payer":"eosio","receiver":"janifer","quant":"100.0000 SYS"}
#         eosio <= eosio::delegatebw            {"from":"eosio","receiver":"janifer","stake_net_quantity":"100.0000 SYS","stake_cpu_quantity":"100.0...
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#     eosio.ram <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#  eosio.ramfee <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
#   eosio.stake <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
warning: transaction executed locally, but may not be confirmed by the network yet         ] 
root@f8f822c1c047:/test_relayer# cleos system newaccount eosio jarry EOS4wD4ystokQxSs681QLJ3sNjbojeUMbff2FVjx8Sjr8nmAeCgiA --stake-net "100 SYS" --stake-cpu "100 SYS" --buy-ram "100 SYS" -p eosio@active
executed transaction: 866e3f510068c877ab339dc499238b0dbfc1fb8d39672b06292ec34a4da75b32  344 bytes  497 us
#         eosio <= eosio::newaccount            {"creator":"eosio","name":"jarry","owner":{"threshold":1,"keys":[{"key":"EOS4wD4ystokQxSs681QLJ3sNjb...
#         eosio <= eosio::buyram                {"payer":"eosio","receiver":"jarry","quant":"100.0000 SYS"}
#         eosio <= eosio::delegatebw            {"from":"eosio","receiver":"jarry","stake_net_quantity":"100.0000 SYS","stake_cpu_quantity":"100.000...
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#     eosio.ram <= eosio.token::transfer        {"from":"eosio","to":"eosio.ram","quantity":"99.5000 SYS","memo":"buy ram"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#  eosio.ramfee <= eosio.token::transfer        {"from":"eosio","to":"eosio.ramfee","quantity":"0.5000 SYS","memo":"ram fee"}
#   eosio.token <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
#         eosio <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
#   eosio.stake <= eosio.token::transfer        {"from":"eosio","to":"eosio.stake","quantity":"200.0000 SYS","memo":"stake bandwidth"}
warning: transaction executed locally, but may not be confirmed by the network yet         ] 
```




#### Proof问题
```sh

event evt
type EthCrossChainManagerCrossChainEvent struct {
	Sender               common.Address
	TxId                 []byte
	ProxyOrAssetContract common.Address
	ToChainId            uint64
	ToContract           []byte
	Rawdata              []byte
	Raw                  types.Log // Blockchain specific contextual infos
}

index := big.NewInt(0)
index.SetBytes(evt.TxId)
txIndex: tools.EncodeBigInt(index)

key := crosstx.txIndex

keyBytes,err := ethMappingKeyAt(key,"01")

//如果报错， MappingKeyAt错误

proofKey := hexutil.Encode(keyBytes)
// 获取交易证明
proof, err := tools.GetProof(this.config.ETHConfig.RestURL, this.config.ETHConfig.ECCDContractAddress, proofKey, heightHex, this.restClient)

txHash, err := this.commitProof(uint32(height), proof, crosstx.value, crosstx.txId)
```




#### 3、研究js中如何通过接口调用智能合约，如何上链查询内存表数据

#### 4、确认GetMerkleProof的函数参数原理

#### 5、确认GetCrossStatesProof的函数参数原理


#### 6、辅助开发

```sh
param := &common2.ToMerkleValue{}

poly 结构体，难以更改
type ToMerkleValue struct {
	TxHash      []byte
	FromChainID uint64
	MakeTxParam *MakeTxParam
}

type MakeTxParam struct {
	TxHash              []byte
	CrossChainID        []byte
	FromContractAddress []byte
	ToChainID           uint64	目标链ID
	ToContractAddress   []byte	目标链合约地址
	Method              string	目标链方法
	Args                []byte
}


```


#### 7、eos-relayer-config.json

配置文件详情解析

目标合约地址集合，第一个变量``"0xD8aE73e06552E...bcAbf9277a1aac99"``为chainId
inbound 为支持目标链relayer中，入链的起始合约地址分别是什么
outbound 为支持起始链relayer中，出链的目标合约地址分别是哪些
```sh
"TargetContracts": [
    {
    "0xD8aE73e06552E...bcAbf9277a1aac99": { 
      "inbound": [6], 
      "outbound": [6]
    }
  }
]
```




未完成的有：
配置
1、relayer-config.json中的TargetContracts，BlockConfig，HeadersPerBatch，MonitorInterval
TargetContracts 目标链地址
BlockConfig 块监听阈值 （EOS侧链）
MonitorInterval 监听时间 
HeadersPerBatch 每次提交poly的块头信息量的阈值


起始链

1、筛选跨链事件-需要对应跨链管理合约，具体的函数名和参数尚不明确，暂时写了个测试函数名和测试参数的Demo
2、 tools.GetEOSProof()获取EOSProof证明

目标链

1、 获取跨链管理合约中的initGenesisBlock中存储的curEpochStartHeight
2、 从目标链跨链管理合约中获取数据GetCurEpochConPubKeyBytes

已开发到
目标链：发送交易到EOS的跨链管理合约

已完成：
起始链：

目标链：
初始化功能、监听Poly区块、


起始链relayer：
要开发的功能点列表：初始化功能、注册Relayer、同步Poly高度、监听EOS区块、处理区块、提交同步头、提交回滚、监听Poly高度、监听跨链事件、提交证明、监听提交结果
已经完成的功能点：初始化功能、注册Relayer、同步Poly高度、监听EOS区块、提交同步头、提交回滚、监听Poly高度、
正在开发的：处理区块、监听跨链事件、提交证明、监听提交结果
还未开始：
问题记录：
1、筛选跨链事件-需要对应跨链管理合约，具体的函数名和参数尚不明确，暂时写了个测试函数名和测试参数的Demo
2、 tools.GetEOSProof()获取EOSProof证明


目标链relayer：
要开发的功能点列表：初始化功能、同步EOS高度、监听Poly区块、判断同步周期、处理块中跨链事件、选择发送账号、提交跨链事件到EOS、选择Router、发送交易到EOS、提交同步周期pubkList
已经完成的功能点：初始化功能、监听Poly区块、
正在开发的：判断同步周期、处理块中跨链事件、同步EOS高度、选择发送账号、提交跨链事件到EOS、
还未开始：选择Router、发送交易到EOS、提交同步周期pubkList
问题记录：
1、 获取跨链管理合约中的initGenesisBlock中存储的curEpochStartHeight
2、 从目标链跨链管理合约中获取数据GetCurEpochConPubKeyBytes



测试：
this.contractAbi.Pack("verifyHeaderAndExecuteTx", rawAuditPath, headerData, rawProof, rawAnchor, sigs)

sender.commitDepositEventsWithHeader(hdr, param, hp, anchor, event.TxHash, auditpath)

func (*EthSender).commitDepositEventsWithHeader(header *polytypes.Header, param *common2.ToMerkleValue, headerProof string, anchorHeader *polytypes.Header, polyTxHash string, rawAuditPath []byte)

hdr = header
param = param
hp = headerProof
anchor = anchorHeader
event.TxHash = polyTxHash
auditpath = rawAuditPath

参数来源

rawAuditPath

```sh
//事件解析 (height) 
proof, err := this.polySdk.GetCrossStatesProof(hdr.Height-1, states[5].(string)
auditpath, _ := hex.DecodeString(proof.AuditPath)
```

headerData

```sh

headerData = header.GetMessage()
hdr = header
hdr, err := this.polySdk.GetHeaderByHeight(height + 1)


```

rawProof

```sh
rawProof, _ := hex.DecodeString(headerProof)
hp = headerProof
hp     string
当前周期中hp = nil
```

rawAnchor
```sh
var rawAnchor []byte
if anchorHeader != nil {
		rawAnchor = anchorHeader.GetMessage()
	}

anchor = anchorHeader

anchor *polytypes.Header 当前周期为nil

```
sigs

hdr = this.polySdk.GetHeaderByHeight(height + 1)



```sh
sigs       []byte
当 headerProof 不为空
for _, sig := range anchorHeader.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			// 转换格式
			newsig, _ := signature.ConvertToEthCompatible(temp)
			sigs = append(sigs, newsig...)
		}

同周期跨链，为空
hp = nil

for _, sig := range header.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			// 转换格式
			newsig, _ := signature.ConvertToEthCompatible(temp)
			sigs = append(sigs, newsig...)
		}


```


##### 获取本周期包含跨链交易的块
anchor *polytypes.Header
hp     string
为nil


hdr = header
param = param
hp = headerProof = nil
anchor = anchorHeader = nil
event.TxHash = polyTxHash
auditpath = rawAuditPath

rawAuditPath, headerData, rawProof, rawAnchor, sigs


###### rawAuditPath
```sh
auditpath = rawAuditPath
//事件解析 (height) 
proof, err := this.polySdk.GetCrossStatesProof(hdr.Height-1, states[5].(string)
auditpath, _ := hex.DecodeString(proof.AuditPath)
```

###### headerData
```sh
headerData = header.GetMessage()
hdr = header
hdr, err := this.polySdk.GetHeaderByHeight(height + 1)
```

###### rawProof = hp = nil

```sh
rawProof, _ := hex.DecodeString(headerProof)
hp = headerProof
hp     string
当前周期中hp = nil
```

###### rawAnchor = anchor = nil

```sh
var rawAnchor []byte
if anchorHeader != nil {
		rawAnchor = anchorHeader.GetMessage()
}
anchor = anchorHeader
anchor *polytypes.Header 当前周期为nil
```

###### sigs
if (hp == nil)
```sh
header = hdr
for _, sig := range header.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			// 转换格式
			newsig, _ := signature.ConvertToEthCompatible(temp)
			sigs = append(sigs, newsig...)
}
```

问题1 states[]内容分别是什么

#### 记录this.polySdk.GetMerkleProof(height+1, lastEpoch+1)怎么生成的逻辑

```sh
func (self *Ledger) GetMerkleProof(proofHeight, rootHeight uint32) ([]byte, error) {
	// 获取当前proofHeight的块的块hash 
	blockHash := self.ldgStore.GetBlockHash(proofHeight)
	//  判空
	if bytes.Equal(blockHash.ToArray(), common.UINT256_EMPTY.ToArray()) {
		return nil, fmt.Errorf("GetBlockHash(%d) empty", proofHeight)
	}
	// 执行 proofHeight + 1, blockHash.ToArray()[proofHeigght], rootHeight
	return self.ldgStore.GetMerkleProof(blockHash.ToArray(), proofHeight+1, rootHeight)
}
```


#### 记录this.polySdk.GetCrossStatesProof(hdr.Height-1, states[5].(string))

type PublicKey struct {
	Algorithm ECAlgorithm
	*ecdsa.PublicKey
}
type PrivateKey struct {
	PublicKey
	D *big.Int
}
type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}
type Curve interface {
	// Params returns the parameters for the curve.
	Params() *CurveParams
	// IsOnCurve reports whether the given (x,y) lies on the curve.
	IsOnCurve(x, y *big.Int) bool
	// Add returns the sum of (x1,y1) and (x2,y2)
	Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int)
	// Double returns 2*(x,y)
	Double(x1, y1 *big.Int) (x, y *big.Int)
	// ScalarMult returns k*(Bx,By) where k is a number in big-endian form.
	ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int)
	// ScalarBaseMult returns k*G, where G is the base point of the group
	// and k is an integer in big-endian form.
	ScalarBaseMult(k []byte) (x, y *big.Int)
}
type CurveParams struct {
	P       *big.Int // the order of the underlying field
	N       *big.Int // the order of the base point
	B       *big.Int // the constant of the curve equation
	Gx, Gy  *big.Int // (x,y) of the base point
	BitSize int      // the size of the underlying field
	Name    string   // the canonical name of the curve
}



poly 验签----底层实现poly公私钥，什么算法，什么椭圆，x，y值的加密方式等

#### poly内存表存储的值

###### 1、同步创世节点

```sh
//GENESIS_HEADER => the genesis header byte code
native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(storeBytes))

//HEADER_INDEX => the mapping of header hash and block header byte code, for querying block header by its hash
native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), blockHeader.Hash().Bytes()),cstates.GenRawStorageItem(storeBytes))

//CURRENT_HEADER_HEIGHT => current block height of side chain in poly relay chain
native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockHeader.Number.Uint64())),cstates.GenRawStorageItem(blockHeader.Hash().Bytes()))

//MAIN_CHAIN => the mapping of block height and block header hash, for querying block header hash by its height
native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(blockHeader.Number.Uint64())))

scom.NotifyPutHeader(native, chainID, blockHeader.Number.Uint64(), blockHeader.Hash().String())
```

###### 2、