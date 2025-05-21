package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/todoflow-labs/shared-dtos/logging"
	"github.com/todoflow-labs/sse-updates/internal/config"
	"github.com/todoflow-labs/sse-updates/internal/sse"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	//init logger
	logger := logging.New(cfg.LogLevel).With().Str("service", "sse-updates").Logger()
	logger.Info().Msg("starting SSE service")
	broker := sse.NewBroker()
	go broker.Run()
	logger.Info().Msg("SSE broker started")

	// Start NATS listener
	logger.Info().Msgf("connecting to NATS at %s", cfg.NATSURL)
	if err := sse.StartNATSListener(cfg.NATSURL, broker); err != nil {
		log.Fatal(err)
	}
	logger.Info().Msg("NATS listener started")

	// Start HTTP server
	logger.Info().Msgf("starting HTTP server on %s", cfg.HTTPAddr)
	http.Handle("/events", authMiddleware(sse.NewHandler(broker), logger))
	if err := http.ListenAndServe(cfg.HTTPAddr, nil); err != nil {
		log.Fatal(err)
	}
}

const UserIDKey string = "user_id"

// type ctxKey string

func authMiddleware(next http.Handler, logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		logger.Debug().Str("X-User-ID from header", userID).Msg("authMiddleware")
		if userID == "" {
			if os.Getenv("ENV") == "development" {
				userID = "test-user"
			} else {
				http.Error(w, "X-User-ID header required", http.StatusUnauthorized)
				return
			}
		}
		logger.Debug().Str("user_id", userID).Msg("authMiddleware")
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
