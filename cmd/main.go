package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rs/zerolog"
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

// Writes a structured JSON error
func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func authMiddleware(next http.Handler, logger zerolog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			if os.Getenv("ENV") != "production" {
				logger.Info().Msg("Using test user ID")
				userID = "test-user"
			} else {
				logger.Error().Msg("Missing X-User-ID header")
				writeError(w, http.StatusUnauthorized, "Missing X-User-ID header")
				return
			}
		}
		logger.Info().Msgf("User authenticated, %s", userID)
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
