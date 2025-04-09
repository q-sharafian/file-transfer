package reqhandler

import (
	"github.com/q-sharafian/file-transfer/internal/server"
)

type ReqHandler interface {
	// Process a request (i.e. download/upload) and response the result to the client
	HandleRequest(ioDetails *ReqDetails)
}

type ioType int

const (
	Upload   ioType = 1
	Download ioType = 2
)

type ReqDetails struct {
	Type ioType
	server.ResponseWriter
	*server.Request
}

type downlaodResponse struct {
	StatusCode int    `json:"status-code"`
	Message    string `json:"message"`
	// A map from file tokens to file urls. If the client hasn't permission to access
	// a file, set value of its corresponding token to an empty string.
	Tokens2URLs map[string]string `json:"tokens2urls"`
}

type uploadResponse struct {
	StatusCode int    `json:"status-code"`
	Message    string `json:"message"`
	// A map from file types to file urls. If the client hasn't permission to access
	// a file, set value of its corresponding token to an empty string.
	// If we want to upload 5 png files, we have a key called png that has 5 uplaod link as the key.
	Tokens2URLs map[string][]string `json:"tokens2urls"`
}
