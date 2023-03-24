package toolbox

import (
	"encoding/hex"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/polynetwork/eos_relayer/log"

	// "github.com/polynetwork/eos_relayer/manager"
	"github.com/polynetwork/poly/common"
	scom "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
)

var (
	ethAddress        = "6964DC8047BdD56A4B8aCCA2803E6B2E026d31bf"
	ethAddrBytes      = []byte{105, 100, 220, 128, 71, 189, 213, 106, 75, 138, 204, 162, 128, 62, 107, 46, 2, 109, 49, 191}
	eosProof          = "1ff0b83f5ee8a15151f61a6225a4a1e42f5a76ec1cae47227b897af0e6ca9cfd"
	poly_go_sdk_Proof = []byte{242, 164, 29, 47, 66, 21, 83, 225, 200, 203, 15, 167, 166, 222, 33, 69, 126, 38, 76, 177, 158, 199, 83, 74, 82, 63, 91, 23, 135, 211, 81, 95, 125, 161, 10, 49, 16, 126, 254, 88, 101, 157, 38, 19, 104, 217, 201, 240, 222, 187, 191, 122, 117, 149, 198, 127, 147, 137, 184, 212, 253, 157, 102, 54, 222, 174, 161, 206, 153, 79, 117, 81, 58, 220, 138, 132, 4, 172, 206, 183, 180, 21, 21, 31, 111, 125, 32, 249, 109, 19, 74, 249, 115, 250, 190, 85}
	//   [242 164 29 47 66 21 83 225 200 203 15 167 166 222 33 69 126 38 76 177 158 199 83 74 82 63 91 23 135 211 81 95 125 161 10 49 16 126 254 88 101 157 38 19 104 217 201 240 222 187 191 122 117 149 198 127 147 137 184 212 253 157 102 54 222 174 161 206 153 79 117 81 58 220 138 132 4 172 206 183 180 21 21 31 111 125 32 249 109 19 74 249 115 250 190 85]
)

func TestTransToEthAddress(t *testing.T) {

	ethAddr := ethcommon.HexToAddress(ethAddress)
	newEthByte, err := hex.DecodeString(ethAddress)
	if err != nil {
		fmt.Printf("error is :%s", err)
	}
	ethHexStr := hex.EncodeToString(ethAddr.Bytes())
	fmt.Printf("newEthByte:%v\n", ethHexStr)
	fmt.Printf("newEthByte:%v\n", string(newEthByte))

}

func TestTeansEthAddressToHex(t *testing.T) {

	fmt.Printf("ethcommon.BytesToAddress(ethAddrBytes).Hex(): %v\n", ethcommon.BytesToAddress(ethAddrBytes).Hex())
	eosByte, _ := hex.DecodeString(eosProof)
	fmt.Printf("eosByte: %v\n", eosByte)
}

// [8 228 13 0 0 0 0 0 0 32 254 92 47 60 232 116 39 5 142 244 59 28 57 229 195 172 24 151 90 199 133 157 158 147 153 71 169 107 34 67 244 17 12 100 100 99 46 99 111 110 116 114 97 99 116 163 191 66 168 208 144 3 0 20 105 100 220 128 71 189 213 106 75 138 204 162 128 62 107 46 2 109 49 191 4 104 101 97 114 175 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 8 100 100 99 46 99 111 110 49 1 12 100 100 99 99 99 109 97 110 97 103 101 114 8 100 100 99 46 99 111 110 50 133 18 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 37 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 49 54 55 54 48 50 49 56 53 53 49 54 56 55 57 50 56 48 48 8 116 101 115 116 100 97 116 97]
// [8 227 13 0 0 0 0 0 0 32 89 12 205 198 22 238 171 44 64 165 25 241 194 139 134 94 144 8 163 103 191 209 202 125 8 194 143 41 61 139 173 70 12 100 100 99 46 99 111 110 116 114 97 99 116 163 191 66 168 208 144 3 0 20 105 100 220 128 71 189 213 106 75 138 204 162 128 62 107 46 2 109 49 191 4 104 101 97 114 175 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 8 100 100 99 46 99 111 110 49 1 12 100 100 99 99 99 109 97 110 97 103 101 114 8 100 100 99 46 99 111 110 50 132 18 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 37 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 49 54 55 54 48 50 49 55 54 48 48 49 57 54 49 49 56 48 48 8 116 101 115 116 100 97 116 97]

// [89 12 205 198 22 238 171 44 64 165 25 241 194 139 134 94 144 8 163 103 191 209 202 125 8 194 143 41 61 139 173 70]

// func TestCalCulateProof(t *testing.T) {
// 	proof := new(proof.EOSProof)
// }

func TestTransE(t *testing.T) {
	var data = []byte{8, 159, 14, 0, 0, 0, 0, 0, 0, 32, 9, 251, 110, 175, 111, 31, 236, 237, 219, 171, 39, 98, 142, 123, 65, 234, 168, 48, 62, 142, 244, 31, 136, 113, 132, 230, 71, 40, 186, 114, 210, 187, 12, 100, 100, 99, 46, 99, 111, 110, 116, 114, 97, 99, 116, 163, 191, 66, 168, 208, 144, 3, 0, 20, 105, 100, 220, 128, 71, 189, 213, 106, 75, 138, 204, 162, 128, 62, 107, 46, 2, 109, 49, 191, 4, 104, 101, 97, 114, 175, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 100, 100, 99, 46, 99, 111, 110, 49, 1, 12, 100, 100, 99, 99, 99, 109, 97, 110, 97, 103, 101, 114, 8, 100, 100, 99, 46, 99, 111, 110, 50, 66, 19, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 37, 104, 116, 116, 112, 115, 58, 47, 47, 103, 105, 116, 104, 117, 98, 46, 99, 111, 109, 49, 54, 55, 54, 53, 51, 52, 53, 52, 50, 48, 54, 55, 48, 53, 57, 48, 48, 48, 8, 116, 101, 115, 116, 100, 97, 116, 97}
	data2 := common.NewZeroCopySource(data)
	// txParam := new(manager.MakeTxParam)
	// if err := txParam.Deserialization(data2); err != nil {
	// 	log.Errorf("<-->本地反序列化错误 error:%s", err)
	// }
	// log.Infof("<-->反序列化解析日志:TxHash is:%v\nCrossChainId is: %v\n,FromContractAddress: %v\n,ToChainId: %v\n,ToContractAddress: %v\n,Method: %v\n,args: %v\n",
	// 	txParam.TxHash, txParam.CrossChainID, txParam.FromContractAddress, txParam.ToChainID, txParam.ToContractAddress, txParam.Method, txParam.Args)
	txParam2 := new(scom.MakeTxParam)
	if err := txParam2.Deserialization(data2); err != nil {
		log.Errorf("<-->Poly反序列化错误 error:%s", err)
	}
	log.Infof("<-->Poly反序列化解析日志:TxHash is:%v\nCrossChainId is: %v\n,FromContractAddress: %v\n,ToChainId: %v\n,ToContractAddress: %v\n,Method: %v\n,args: %v\n",
		txParam2.TxHash, txParam2.CrossChainID, txParam2.FromContractAddress, txParam2.ToChainID, txParam2.ToContractAddress, txParam2.Method, txParam2.Args)

}

func TestDemo(t *testing.T) {
	// demo := `6d00000000000000fe11810160fb0eb8fe7ffa027c69ca6545ab5e4e5a5988f15267d8eeceeb8ccbd62c8b2a4d05b277229c27730a4092a06dd02b1653c43b1365207418112bb160dad865e9fbdb95bb8366d5859809a86072b73867ac4b641475962271c0d74544a38ff8036814814c47d24c97370133968cd2de9a21739206ed8cfd090108150f0000000000002096d96a87310ac8da15f086f176188992834d70bfe9b70d18167ecc41154ea2960c6464632e636f6e7472616374a3bf42a8d0900300146964dc8047bdd56a4b8acca2803e6b2e026d31bf0468656172af01000000000000000000000000000000000000
	// 00000000000000000000000000086464632e636f6e31010c64646363636d616e61676572086464632e636f6e32b81300000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002568747470733a2f2f6769746875622e636f6d3136373639353130323235333632363836303008746573746461746100`
	// fmt.Printf("demo:%s\n", demo)

	// demoByte, err := hex.DecodeString(demo)
	// if err != nil {
	// 	fmt.Printf("hex.DecodeString error:%s\n", err)
	// }

	// params := new(scom.EntranceParam)

	// if err := params.Deserialization(common.NewZeroCopySource(demoByte)); err != nil {
	// 	fmt.Printf("EOS MakeDepositProposal, contract params deserialize error: %s\n", err)
	// }

	// fmt.Printf("demoString address:%v", &demo)

	// fmt.Printf("params address:%v", &params)

	demo := []byte{40, 246, 153, 5, 252, 92, 19, 9, 248, 108, 36, 226, 177, 122, 155, 201, 110, 229, 128, 144, 48, 157, 252, 133, 164, 102, 170, 62, 231, 144, 181, 165}

	st1 := new(Stdo)

	if err := st1.Deserialization(demo); err != nil {
		fmt.Printf("demo Deserialization error:%s", err)
	}
	fmt.Printf("st1.leaf: %v", st1.leaf)

	st1.leaf[0] = 1
	fmt.Printf("更改后:\n")

	fmt.Printf("demo: %v\n", demo)
	fmt.Printf("st1: %v\n", st1)
}

type checksumPath []byte

type Stdo struct {
	leaf checksumPath
}

func (st *Stdo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBytes(st.leaf)
}

func (st *Stdo) Deserialization(data []byte) error {

	source := common.NewZeroCopySource(data)
	var leaf checksumPath
	leaf, eof := source.NextBytes(32)
	if eof {
		return fmt.Errorf("Waiting deserialize leaf error")
	}

	st.leaf = leaf
	return nil
}
