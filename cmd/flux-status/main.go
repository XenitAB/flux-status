package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
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

func main() {
	// Flags
	port := flag.Int("port", 3000, "Port to bind server to.")
	gitUrl := flag.String("git-url", "", "URL for git repository, should be same as flux.")
	azdoPat := flag.String("azdo-pat", "", "PAT to authenticate to Azure DevOps with.")
	fluxPort := flag.Int("flux-port", 3030, "Port for Flux api.")
	pollInterval := flag.Int("poll-intervall", 2, "Duration in seconds between each service poll.")
	pollTimeout := flag.Int("poll-timeout", 20, "Duration in seconds before stopping poll.")
	flag.Parse()

	// Logs
	var log logr.Logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)
	setupLog := log.WithName("setup")

	// Setup
	setupLog.Info("Staring flux-status")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start Server
	azdo, err := exporter.NewAzureDevops(*azdoPat, *gitUrl)
	if err != nil {
		setupLog.Error(err, "Could not configure exporter")
		os.Exit(1)
	}
	poller := poller.NewPoller(log.WithName("poller"), *fluxPort, *pollInterval, *pollTimeout)
	apiServer := api.NewServer(azdo, poller, log.WithName("api-server"))
	go func() {
		if err := apiServer.Start("localhost:" + strconv.Itoa(*port)); err != nil {
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
