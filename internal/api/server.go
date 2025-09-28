package api

import (
	"net/http"

	"github.com/rs/zerolog/log"

	gerrors "errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func StartServer() {
	log.Info().Msg("Starting a server")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// health check endpoint
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// readiness check endpoint
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("..."))
	})

	if err := http.ListenAndServe(":8080", r); err != nil && !gerrors.Is(err, http.ErrServerClosed) {
		log.Error().Msg("Could not start a server")
	}
}
