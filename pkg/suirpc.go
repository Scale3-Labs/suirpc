package suirpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type httpClient interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

type logger interface {
	Println(v ...interface{})
}

// Sui Error
type SuiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err SuiError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type suiResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *SuiError       `json:"error"`
}

type suiRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// SuiRPC - Sui rpc client
type SuiRPC struct {
	url    string
	client httpClient
	log    logger
	Debug  bool
}

// New create new rpc client with given url
func New(url string, options ...func(rpc *SuiRPC)) *SuiRPC {
	rpc := &SuiRPC{
		url:    url,
		client: http.DefaultClient,
		log:    log.New(os.Stderr, "", log.LstdFlags),
	}
	for _, option := range options {
		option(rpc)
	}

	return rpc
}

// NewSuiRPC create new rpc client with given url
func NewSuiRPC(url string, options ...func(rpc *SuiRPC)) *SuiRPC {
	return New(url, options...)
}

func (rpc *SuiRPC) call(method string, target interface{}, params ...interface{}) error {
	result, err := rpc.Call(method, params...)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	return json.Unmarshal(result, target)
}

// URL returns client url
func (rpc *SuiRPC) URL() string {
	return rpc.url
}

// Call returns raw response of method call
func (rpc *SuiRPC) Call(method string, params ...interface{}) (json.RawMessage, error) {
	request := suiRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := rpc.client.Post(rpc.url, "application/json", bytes.NewBuffer(body))
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if rpc.Debug {
		rpc.log.Println(fmt.Sprintf("%s\nRequest: %s\nResponse: %s\n", method, body, data))
	}

	resp := new(suiResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, *resp.Error
	}

	return resp.Result, nil

}

// service discovery method that will return the OpenRPC schema for the JSON-RPC API.
func (rpc *SuiRPC) Discover() (map[string]interface{}, error) {
	var serviceDiscovery map[string]interface{}

	err := rpc.call("rpc.discover", &serviceDiscovery)
	return serviceDiscovery, err
}