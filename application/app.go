package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	router http.Handler
}

func New() *App {
	app := &App{
		router: loadRoutes(),
	}
	return app
}

func (a *App) Start(ctx context.Context) error {

	// Start server
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	// Handle error on startup
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}

	// Return nil
	return nil
}
