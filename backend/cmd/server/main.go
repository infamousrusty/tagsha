package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/infamousrusty/tagsha/internal/api"
	"github.com/infamousrusty/tagsha/internal/cache"
	"github.com/infamousrusty/tagsha/internal/config"
	ghclient "github.com/infamousrusty/tagsha/internal/github"
)

// version is injected at build time via -ldflags.
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Configure zerolog.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// In development, pretty-print logs.
	if cfg.Env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	log.Info().
		Str("version", version).
		Str("env", cfg.Env).
		Int("port", cfg.Port).
		Bool("github_token_set", cfg.GitHubToken != "").
		Msg("TagSha starting")

	// Initialise Redis cache.
	cacheStore, err := cache.New(cfg.RedisURL, cfg.CacheTTL)
	if err != nil {
		return fmt.Errorf("initialising cache: %w", err)
	}
	log.Info().Str("redis_url", cfg.RedisURL).Msg("Redis connected")

	// Initialise GitHub client and service.
	ghClient := ghclient.New(cfg.GitHubToken, cfg.GitHubAPIBaseURL)
	ghService := ghclient.NewService(ghClient, cfg.MaxPages)

	// Build Redis client for rate limiter (reuses same URL).
	rlOpts, _ := redis.ParseURL(cfg.RedisURL)
	rlClient := redis.NewClient(rlOpts)

	// Build handler and router.
	h := api.New(ghService, cacheStore, version)
	router := api.NewRouter(h, rlClient, cfg.RateLimitRPM, cfg.RequestSizeLimit)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serveErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("HTTP server listening")
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		log.Info().Msg("server shut down cleanly")
	}

	return nil
}
