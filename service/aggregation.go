package service

import (
	"encoding/json"
	"fmt"

	"github.com/polynetwork/poly/common"
)

type CrossChainInfo struct {
	CrossChain_id   uint64 `json:"cross_chain_id"`  // ddc跨链唯一标识符
	DDC_amount      uint32 `json:"ddc_amount"`      // ddc数量
	DDC_id          string `json:"ddc_id"`          // 起始链ddcID
	DDC_type        uint32 `json:"ddc_type"`        // 起始链ddc类型
	DDC_uri         string `json:"ddc_uri"`         // ddc的URI
	Dynamic_fee_tx  string `json:"dynamic_fee_tx"`  // 以太坊动态费用交易（目标链是以太坊时才会用到）
	From_address    string `json:"from_address"`    // 起始链ddc Owner
	From_cc_addr    string `json:"from_cc_addr"`    // 起始链合约地址
	From_chainid    string `json:"from_chainid"`    // 起始链侧链ID
	From_tx         string `json:"from_tx"`         // 起始链交易hash
	Poly_key        string `json:"poly_key"`        //poly交易id（目标链是以太坊时才会用到）
	Poly_tx         string `json:"poly_tx"`         // 中继链poly交易hash
	Sender          string `json:"sender"`          //发起跨链的地址
	To_cc_addr      string `json:"to_cc_addr"`      //目标链合约地址
	To_address      string `json:"to_address"`      // 目标链上跨链后的ddc所有者
	To_chainId      string `json:"to_chainid"`      // 目标链侧链id
	To_tx           string `json:"to_tx"`           // 目标链上交易hash
	Token_id        string `json:"token_id"`        // 目标链上跨链后的ddcId
	Tx_createtime   string `json:"tx_createtime"`   // 起始链relayer上接收到跨链信息的时间
	Tx_signer       string `json:"tx_signer"`       // 公链上交易指定的签名者(目标链是以太坊时才会用到)
	Tx_status       uint32 `json:"tx_status"`       // 跨链状态，0未知，1成功，2失败（起始链拿到跨链信息了，跨链状态为0；目标链relayer发起交易时成功了是1，失败了是2）
	Tx_time         string `json:"tx_time"`         // 目标链relayer上接收到跨链信息的时间
	Cross_chain_fee uint64 `json:"cross_chain_fee"` // 跨链费用
}

func (crossInfo *CrossChainInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(crossInfo.CrossChain_id)
	sink.WriteUint32(crossInfo.DDC_amount)
	sink.WriteString(crossInfo.DDC_id)
	sink.WriteUint32(crossInfo.DDC_type)
	sink.WriteString(crossInfo.DDC_uri)
	sink.WriteString(crossInfo.Dynamic_fee_tx)
	sink.WriteString(crossInfo.From_address)
	sink.WriteString(crossInfo.From_cc_addr)
	sink.WriteString(crossInfo.From_chainid)
	sink.WriteString(crossInfo.From_tx)
	sink.WriteString(crossInfo.Poly_key)
	sink.WriteString(crossInfo.Poly_tx)
	sink.WriteString(crossInfo.Sender)
	sink.WriteString(crossInfo.To_cc_addr)
	sink.WriteString(crossInfo.To_address)
	sink.WriteString(crossInfo.To_chainId)
	sink.WriteString(crossInfo.To_tx)
	sink.WriteString(crossInfo.Token_id)
	sink.WriteString(crossInfo.Tx_createtime)
	sink.WriteString(crossInfo.Tx_signer)
	sink.WriteUint32(crossInfo.Tx_status)
	sink.WriteString(crossInfo.Tx_time)
	sink.WriteUint64(crossInfo.Cross_chain_fee)
}

func (crossInfo *CrossChainInfo) Deserialization(source *common.ZeroCopySource) error {
	crossChain_id, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize crossChain_id error")
	}
	ddc_amount, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize ddc_amount error")
	}
	ddc_id, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize ddc_id error")
	}
	ddc_type, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize ddc_type error")
	}
	ddc_uri, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize ddc_uri error")
	}
	dynamic_fee_tx, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize dynamic_fee_tx error")
	}
	from_address, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize from_address error")
	}
	from_cc_addr, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize from_cc_addr error")
	}
	from_chainid, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize from_chainid error")
	}
	from_tx, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize from_tx error")
	}
	poly_key, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize poly_key error")
	}
	poly_tx, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize poly_tx error")
	}
	sender, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize sender error")
	}
	to_cc_addr, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize to_cc_addr error")
	}
	to_address, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize to_address error")
	}
	to_chainid, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize to_chainid error")
	}
	to_tx, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize to_tx error")
	}
	token_id, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize token_id error")
	}
	tx_createtime, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize tx_createtime error")
	}
	tx_signer, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize tx_signer error")
	}
	tx_status, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize tx_status error")
	}
	tx_time, eof := source.NextString()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize tx_time error")
	}
	cross_chain_fee, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("CrossChainInfo deserialize cross_chain_fee error")
	}
	crossInfo.CrossChain_id = crossChain_id
	crossInfo.DDC_amount = ddc_amount
	crossInfo.DDC_id = ddc_id
	crossInfo.DDC_type = ddc_type
	crossInfo.DDC_uri = ddc_uri
	crossInfo.Dynamic_fee_tx = dynamic_fee_tx
	crossInfo.From_address = from_address
	crossInfo.From_cc_addr = from_cc_addr
	crossInfo.From_chainid = from_chainid
	crossInfo.From_tx = from_tx
	crossInfo.Poly_key = poly_key
	crossInfo.Poly_tx = poly_tx
	crossInfo.Sender = sender
	crossInfo.To_cc_addr = to_cc_addr
	crossInfo.To_address = to_address
	crossInfo.To_chainId = to_chainid
	crossInfo.To_tx = to_tx
	crossInfo.Token_id = token_id
	crossInfo.Tx_createtime = tx_createtime
	crossInfo.Tx_signer = tx_signer
	crossInfo.Tx_status = tx_status
	crossInfo.Tx_time = tx_time
	crossInfo.Cross_chain_fee = cross_chain_fee

	return nil
}

func NewCrossChainInfo() *CrossChainInfo {
	return &CrossChainInfo{Tx_status: 0}
}

func (crossInfo *CrossChainInfo) TransToJson() (string, error) {
	crossInfoJson, err := json.Marshal(crossInfo)
	if err != nil {
		return "", err
	}
	return string(crossInfoJson), nil
}

// func SendCrossChainInfo(crossInfo *CrossChainInfo, url string) {

// }
