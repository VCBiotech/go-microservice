package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"

	"vcbiotech/microservice/telemetry"
)

type App struct {
	router *echo.Echo
	db     *sql.DB
	config Config
}

func New(config Config) *App {
	app := &App{
		config: config,
	}

	// Connect to database
	db, err := sql.Open("postgres", config.LoadDbUri())
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		log.Println("Could not connect to database.", errMsg)
	} else {
		// Add connection to current app
		app.db = db
	}

	app.loadMiddleware()
	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	logger := telemetry.SLogger(ctx)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", a.config.ServerPort),
		Handler: a.router,
	}

	// Database connection is now established in New()
	logger.Info("Starting application server...")

	ch := make(chan error, 1)

	// Call the main function using another thread
	go func() {
		// Handle error on startup
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		// Close the channel
		close(ch)
	}()

	logger.Info("Application startup complete.")
	logger.Info(fmt.Sprintf("Listening on address localhost, port %d", a.config.ServerPort))

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
	router.Use(telemetry.Tracing())
	router.Use(middleware.Recover())
	router.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1000)))
	a.router = router
}

func (a *App) loadRoutes() {
	a.router.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "File Manager Service.")
	})

	// App V1
	userGroup := a.router.Group("/v1/users")
	a.loadUserRoutes(userGroup)

	fileGroup := a.router.Group("/v1/files")
	a.loadFileRoutes(fileGroup)
}

// func (a *App) loadOrderRoutes(router chi.Router) {
// 	orderHandler := &order.OrderRepo{
// 		Repo: &order.RedisRepo{
// 			Client: a.rdb,
// 		},
// 	}
//
// 	router.Post("/", orderHandler.Create)
// 	router.Get("/", orderHandler.List)
// 	router.Get("/{id}", orderHandler.GetByID)
// 	router.Put("/{id}", orderHandler.UpdateById)
// 	router.Delete("/{id}", orderHandler.DeleteById)
// }
