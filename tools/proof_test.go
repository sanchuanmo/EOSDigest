package tools

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/db"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/eos_relayer/proof"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	"github.com/qqtou/eos-go"
)

var ConfigPath string = "../config_eos.json"
var LogDir string = "../Log/"
var StartHeight uint64 = 0
var PolyStartHeight uint64 = 0
var StartForceHeight uint64 = 0
var eosUrl = "http://0.0.0.0:8888"
var HEITHT = 8651667
var testDataHeight uint32 = 18541467

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
测试eos-go-sdk Id的Hash算法是否稳定
*/
func TestBlockID(t *testing.T) {
	eosSdk := getEOSServer()
	var res [][]byte
	for i := 0; i < 10; i++ {
		signedHeader, _ := GetEOSHeaderByNum(eosSdk, testDataHeight)
		cereal, _ := signedHeader.BlockID()
		res = append(res, cereal)
	}
	for i := 0; i < 9; i++ {
		for j := i + 1; j < 10; j++ {
			fmt.Printf("%d 和 %d 相比，是否相等:%v\n", i, j, bytes.Equal(res[i], res[j]))
		}
	}
}

/*
eos-go-sdk自带的二进制编码无法处理结构体参数的指针问题
*/

func TestGetHeader(t *testing.T) {
	eosSdk := getEOSServer()
	var res [][]byte
	for i := 0; i < 10; i++ {
		signedHeader, _ := GetEOSHeaderByNum(eosSdk, testDataHeight)
		cereal, _ := eos.MarshalBinary(signedHeader)
		res = append(res, cereal)
	}
	for i := 0; i < 9; i++ {
		for j := i + 1; j < 10; j++ {
			fmt.Printf("%d 和 %d 相比，是否相等:%v\n", i, j, bytes.Equal(res[i], res[j]))
		}
	}
}

/*
测试序列化内容是否包含指针
*/
func TestGetTran(t *testing.T) {
	eosSdk := getEOSServer()
	var res [][]byte

	for i := 0; i < 10; i++ {
		blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
		tras1 := blockRespR1.Transactions

		// h := sha256.New()
		// _, _ = h.Write(cereal)
		// hashed := h.Sum(nil)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, tras1[0].Status)
		binary.Write(buf, binary.LittleEndian, tras1[0].CPUUsageMicroSeconds)
		binary.Write(buf, binary.LittleEndian, tras1[0].NetUsageWords)
		temp, _ := tras1[0].Transaction.MarshalJSON()
		binary.Write(buf, binary.LittleEndian, temp)
		res = append(res, buf.Bytes())
	}
	for i := 0; i < 9; i++ {
		for j := i + 1; j < 10; j++ {
			fmt.Printf("%d 和 %d 相比，是否相等:%v\n", i, j, bytes.Equal(res[i], res[j]))
		}
	}
}

/*
测试叶子节点序列化逻辑
*/
func TestCaulLeaf(t *testing.T) {
	eosSdk := getEOSServer()
	var res [][]byte
	for i := 0; i < 10; i++ {
		blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
		tras1 := blockRespR1.Transactions

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, tras1[0].Status)
		binary.Write(buf, binary.LittleEndian, tras1[0].CPUUsageMicroSeconds)
		binary.Write(buf, binary.LittleEndian, tras1[0].NetUsageWords)
		temp, _ := tras1[0].Transaction.MarshalJSON()
		binary.Write(buf, binary.LittleEndian, temp)
		bufByte := buf.Bytes()
		h := sha256.New()
		_, _ = h.Write(bufByte)
		hashed := h.Sum(nil)
		res = append(res, hashed)
	}
	for i := 0; i < 9; i++ {
		for j := i + 1; j < 10; j++ {
			fmt.Printf("%d 和 %d 相比，是否相等:%v\n", i, j, bytes.Equal(res[i], res[j]))
		}
	}
	fmt.Println(len(res[0]))
}

/*
验证EOS源码左右标志是否可以判断左右
*/
func TestSignChange(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tras1 := blockRespR1.Transactions
	tra0Ser := proof.SerializationTrans(tras1[0])
	tra1Ser := proof.SerializationTrans(tras1[1])
	tra0Ser[0] &= byte(proof.LeftSign)
	tra1Ser[0] |= byte(proof.RightSign)
	fmt.Printf("JudgeLeft(tra0Ser): %v\n", proof.JudgeLeft(tra0Ser))
	fmt.Printf("Judgeright(tra1Ser): %v\n", proof.Judgeright(tra1Ser))
}

func TestFunc(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tras1 := blockRespR1.Transactions
	for i, tra := range tras1 {
		trai := proof.SerializationTrans(tra)
		fmt.Printf("trai[%d]: %v\n", i, trai)
		trai = proof.CalculateHash(trai)
		fmt.Printf("trai[%d] hashed: %v\n", i, trai)
		fmt.Printf("proof.SignToLeft(tra0Ser): %v\n", proof.SignToLeft(trai))
		temp := trai
		temp[0] &= byte(proof.LeftSign)
		fmt.Printf("nomal left: %v\n", temp)
		fmt.Printf("proof.SignToRight(tra1Ser): %v\n", proof.SignToRight(trai))
		temp02 := trai
		temp02[0] |= byte(proof.RightSign)
		fmt.Printf("nomal right: %v\n", temp02)
	}

}

/*
测试获取默克尔树
*/
func TestMerklePath(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tras1 := blockRespR1.Transactions
	fmt.Println(len(tras1))
	tree, _ := proof.NewTree(tras1)
	fmt.Println(len(tree.Leafs))
	for i := 0; i < len(tree.Leafs); i++ {
		fmt.Printf("the %d index leaf hash is %v\n", i, tree.Leafs[i].Hash)
	}
	fmt.Printf("block txRoot is :%v\n", hex.EncodeToString(blockRespR1.TransactionMRoot))
	fmt.Println(hex.EncodeToString(tree.GetMerkleRoot()))

	path, _, err := tree.GetMerklePath(proof.SerializationTrans(tras1[0]))
	if err != nil {
		panic("Utils: GetMerklePath error" + err.Error())
	}
	fmt.Printf("Merkle Proof path len is: %v\n", len(path))
	fmt.Printf("Merkle Proof path is %v\n", path)
	for i, pa := range path {
		if proof.JudgeLeft(pa) {
			fmt.Printf("the index %d node is left\n", i)
		}
		if proof.Judgeright(pa) {
			fmt.Printf("the index %d node is right\n", i)
		}
		fmt.Printf("the index %d node len is:%d\n", i, len(pa))
	}
}

/*
验证自身构建默克尔树的路径有效性，可后续优化，优化掉state
已优化掉state，重写了EOSproof结构体的序列化
*/
func TestVerifyLeaf(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tras1 := blockRespR1.Transactions
	tree, _ := proof.NewTree(tras1)
	fmt.Printf("tree.Leafs: %v\n", tree.Leafs)
	fmt.Printf("tree.Root.Hash: %v\n", tree.Root.Hash)
	fmt.Printf("tree.GetMerkleRoot(): %v\n", tree.GetMerkleRoot())
	// tree.ToString()
	path, _, err := tree.GetMerklePath(proof.SerializationTrans(tras1[0]))
	if err != nil {
		panic("GetMerklePath error:" + err.Error())
	}
	for i, pa := range path {
		fmt.Printf("the index %d pa is:%v\n", i, pa)
	}

	// Tree ToString()展示所有节点
	// Path 展示Path的节点
	temp := proof.SerializationTrans(tras1[0])
	hashed := proof.CalculateHash(temp)
	fmt.Printf("the leaf Hash is %v\n", hashed)
	rootCal := proof.VerifyLeaf(path, hashed)
	fmt.Printf("MerkleRoot is %v\n", tree.GetMerkleRoot())
	fmt.Printf("Caculate Root is %v\n", rootCal)
	fmt.Printf("Caculate Root Sign Left is %v\n", proof.SignToLeft(rootCal))
	fmt.Printf("Caculate Root Sign Right is %v\n", proof.SignToRight(rootCal))
	fmt.Printf("Calculate is:%v", bytes.Equal(rootCal, tree.GetMerkleRoot()))
}

/*
验证Proof的序列化和反序列化
*/
func TestProofSD(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tras1 := blockRespR1.Transactions
	tree, _ := proof.NewTree(tras1)
	// path, state, _ := tree.GetMerklePath(proof.SerializationTrans(tras1[0]))
	proof, err := GetEOSProof(tree, proof.SerializationTrans(tras1[0]))
	if err != nil {
		panic("GetEOSProof Error:" + err.Error())
	}
	sink := common.NewZeroCopySink(nil)
	proof.Serialization(sink)
	// fmt.Printf("Serialization proofByte is: %v\n", sink.Bytes())
	// fmt.Printf("Serialization Proof len is: %d", len(sink.Bytes()))
	proof2 := new(EOSProof)
	err = proof2.Deserialization(common.NewZeroCopySource(sink.Bytes()))
	if err != nil {
		fmt.Printf("Deserialization error: %v", err)
	}
	fmt.Printf("Deserialization proof is: %v\n", proof2)

	fmt.Printf("proof2.leaf: %v\n, len(leaf) is: %d", proof2.leaf, len(proof2.leaf))
	fmt.Printf("proof2.path: %v\n len(path) is: %d", proof2.path, len(proof2.path))
}

/*
验证自行构造的默克尔树根节点和EOS块构造的默克尔交易根节点是否相同
*/
func TestEqualTree(t *testing.T) {
	eosSdk := getEOSServer()
	blockRespR1, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	tree, _ := proof.NewTree(blockRespR1.Transactions)
	caculateRoot := hex.EncodeToString(tree.GetMerkleRoot())
	blockRoot := blockRespR1.TransactionMRoot.String()

	fmt.Printf("the caculateRoot is: %v\n,the block transactionRoot is: %v\n,Is equal:%v\n", caculateRoot, blockRoot, blockRoot == caculateRoot)

}

func TestProofNil(t *testing.T) {
	eosSdk := getEOSServer()
	block, _ := GetEOSBlockByNum(eosSdk, testDataHeight+1)
	tree, _ := proof.NewTree(block.Transactions)
	fmt.Printf("tree: %v\n", tree)
	fmt.Printf("tree.Root: %v\n", tree.Root)
}

func TestSetRoot(t *testing.T) {
	eosSdk := getEOSServer()
	block, _ := GetEOSBlockByNum(eosSdk, testDataHeight+1)
	fmt.Printf("block.TransactionMRoot: %v\n", block.TransactionMRoot)
	empty := ""
	hash := proof.CalculateHash([]byte(empty))
	fmt.Printf("hex.EncodeToString(hash): %v\n", hex.EncodeToString(hash))
	empty2 := make([]byte, 32)
	fmt.Printf("hex.EncodeToString(empty2): %v\n", hex.EncodeToString(empty2))
}

func TestBlockRootChange(t *testing.T) {
	eosSdk := getEOSServer()
	block, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	fmt.Printf("be change root block.TransactionMRoot: %v\n", block.TransactionMRoot)
	fmt.Println(block.BlockID())
	tree, _ := proof.NewTree(block.Transactions)
	block.TransactionMRoot = eos.Checksum256(tree.GetMerkleRoot())
	rawHdr, _ := eos.MarshalBinary(block.SignedBlockHeader)
	var hdrRe *eos.SignedBlockHeader
	err := eos.UnmarshalBinary(rawHdr, &hdrRe)
	if err != nil {
		panic("error UnmarshalBinary" + err.Error())
	}
	fmt.Printf("af change root block.TransactionMRoot: %v\n", hdrRe.TransactionMRoot)
	fmt.Printf("is Equal hdrRe.TransactionMRoot and tree.Root%v", hdrRe.TransactionMRoot.String() == eos.Checksum256(tree.GetMerkleRoot()).String())
	fmt.Printf("Bytes Equal hdrRe.TransactionMRoot and tree.Root: %v", bytes.Equal(hdrRe.TransactionMRoot, tree.GetMerkleRoot()))
	// fmt.Println(block.BlockID())

}

func TestSeDeProof(t *testing.T) {
	eosSdk := getEOSServer()
	block, _ := GetEOSBlockByNum(eosSdk, testDataHeight)
	hdrBytes, _ := eos.MarshalBinary(block.SignedBlockHeader)
	var newHdr *eos.SignedBlockHeader
	err := eos.UnmarshalBinary(hdrBytes, &newHdr)
	if err != nil {
		fmt.Printf("unmarshalbinary eos error: %v", err)
	} else {
		fmt.Printf("success")
	}
}

func TestCal2(t *testing.T) {
	eosSdk := getEOSServer()
	node := "2d21b3a3ccedd05977b5ad45b854dfefa2c0b4b4b539101cfc8a6549ea106da8"
	block, _ := GetEOSBlockByNum(eosSdk, uint32(18193105))
	fmt.Printf("block.Transactions[0].Transaction.ID: %v\n", block.Transactions[0].Transaction.ID)
	nodeByte, _ := hex.DecodeString(node)
	root := proof.CalculateNodeHash(proof.SignToLeft(nodeByte), proof.SignToRight(nodeByte))
	fmt.Printf("root: %v\n", hex.EncodeToString(root))
}

func TestCalRoot(t *testing.T) {
	eosSdk := getEOSServer()
	// var ctx context.Context = context.Background()

	block, _ := GetEOSBlockByNum(eosSdk, uint32(18561780)) //仅有一个transaction
	// chainInfo, _ := eosSdk.GetInfo(ctx)
	// chainID := []byte(chainInfo.ChainID)
	var transactions [][]byte
	for _, tr := range block.Transactions {
		// transaction := tr.Transaction.Packed.Transaction
		// 序列化
		// trByte, _ := eos.MarshalBinary(transaction)
		trByte := SerializationTrx(&tr)
		fmt.Printf("[]byte(trByte): %v\n", []byte(trByte))
		//
		// transactions = append(transactions, buildTxDigest(chainID, []byte(trByte)))
		transactions = append(transactions, proof.CalculateHash(trByte))
	}
	// 补2
	transactions = append(transactions, transactions[len(transactions)-1])

	fmt.Printf("len(transactions): %v\n", len(transactions))
	fmt.Printf("proof.SignToLeft(transactions[0]): %v\n", proof.SignToLeft(proof.CalculateHash(transactions[0])))
	fmt.Printf("proof.SignToRight(transactions[1]): %v\n", proof.SignToRight(proof.CalculateHash(transactions[1])))
	calRoot := proof.CalculateNodeHash(proof.SignToLeft(proof.CalculateHash(transactions[0])), proof.SignToRight(proof.CalculateHash(transactions[1])))

	fmt.Printf("root: %v\n", hex.EncodeToString(calRoot))
	fmt.Printf("block.TransactionMRoot: %v\n", block.TransactionMRoot.String())
}

// func CalRoot(transactions [][]byte) {
// 	if len(transactions)%2 == 1 {
// 		transactions = append(transactions, transactions[len(transactions)-1])
// 	}

// 	for i := 0; i < len(transactions)%2; i++ {

// 	}
// }

// 构造transaction Digest
func buildTxDigest(chainID, transactionByte []byte) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, chainID)
	binary.Write(buf, binary.LittleEndian, transactionByte)
	return buf.Bytes()
}

func SerializationTrx(transaction *eos.TransactionReceipt) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, transaction.Status)
	binary.Write(buf, binary.LittleEndian, transaction.CPUUsageMicroSeconds)
	binary.Write(buf, binary.LittleEndian, transaction.NetUsageWords)
	binary.Write(buf, binary.LittleEndian, []byte(transaction.Transaction.Packed.PackedTransaction))
	return buf.Bytes()
}
