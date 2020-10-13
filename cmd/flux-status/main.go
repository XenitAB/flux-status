package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/xenitab/flux-status/pkg/api"
	"github.com/xenitab/flux-status/pkg/notifier"
	"github.com/xenitab/flux-status/pkg/poller"
)

func getLogger(debug bool) (logr.Logger, error) {
	var zapLog *zap.Logger
	var err error
	if debug {
		zapLog, err = zap.NewDevelopment()
	} else {
		zapLog, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	return zapr.NewLogger(zapLog), nil
}

func main() {
	// Flags
	debug := flag.Bool("debug", false, "Enables debug mode.")
	listenAddr := flag.String("listen", ":3000", "Address to serve events API on.")
	fluxAddr := flag.String("flux", "localhost:3030", "Address to communicate with the Flux API through.")
	instance := flag.String("instance", "default", "Id to differentiate between multiple flux-status updating the same repository.")
	enablePoller := flag.Bool("poll-workloads", true, "Enables polling of workloads after sync.")
	pollInterval := flag.Int("poll-intervall", 5, "Duration in seconds between each service poll.")
	pollTimeout := flag.Int("poll-timeout", 360, "Duration in seconds before stopping poll.")
	gitURL := flag.String("git-url", "", "URL for git repository, should be same as flux.")
	azdoPat := flag.String("azdo-pat", "", "Tokent to authenticate with Azure DevOps.")
	glToken := flag.String("gitlab-token", "", "Token to authenticate with Gitlab.")
	ghToken := flag.String("github-token", "", "Token to authenticate with GitHub.")
	flag.Parse()

	// Logs
	log, err := getLogger(*debug)
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	setupLog := log.WithName("setup")
	setupLog.Info("Staring flux-status")

	// Get Notifier
	notifier, err := notifier.GetNotifier(*instance, *gitURL, *azdoPat, *glToken, *ghToken)
	if err != nil {
		setupLog.Error(err, "Error getting Notifier", "url", gitURL)
		os.Exit(1)
	}
	setupLog.Info("Using notifier", "name", notifier.String())

	// Setup
	shutdownWg := &sync.WaitGroup{}
	shutdown := make(chan struct{})
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Channel is nil if poller is not enabled
	var events chan string = nil

	// Start Poller
	if *enablePoller {
		events = make(chan string, 1)
		shutdownWg.Add(1)
		p, err := poller.NewPoller(log.WithName("poller"), notifier, events, *fluxAddr, *pollInterval, *pollTimeout)
		if err != nil {
			errc <- err
		}
		go p.Start()
		go func() {
			defer shutdownWg.Done()
			<-shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := p.Stop(ctx); err != nil {
				setupLog.Error(err, "Error occured when stopping poller")
			}
			setupLog.Info("Stopped poller")
		}()
	}

	// Start Server
	shutdownWg.Add(1)
	apiServer := api.NewServer(notifier, events, log.WithName("api-server"))
	go func() {
		errc <- apiServer.Start(*listenAddr)
	}()
	go func() {
		defer shutdownWg.Done()
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := apiServer.Stop(ctx); err != nil {
			setupLog.Error(err, "Error occured when stopping server")
		}
		setupLog.Info("Stopped server")
	}()

	// Wait until stop signal or error
	setupLog.Error(<-errc, "Stopping flux-status")
	close(shutdown)
	shutdownWg.Wait()
	setupLog.Info("Stopped flux-status")
}
