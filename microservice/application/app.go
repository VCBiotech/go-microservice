package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"vcbiotech/microservice/telemetry"
)

type App struct {
	router *chi.Mux
	rdb    *redis.Client
	config Config
}

func New(config Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
		config: config,
	}

	app.loadMiddleware()
	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	logger := telemetry.GetLogger()

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	// Check if redis is working
	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}

	defer func() {
		if err := a.rdb.Close(); err != nil {
			logger.Error("Failed to close Redis Client: %s", err)
		}
	}()

	logger.Info("Starting application server")

	ch := make(chan error, 1)

	// Call the main function using another thread
	go func() {
		// Handle error on startup
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("Failed to start server: %w", err)
		}
		// Close the channel
		close(ch)
	}()

	logger.Info("Application startup complete.")
	logger.Info(fmt.Sprintf("Listening on address localhost, port %d", a.config.ServerPort))

	select {
	// Select one of the channels, the one that returns first
	case err = <-ch: // This channel returns if the server dies
		return err
	case <-ctx.Done(): // This channel returns on SIGINT
		logger.Info("Shutting down gracefully...")
		// Let's have a graceful timeout of 10 seconds
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// Run Cancel at the very end of execution
		defer cancel()
		// Shutdown server after said timeout
		return server.Shutdown(timeout)
	}
}
