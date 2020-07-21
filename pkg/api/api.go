package api

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	"github.com/xenitab/flux-status/pkg/notifier"
	"github.com/xenitab/flux-status/pkg/poller"
)

type Server struct {
	Notifier   notifier.Notifier
	Poller     *poller.Poller
	Instance   string
	Log        logr.Logger
	httpServer *http.Server
}

func NewServer(e notifier.Notifier, p *poller.Poller, i string, l logr.Logger) *Server {
	return &Server{
		Notifier: e,
		Poller:   p,
		Instance: i,
		Log:      l,
	}
}

func (s *Server) Start(addr string) error {
	router := mux.NewRouter()
	router.HandleFunc("/v6/events", s.HandleEvents)
	router.PathPrefix("/").HandlerFunc(s.HandleWebsocket)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return s.httpServer.ListenAndServe() // blocking
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
