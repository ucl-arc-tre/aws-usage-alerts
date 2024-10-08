package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/health"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/manager"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	config.Init()
	go manager.New().Loop(ctx, &wg)
	go health.Run(ctx, &wg)

	<-signalChannel()
	log.Info().Msg("Received termination signal")
	cancel()
	wg.Wait()
	log.Debug().Msg("Exited successfully")
}

func signalChannel() chan os.Signal {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)
	return channel
}
