package db

import (
	"fmt"
	"testing"

	"github.com/polynetwork/eos_relayer/config"
	"github.com/polynetwork/eos_relayer/log"
	"github.com/polynetwork/poly/common"
	"github.com/qqtou/eos-go"
)

func Test_NewBoltDB(t *testing.T) {

	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}
	fmt.Printf("db:%v", db)
}

type CrossStatus struct {
	bolckNum   uint32
	txId       string
	sendStatus bool
}

var ConfigPath string = "../config_eos.json"

// 序列化
func (this *CrossStatus) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.txId)
	sink.WriteUint32(this.bolckNum)
	sink.WriteBool(this.sendStatus)
}

// 反序列化
func (this *CrossStatus) Deserialization(source *common.ZeroCopySource) error {
	txId, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize txId error")
	}
	bolckNum, eof := source.NextUint32()
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

func getEOSServer() *eos.API {
	// read config
	servConfig := getEOSServerConfig()
	// 注册api
	chainApi := eos.New(servConfig.EOSConfig.RestURL)
	return chainApi
}

/*获取EOSSerrverConfig*/
func getEOSServerConfig() config.ServiceEOSConfig {
	servConfig := config.NewServiceEOSConfig(ConfigPath)
	return *servConfig
}

/*获取BoltDB*/
func getBoltDB() (*BoltDB, error) {
	var boltDB *BoltDB
	var err error
	servConfig := getEOSServerConfig()
	if servConfig.BoltDbPath == "" {
		boltDB, err = NewBoltDB("boltdb")
	} else {
		boltDB, err = NewBoltDB(servConfig.BoltDbPath)
	}
	if err != nil {
		log.Fatalf("db.NewWaitingDB error:%s", err)
	}
	return boltDB, err
}

func TestUpdateHeight(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}
	err = db.UpdatePolyHeight(123)
	if err != nil {
		panic("failed to save height of poly: %v" + err.Error())
	}
	fmt.Printf("db.GetPolyHeight(): %v\n", db.GetPolyHeight())

}

func Test_PutRetry(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}
	crossTx := &CrossStatus{
		txId:       "txid68",
		bolckNum:   68,
		sendStatus: true,
	}
	sink := common.NewZeroCopySink(nil)
	crossTx.Serialization(sink)
	err = db.PutRetry(sink.Bytes())
	if err != nil {
		fmt.Printf("this.db.PutRetry error: %s", err)
	} else {
		fmt.Printf("db.put retry success bolckNum : %d\n txId :%v \n sendStatus %v \n", crossTx.bolckNum, crossTx.txId, crossTx.sendStatus)
	}
	db.Close()
}

func Test_DeleteRetry(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}
	crossTx := &CrossStatus{
		txId:       "txid68",
		bolckNum:   68,
		sendStatus: true,
	}
	sink := common.NewZeroCopySink(nil)
	crossTx.Serialization(sink)
	err = db.DeleteRetry(sink.Bytes())
	if err != nil {
		fmt.Printf("this.db.PutRetry error: %s", err)
	} else {
		fmt.Printf("db.put retry success bolckNum : %d\n txId :%v \n sendStatus %v \n", crossTx.bolckNum, crossTx.txId, crossTx.sendStatus)
	}
	db.Close()
}

func Test_GetAllRetry(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}

	retryList, err := db.GetAllRetry()
	if err != nil {
		fmt.Printf("db.get all retry error :%s\n", err)
	} else {
		for i, v := range retryList {
			crossEvent := new(CrossStatus)
			err := crossEvent.Deserialization(common.NewZeroCopySource(v))
			if err != nil {
				fmt.Printf("handleLockDepositEvents - retry.Deserialization error: %s", err)
				continue
			}

			TxId := crossEvent.txId
			bolckNum := crossEvent.bolckNum
			sendStatus := crossEvent.sendStatus
			fmt.Printf("the %d event : txId:%v,\n bolckNum:%d\n sendStatus %v \n", i, TxId, bolckNum, sendStatus)
			// response, _ := json.Marshal(crossEvent)
			// fmt.Printf("index is %d,value is %s", i, string(response))
		}
	}
	db.Close()
}

/*
func Test_PutStatus(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}
	crossTx := &CrossStatus{
		txId:       "txid68",
		bolckNum:   68,
		sendStatus: false,
	}
	sink := common.NewZeroCopySink(nil)
	crossTx.Serialization(sink)
	err = db.PutStatus(sink.Bytes())
	if err != nil {
		fmt.Printf("this.db.PutRetry error: %s", err)
	} else {
		fmt.Printf("db.put retry success bolckNum : %d\n txId :%v \n sendStatus %v \n", crossTx.bolckNum, crossTx.txId, crossTx.sendStatus)
	}
	db.Close()
}
func Test_GetAllStatus(t *testing.T) {
	db, err := NewBoltDB("./")
	if err != nil {
		fmt.Printf("NewBoltDB err:%v", err)
	}

	retryList, err := db.GetAllStatus()
	if err != nil {
		fmt.Printf("db.get all retry error :%s\n", err)
	} else {
		for i, v := range retryList {
			crossEvent := new(CrossStatus)
			err := crossEvent.Deserialization(common.NewZeroCopySource(v))
			if err != nil {
				fmt.Printf("handleLockDepositEvents - retry.Deserialization error: %s", err)
				continue
			}

			TxId := crossEvent.txId
			bolckNum := crossEvent.bolckNum
			sendStatus := crossEvent.sendStatus
			fmt.Printf("the %d event : txId:%v,\n bolckNum:%d\n sendStatus %v \n", i, TxId, bolckNum, sendStatus)
			// response, _ := json.Marshal(crossEvent)
			// fmt.Printf("index is %d,value is %s", i, string(response))
		}
	}
	db.Close()
}
*/
