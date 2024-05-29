package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

// Run a HTTP server on a port suitable to respond to health pings
func Run(ctx context.Context, wg *sync.WaitGroup) {
	http.HandleFunc("/ping", ping)
	server := http.Server{Addr: ":" + config.HealthPort()}

	wg.Add(1)
	defer wg.Done()
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Shutting down the http server")
	if err := server.Shutdown(ctx); err != nil {
		log.Err(err).Msg("Failed to shutdown server")
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}
