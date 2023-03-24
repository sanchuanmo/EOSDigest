package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/polynetwork/eos_relayer/log"
)

// RPCClient API

type RpcClient struct {
	addr      string
	httpCient *http.Client
}

// NewRpcClient return RpcClient instance
func NewRpcClient() *RpcClient {
	return &RpcClient{
		httpCient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				DisableKeepAlives:     false, // enable keepalive
				IdleConnTimeout:       time.Second * 300,
				ResponseHeaderTimeout: time.Second * 300,
			},
			Timeout: time.Second * 300, // timeout for http response
		},
	}
}

// SetAddress set rpc server address. Simple http://localhost:20336
func (client *RpcClient) SetAddress(addr string) *RpcClient {
	client.addr = addr
	return client
}

// SetHttpClient set http client to RpcClient. In most cases SetHttpClient is not necessary
func (client *RpcClient) SetHttpClient(htttpClient *http.Client) *RpcClient {
	client.httpCient = htttpClient
	return client
}

func (client *RpcClient) sendRPCRequest(data []byte) (*JsonRpcResponse, error) {
	resp, err := client.httpCient.Post(client.addr, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, PostErr{fmt.Errorf("http post request:%s error:%s", data, err)}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rpc response body errpr: %s", err)
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal JsonRpcResponse%s error:%s", body, err)
	}
	return rpcRsp, nil
}

type PostErr struct {
	Err error
}

func (err PostErr) Error() string {
	return err.Err.Error()
}

type JsonRpcResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

func (client *RpcClient) SendCrossChainInfo(crossInfo *CrossChainInfo) (string, string, error) {
	crossInfoByte, err := json.Marshal(crossInfo)
	if err != nil {
		return "", "", fmt.Errorf("sendCrossChainInfo json.Marshal crossInfo error: %s", err)
	}
	log.Infof("11111111111111SendCrossChainInfo data is:%v", string(crossInfoByte))
	resp, err := client.sendRPCRequest(crossInfoByte)
	return resp.Code, resp.Msg, err
}
