package rpc

import (
	"encoding/json"
	"net/http"
)

// RPCRequest represents the JSON-RPC request payload
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
}

// RPCResponse represents the JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
	ID      uint64          `json:"id"`
}

// RPCError handles any error from the RPC response
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCClient struct {
	endpoint string
	client   *http.Client
}
