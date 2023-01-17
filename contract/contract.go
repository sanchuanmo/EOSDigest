package contract

import (
	"github.com/qqtou/eos-go"
)

const (
	VERIFYEXETXE = "verifyexetx"
	CHBOOKKEEPE  = "chbookkeepee"
	INITGENBLOCK = "initgenblock"
)

type InputCrosschain struct {
	ToChainId  float64 `json:"toChainId"`
	ToContract string  `json:"toContract"`
	Method     string  `json:"method"`
	TxData     string  `json:"txData"`
}

func (basics *Basics) Crosschain(toChainId float64, toContract, method, txData string) *eos.Action {
	return &eos.Action{
		Account: basics.Contract,
		Name:    basics.ActionName,
		Authorization: []eos.PermissionLevel{
			{Actor: basics.Caller, Permission: eos.PermissionName(basics.Per)},
		},
		ActionData: eos.NewActionData(InputCrosschain{
			ToChainId:  toChainId,
			ToContract: toContract,
			Method:     method,
			TxData:     txData,
		}),
	}
}

/*
以上为测试
*/
type Basics struct {
	Caller     eos.AccountName //调用账户
	Contract   eos.AccountName //调用合约
	ActionName eos.ActionName  //调用方法
	Per        string          // 调用权限
}

type InputChbookkeeper struct {
	RawHeader  string `json:"rawHeader,omitempty"`
	PubKeyList string `json:"pubKeyList,omitempty"`
	SigList    string `json:"sigList,omitempty"`
}

type InputVerifyexetx struct {
	Proof        string `json:"proof,omitempty"`
	RawHeader    string `json:"rawHeader,omitempty"`
	HeaderProof  string `json:"headerProof,omitempty"`
	CurRawHeader string `json:"curRawHeader,omitempty"`
	HeaderSig    string `json:"headerSig,omitempty"`
}

type InputInitgenblock struct {
	RawHeader  string `json:"rawHeader,omitempty"`
	PubKeyList string `json:"pubKeyList,omitempty"`
}

func (basics *Basics) Initgenblock(rawHeader, pubKeyList string) *eos.Action {
	return &eos.Action{
		Account: basics.Contract,
		Name:    basics.ActionName,
		Authorization: []eos.PermissionLevel{
			{Actor: basics.Caller, Permission: eos.PermissionName(basics.Per)},
		},
		ActionData: eos.NewActionData(InputInitgenblock{rawHeader, pubKeyList}),
	}
}

// 更新bookkeeper
func (basics *Basics) Chbookkeeper(rawHeader, pubKeyList, sigList string) *eos.Action {
	return &eos.Action{
		Account: basics.Contract,
		Name:    basics.ActionName,
		Authorization: []eos.PermissionLevel{
			{Actor: basics.Caller, Permission: eos.PermissionName(basics.Per)},
		},
		ActionData: eos.NewActionData(InputChbookkeeper{rawHeader, pubKeyList, sigList}),
	}
}

// 验证跨链交易并执行
func (basics *Basics) Verifyexetx(proof, rawHeader, headerProof, curRawHeader, headerSig string) *eos.Action {
	return &eos.Action{
		Account: basics.Contract,
		Name:    basics.ActionName,
		Authorization: []eos.PermissionLevel{
			{Actor: basics.Caller, Permission: eos.PermissionName(basics.Per)},
		},
		ActionData: eos.NewActionData(InputVerifyexetx{proof, rawHeader, headerProof, curRawHeader, headerSig}),
	}
}

/** 上方为测试结构，待合约完成以此结构为准
type InputCrosschain struct {
	TxData []byte `json:"txData"`
}

func Crosschain(caller eos.AccountName, contract eos.AccountName, actionName eos.ActionName, per string, txData []byte) *eos.Action {
	return &eos.Action{
		Account: contract,
		Name:    actionName,
		Authorization: []eos.PermissionLevel{
			{Actor: caller, Permission: eos.PN(per)},
		},
		ActionData: eos.NewActionData(InputCrosschain{
			TxData: txData,
		}),
	}
}
*/
