package tools

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"

	sdk "github.com/polynetwork/poly-go-sdk"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ontio/ontology-crypto/ec"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology-crypto/sm2"
	"github.com/polynetwork/eos_relayer/proof"
	"github.com/polynetwork/poly/common"
	eos "github.com/qqtou/eos-go"
)

const (
	HashLength          = 32
	CUREPOCHSTARTHEIGHT = "curEpochStartHeight"
	COMKEEPERSPKBYTES   = "conKeepersPkBytes"
	CROSSCONTRACTTABLE  = "polyglobal"
	TRANSFEE            = 10000
)

type EOSProof struct {
	path [][]byte
	leaf []byte
}

func (this *EOSProof) GetPath() [][]byte {
	return this.path
}

func (this *EOSProof) GetLeaf() []byte {
	return this.leaf
}

func (this *EOSProof) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBytes(this.leaf)
	for i := 0; i < len(this.path); i++ {
		sink.WriteBytes(this.path[i])
	}
}

func (this *EOSProof) Deserialization(data []byte) error {
	source := common.NewZeroCopySource(data)
	n := source.Len()
	if (n % 32) != 0 {
		return fmt.Errorf("Deserialization error : len is illegal")
	}
	var path [][]byte
	leaf, eof := source.NextBytes(32)
	if eof {
		return fmt.Errorf("Waiting deserialize leaf error")
	}
	for i := 0; i < int(n/32)-1; i++ {
		pa, eof := source.NextBytes(32)
		if eof {
			return fmt.Errorf("Waiting deserialize %d path error", i)
		}
		path = append(path, pa)
	}
	this.leaf = leaf
	this.path = path
	return nil
}

func GetEOSChainId(eosSdk *eos.API) (string, error) {
	var ctx context.Context = context.Background()
	chainInfo, err := eosSdk.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("GetEOSChainId: err:%s", err)
	}
	return chainInfo.ChainID.String(), nil
}

func GetEOSNodeHeight(eosSdk *eos.API) (uint64, error) {
	var ctx context.Context = context.Background()
	infoResp, err := eosSdk.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("GetEOSNodeHeight: err:%s", err)
	}
	return uint64(infoResp.LastIrreversibleBlockNum), nil

}

/*
替换目前交易默克尔根TransactionMRoot，
修改BlockID
// 优化一个有且仅计算默克尔树根的算法
*/
func GetEOSBlockByNum(eosSdk *eos.API, height uint32) (*eos.BlockResp, error) {
	var ctx context.Context = context.Background()
	var infoResp *eos.BlockResp
	infoResp, err := eosSdk.GetBlockByNum(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("GetEOSBlockByNum: err:%v", err)
	}

	tree, err := proof.NewTree(infoResp.Transactions)
	// Proof start 修改当前TransactionMRoot
	infoResp.TransactionMRoot = eos.Checksum256(tree.GetMerkleRoot())
	// 更新BlockID
	newBlockID, err := infoResp.BlockID()
	infoResp.ID = newBlockID

	return infoResp, nil
}

func GetEOSTraceBlockByNum(eosSdk *eos.API, height uint32) (*eos.BlockTraceResp, error) {
	var ctx context.Context = context.Background()
	infoResp, err := eosSdk.GetBlockTraceByNum(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("GetEOSTraceBlockByNum: err:%s", err)
	}

	return infoResp, nil
}

func GetEOSDeTraceData(eosSdk *eos.API, contractName eos.AccountName, actionName eos.Name, data string) (map[string]interface{}, error) {
	var ctx context.Context = context.Background()
	dataByte, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("GetEOSDeTraceData: DecodeString data err: %s", err)
	}
	resp, err := eosSdk.ABIBinToJSON(ctx, contractName, actionName, dataByte)
	if err != nil {
		return nil, fmt.Errorf("GetEOSDeTraceData: Get ABIBinToJSON err:%s", err)
	}
	return resp, nil
}

func GetEOSHeaderByNum(eosSdk *eos.API, height uint32) (*eos.SignedBlockHeader, error) {
	var infoResp *eos.BlockResp
	infoResp, err := GetEOSBlockByNum(eosSdk, height)
	if err != nil {
		return nil, fmt.Errorf("GetEOSHeaderByNum: err:%v", err)
	}
	return &infoResp.SignedBlockHeader, nil
}

func GetTableRowsMap(eosSdk *eos.API, request eos.GetTableRowsRequest) ([]map[string]interface{}, error) {
	var ctx context.Context = context.Background()
	response, err := eosSdk.GetTableRows(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("get table rows:%s", err)
	}
	rows := string(response.Rows)
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(rows), &data); err != nil {
		return nil, fmt.Errorf("json unmarshal: %s", err)
	}

	return data, nil
}

func GetEOSStartHeight(eosSdk *eos.API, contractAddress, table string) (uint32, error) {

	var request = eos.GetTableRowsRequest{
		JSON:    true,
		Code:    contractAddress,
		Scope:   contractAddress,
		Table:   table,
		Reverse: false,
	}
	data, err := GetTableRowsMap(eosSdk, request)
	if err != nil {
		return 0, err
	}

	height := data[0][CUREPOCHSTARTHEIGHT].(float64)

	return uint32(height), nil
}

func GetEOSRawKeepers(eosSdk *eos.API, contractAddress, table string) ([]byte, error) {
	var request = eos.GetTableRowsRequest{
		JSON:    true,
		Code:    contractAddress,
		Scope:   contractAddress,
		Table:   table,
		Reverse: false,
	}
	data, err := GetTableRowsMap(eosSdk, request)
	if err != nil {
		return nil, err
	}
	fmt.Printf("data:%v\n", data)

	rawKeepers := data[0][COMKEEPERSPKBYTES].([]interface{})
	rawKeepersBytes := TransInterfacesToBytes(rawKeepers)
	return rawKeepersBytes, nil
}

func GetEOSProof(merkleTree *proof.MerkleTree, node []byte) (*EOSProof, error) {
	path, nodeHash, err := merkleTree.GetMerklePath(node)
	if err != nil {
		return nil, err
	}
	proof := &EOSProof{
		leaf: nodeHash,
		path: path,
	}
	return proof, nil
}

func EOSMarshalBinary(v interface{}) ([]byte, error) {
	res, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("JSON:Marshal error: %v", err)
	}
	return res, nil
}

func EOSSha256(cereal []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(cereal)

	hashed := h.Sum(nil)
	return hashed
}

func GetNoCompresskey(key keypair.PublicKey) []byte {
	var buf bytes.Buffer
	switch t := key.(type) {
	case *ec.PublicKey:
		switch t.Algorithm {
		case ec.ECDSA:
			// Take P-256 as a special case
			if t.Params().Name == elliptic.P256().Params().Name {
				return ec.EncodePublicKey(t.PublicKey, false)
			}
			buf.WriteByte(byte(0x12))
		case ec.SM2:
			buf.WriteByte(byte(0x13))
		}
		label, err := GetCurveLabel(t.Curve.Params().Name)
		if err != nil {
			panic(err)
		}
		buf.WriteByte(label)
		buf.Write(ec.EncodePublicKey(t.PublicKey, false))
	case ed25519.PublicKey:
		panic("err")
	default:
		panic("err")
	}
	return buf.Bytes()
}

func GetCurveLabel(name string) (byte, error) {
	switch strings.ToUpper(name) {
	case strings.ToUpper(elliptic.P224().Params().Name):
		return 1, nil
	case strings.ToUpper(elliptic.P256().Params().Name):
		return 2, nil
	case strings.ToUpper(elliptic.P384().Params().Name):
		return 3, nil
	case strings.ToUpper(elliptic.P521().Params().Name):
		return 4, nil
	case strings.ToUpper(sm2.SM2P256V1().Params().Name):
		return 20, nil
	case strings.ToUpper(btcec.S256().Name):
		return 5, nil
	default:
		panic("err")
	}
}

func ParseAuditpath(path []byte) ([]byte, []byte, [][32]byte, error) {
	source := common.NewZeroCopySource(path)

	value, eof := source.NextVarBytes()
	if eof {
		return nil, nil, nil, nil
	}
	size := int((source.Size() - source.Pos()) / common.UINT256_SIZE)
	pos := make([]byte, 0)
	hashs := make([][32]byte, 0)
	for i := 0; i < size; i++ {
		f, eof := source.NextByte()
		if eof {
			return nil, nil, nil, nil
		}
		pos = append(pos, f)

		v, eof := source.NextHash()
		if eof {
			return nil, nil, nil, nil
		}
		var onehash [32]byte
		copy(onehash[:], (v.ToArray())[0:32])
		hashs = append(hashs, onehash)
	}

	return value, pos, hashs, nil
}

func ConvertToEosCompatible(sig []byte) ([]byte, error) {
	s, err := signature.Deserialize(sig)
	if err != nil {
		return nil, err
	}

	t, ok := s.Value.([]byte)
	if !ok {
		return nil, errors.New("invalid signature type")
	}

	if len(t) != 65 {
		return nil, errors.New("invalid signature length")
	}

	v := t[0] - 27
	copy(t, t[1:])
	t[64] = v
	return t, nil
}

func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

func GetPolyStorageHeaderID(polySdk *sdk.PolySdk, height uint64, sideChainIdBytes [8]byte) (*eos.Checksum256, error) {

	contractAddress := autils.HeaderSyncContractAddress
	key := append(append([]byte(scom.MAIN_CHAIN), sideChainIdBytes[:]...), autils.GetUint64Bytes(height)...)
	result, err := polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil && result != nil {
		return nil, fmt.Errorf("findLastestHeight: GetStorage MAIN_CHAIN error" + err.Error())
	}
	var BlockID eos.Checksum256
	err = BlockID.UnmarshalJSON(result)
	return &BlockID, nil
}

func TransInterfacesToBytes(data []interface{}) []byte {
	var dataBytes []byte
	for _, da := range data {
		dataBytes = append(dataBytes, byte(da.(float64)))
	}
	return dataBytes
}

func FeeStrToInt(fee string) (uint64, error) {
	feevalue := strings.Split(fee, " ")[0]
	feeDou, err := strconv.ParseFloat(feevalue, 64)
	if err != nil {
		return 0, err
	}
	return uint64(feeDou * TRANSFEE), nil

}
