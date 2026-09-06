package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"

	"file-manager/auth"
	"file-manager/metadata"
	"file-manager/storage"
	"file-manager/telemetry"
)

type App struct {
	router         *echo.Echo
	db             *sql.DB
	config         *AppConfig
	storageManager *storage.StorageManager
	metadataStore  metadata.MetadataStore
}

// GetRouter returns the router for testing purposes
func (a *App) GetRouter() *echo.Echo {
	return a.router
}

func New(config *AppConfig) *App {
	app := &App{
		config: config,
	}

	storageManager, err := storage.NewStorageManager(config)
	if err != nil {
		log.Printf("Storage manager not initialized: %v", err)
	} else {
		app.storageManager = storageManager
	}

	// Initialize Metadata Store (using in-memory for demo)
	metadataStore := metadata.NewInMemoryMetadataStore()
	app.metadataStore = metadataStore

	app.loadMiddleware()
	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	logger := telemetry.SLogger(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", a.config.ServerPort)
	}

	// Bind all interfaces so Vercel (and other hosts) can reach the server.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: a.router,
	}

	logger.Info("Starting application server...")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	logger.Info("Application startup complete.")
	logger.Info(fmt.Sprintf("Listening on port %s", port))

	select {
	// Select one of the channels, the one that returns first
	case err := <-ch: // This channel returns if the server dies
		return err
	case <-ctx.Done(): // This channel returns on SIGINT
		logger.Info("Shutting down gracefully...")
		// Close database connection
		if a.db != nil {
			if err := a.db.Close(); err != nil {
				errMsg := map[string]string{"Error": err.Error()}
				logger.Error("Failed to close postgres Client", errMsg)
			}
		}
		// Let's have a graceful timeout of 10 seconds
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// Run Cancel at the very end of execution
		defer cancel()
		// Shutdown server after said timeout
		return server.Shutdown(timeout)
	}
}

func (a *App) loadMiddleware() {
	router := echo.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(telemetry.Tracing())
	router.Use(middleware.Recover())
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1000)))
	a.router = router
}

func (a *App) loadRoutes() {
	a.router.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok")
	})
	a.router.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "File Manager Service.")
	}, auth.ServerAuthMiddleware())

	fileGroup := a.router.Group("/v1/files", auth.ServerAuthMiddleware())
	a.loadFileRoutes(fileGroup)
}
