package eosdigest_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"testing"

	"github.com/sanchuanmo/eosdigest"

	eos "github.com/qqtou/eos-go"
)

var (
	restURL        = "http://101.43.228.245:8888"
	height  uint32 = 3138418
)

func GetEOSSdk() *eos.API {
	return eos.New(restURL)
}

func GetEOSNodeHeight(eosSdk *eos.API) (uint64, error) {
	var ctx context.Context = context.Background()
	infoResp, err := eosSdk.GetInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("GetEOSNodeHeight: err:%s", err)
	}
	return uint64(infoResp.LastIrreversibleBlockNum), nil

}

func GetEOSBlockByNum(eosSdk *eos.API, height uint32) (*eos.BlockResp, error) {
	var ctx context.Context = context.Background()
	var infoResp *eos.BlockResp
	infoResp, err := eosSdk.GetBlockByNum(ctx, height)
	if err != nil {
		return nil, fmt.Errorf("GetEOSBlockByNum: err:%v", err)
	}

	return infoResp, nil
}

func CalCuTreeDigest(trx eos.TransactionReceipt) string {
	status := trx.Status.String()
	compression := trx.Transaction.Packed.Compression.String()
	cpu_usage_us := uint(trx.CPUUsageMicroSeconds)
	net_usage_words := uint(trx.NetUsageWords)
	sig := trx.Transaction.Packed.Signatures
	signatures := eosdigest.NewStringVector()

	for _, value := range sig {
		signatures.Add(value.String())
	}
	packed_trxS := trx.Transaction.Packed.PackedTransaction
	packed_trx_Vector := eosdigest.NewUCharVector()

	for _, value := range packed_trxS {
		packed_trx_Vector.Add(value)
	}

	context_free_dataS := trx.Transaction.Packed.PackedContextFreeData
	context_free_dataVector := eosdigest.NewUCharVector()

	for _, value := range context_free_dataS {
		context_free_dataVector.Add(value)
	}

	return eosdigest.EosServializationTxDigest(status, cpu_usage_us, net_usage_words, compression, packed_trx_Vector, signatures, context_free_dataVector)
}

func TestMerkleTree(t *testing.T) {

	eosSdk := GetEOSSdk()
	blockRsp, _ := GetEOSBlockByNum(eosSdk, height)

	var leafs [][]byte

	for _, value := range blockRsp.Transactions {
		res := CalCuTreeDigest(value)
		hash, _ := hex.DecodeString(res)
		leafs = append(leafs, hash)
		fmt.Printf("res: %v\n", res)
	}

	temp := append(leafs[0], leafs[len(leafs)-1]...)

	enc := sha256.New()

	enc.Write(temp)

	fmt.Printf("enc.Sum(nil): %v\n", hex.EncodeToString(enc.Sum(nil)))

}

//e14869d99fdcf8944ac67d0f1f30fe9fb8ccaa609bb58caf8621a536fe4865ca
//e14869d99fdcf8944ac67d0f1f30fe9fb8ccaa609bb58caf8621a536fe4865ca
