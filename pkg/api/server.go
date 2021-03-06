package api

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	"github.com/xenitab/flux-status/pkg/notifier"
)

// Server implements the api endpoints to receive events sent by Flux.
type Server struct {
	Notifier   notifier.Notifier
	Events     chan<- string
	Log        logr.Logger
	httpServer *http.Server
}

// NewServer creates and returns a Server instance.
func NewServer(n notifier.Notifier, e chan<- string, l logr.Logger) *Server {
	return &Server{
		Notifier: n,
		Events:   e,
		Log:      l,
	}
}

// Start starts serving the api server.
func (s *Server) Start(addr string) error {
	router := mux.NewRouter()
	router.HandleFunc("/v6/events", eventHandler(s.Log, s.Notifier, s.Events))
	router.PathPrefix("/").HandlerFunc(websocketHandler(s.Log))

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return s.httpServer.ListenAndServe() // blocking
}

// Stop gracefully stops serving the api server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
