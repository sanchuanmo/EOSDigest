package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/polynetwork/eos_relayer/log"
)

const (
	EOS_MONITOR_INTERVAL         = 15 * time.Second
	POLY_MONITOR_INTERVAL        = 1 * time.Second
	EOS_USEFUL_BLOCK_NUM         = 3
	EOS_PROOF_USERFUL_BLOCK      = 12
	POLY_USEFUL_BLOCK_NUM        = 1
	DEFAULT_EOS_CONFIG_FILE_NAME = "./config_eos.json"
	Version                      = "1.0"
	DEFAULT_LOG_LEVEL            = log.InfoLog // 默认日志等级 default log level = log.InfoLogs
)

type ServiceEOSConfig struct {
	PolyConfig        *PolyConfig
	EOSConfig         *EOSConfig
	CollectInfoConfig *CollectInfoConfig
	BoltDbPath        string
	RoutineNum        int64
	TargetContracts   []map[string]map[string][]uint64
}

type PolyConfig struct {
	RestURL                 string // resturl
	EntranceContractAddress string // entrance contract address 合约地址入口
	WalletFile              string // 钱包文件
	WalletPwd               string // 钱包密码
}

type EOSConfig struct {
	RestURL         string
	SideChainId     uint64
	HeadersPerBatch int
	MonitorInterval uint64
	BlockConfig     uint64
	StoreAccounts   []map[string]string
	ContractAddress string
}

type CollectInfoConfig struct {
	RestURL              string
	MonitorInterval      uint64
	RetryMonitorInterval uint64
	ReSendNum            int
}

func ReadFile(fileName string) ([]byte, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: open file %s error %s", fileName, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Errorf("ReadFile: File %s close error %s", fileName, err)
		}
	}()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: ioutil.ReadAll %s error %s", fileName, err)
	}
	return data, nil
}

func NewServiceEOSConfig(configFilePath string) *ServiceEOSConfig {
	fileContent, err := ReadFile(configFilePath)
	if err != nil {
		log.Errorf("NewServiceConfig: failed, err: %s", err)
		return nil
	}
	servEOSConfig := &ServiceEOSConfig{}
	err = json.Unmarshal(fileContent, servEOSConfig)
	if err != nil {
		log.Errorf("NewServiceConfig: failed, err: %s", err)
		return nil
	}

	return servEOSConfig
}
