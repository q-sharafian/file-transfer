/*
Implementation of simple web framework/server.
*/
package server

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	l "github.com/q-sharafian/file-transfer/pkg/logger"
)

type oneTimeMux struct {
	*http.ServeMux
	activeRoutes sync.Map
}

func newOneTimeMux() *oneTimeMux {
	return &oneTimeMux{
		ServeMux: http.NewServeMux(),
	}
}

func (mux *oneTimeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if route is still active
	if active, ok := mux.activeRoutes.Load(r.URL.Path); !ok || !active.(bool) {
		http.Error(w, "Upload URL has expired or already been used", http.StatusGone)
		return
	}

	// Mark route as inactive before serving
	mux.activeRoutes.Store(r.URL.Path, false)
	defer mux.activeRoutes.Delete(r.URL.Path)

	// Serve the request
	mux.ServeMux.ServeHTTP(w, r)
}

func (mux *oneTimeMux) HandleOneTime(pattern string, handler http.Handler) {
	mux.activeRoutes.Store(pattern, true)
	mux.Handle(pattern, handler)
}

// This simple server uses net/http package.
type simpleServer struct {
	mux                 *http.ServeMux
	mu                  sync.RWMutex
	oneTimeMux          *oneTimeMux
	oneTimeActiveRoutes sync.Map
	serverAddr          string
	logger              l.Logger
}

func NewSimpleServer(logger l.Logger) Server {
	logger.Infof("Initializing simple server on port %s", os.Getenv("SERVER_PORT"))
	server := simpleServer{
		mux:                 http.NewServeMux(),
		mu:                  sync.RWMutex{},
		oneTimeMux:          newOneTimeMux(),
		oneTimeActiveRoutes: sync.Map{},
		serverAddr:          fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		logger:              logger,
	}
	// Start the server in a separate goroutine
	go func() {
		logger.Debugf("Starting listening on %s", server.serverAddr)
		if err := http.ListenAndServe(server.serverAddr, server.mux); err != nil {
			logger.Panicf("Server failed to start: %s", err.Error())
		}
	}()
	return &server
}

type handler struct {
	handler func(w ResponseWriter, r *Request)
	mu      *sync.RWMutex
}

func newHandler(hndr func(w ResponseWriter, r *Request), mu *sync.RWMutex) *handler {
	return &handler{hndr, mu}
}
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.handler(w, &Request{r})
}

func httpLogger(logger l.Logger, next *handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf(`-----------------------------------------------------------
Handling http request with %s method to url %s`, r.Method, r.URL.String())
		next.ServeHTTP(w, r)
		// Write codes run after running the handler
	})
}

func (s *simpleServer) AddHandler(path string, handler func(w ResponseWriter, r *Request)) {
	s.logger.Debugf("Adding HTTP path handler for \"%s\"", path)
	stdHandler := newHandler(handler, &s.mu)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mux.Handle(path, httpLogger(s.logger, stdHandler))
	// http.ListenAndServe(s.serverAddr, s.mux)
}

// FIXME: It's wrong and maybe doesn't work!
func (s *simpleServer) HandleOneTime(path string, handler func(w ResponseWriter, r *Request)) {
	s.oneTimeActiveRoutes.Store(path, true)
	s.AddHandler(path, handler)

}
