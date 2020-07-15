package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/xenitab/flux-status/pkg/api"
	"github.com/xenitab/flux-status/pkg/exporter"
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
	pollWorkloads := flag.Bool("poll-workloads", true, "Enables polling of workloads after sync.")
	pollInterval := flag.Int("poll-intervall", 5, "Duration in seconds between each service poll.")
	pollTimeout := flag.Int("poll-timeout", 0, "Duration in seconds before stopping poll.")
	gitUrl := flag.String("git-url", "", "URL for git repository, should be same as flux.")
	azdoPat := flag.String("azdo-pat", "", "Tokent to authenticate with Azure DevOps.")
	gitlabToken := flag.String("gitlab-token", "", "Token to authenticate with Gitlab.")
	flag.Parse()

	// Logs
	log, err := getLogger(*debug)
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	setupLog := log.WithName("setup")

	// Start Server
	setupLog.Info("Staring flux-status")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	exporter, err := exporter.GetExporter(*gitUrl, *azdoPat, *gitlabToken)
	if err != nil {
		setupLog.Error(err, "Error getting exporter", "url", gitUrl)
		os.Exit(1)
	}
	log.Info("Using exporter with name", "exporter", exporter.String())

	var p *poller.Poller
	if *pollWorkloads {
		p = poller.NewPoller(log.WithName("poller"), *fluxAddr, *pollInterval, *pollTimeout, *instance)
	}

	apiServer := api.NewServer(exporter, p, *instance, log.WithName("api-server"))
	go func() {
		if err := apiServer.Start(*listenAddr); err != nil {
			log.Error(err, "Error occured when running http server")
			os.Exit(1)
		}
	}()

	// Blocks until stop singal is sent
	<-done
	setupLog.Info("Stopping flux-status")

	// Stop server with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Stop(ctx); err != nil {
		log.Error(err, "Error occured when stopping api server")
		os.Exit(1)
	}

	setupLog.Info("Stopped flux-status successfully")
}
