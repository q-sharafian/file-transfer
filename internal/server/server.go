/*
Interface of web framework/server.
*/
package server

import "net/http"

type Server interface {
	// Handle HTTP requests for long time.
	AddHandler(path string, handler func(w ResponseWriter, r *Request))
	// After receiving the first request and sending response, the handler will be removed
	// TODO: Handle timeout to if the user didn't use that route until that, remove the handler
	HandleOneTime(path string, handler func(w ResponseWriter, r *Request))
}

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
}

type Request struct {
	*http.Request
}
