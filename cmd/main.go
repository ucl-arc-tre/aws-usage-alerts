package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	//"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

func main() {
	log.Debug().Msg("Hello")

	// config.Init()
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	go eventLoop(ctx, &wg)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	<-signalCh
	log.Debug().Msg("Received termination signal")
	cancel()
	wg.Wait()
	log.Debug().Msg("Exited gracefully")
}

func eventLoop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			things()
		}
	}
}

func things() {
	log.Debug().Msg("started a")

	a()
	time.Sleep(50 * time.Millisecond)
	log.Debug().Msg("done loop")
}

func a() {
	time.Sleep(1 * time.Second)
	log.Debug().Msg("done a")
}
