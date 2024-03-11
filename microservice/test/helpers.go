package test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	"vcbiotech/microservice/application"
	"vcbiotech/microservice/telemetry"
)

type TestApp struct {
	Host   string
	Config application.Config
	DBPool *sqlx.DB
	Client http.Client
}

func configureDatabase(c *application.DatabaseConfig) *sqlx.DB {
	// Get logger
	logger := telemetry.NewLogger(context.Background())
	// Get a connection using sqlx
	connWithoutDB := fmt.Sprintf(
		"postgres://%s:%s@%s:%d?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port,
	)
	// Force connection, ping and panic on failure
	db := sqlx.MustConnect("postgres", connWithoutDB)
	// Create a database with this new connection
	query := fmt.Sprintf("CREATE DATABASE db_%s", c.Name)
	db.MustExec(query)
	// Get a connection using sqlx
	connWithDB := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/db_%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, c.Name,
	)
	logger.Info("Running migrations...")
	// Set up migrations
	m, err := migrate.New(
		"file://../migrations",
		connWithDB,
	)
	if err != nil {
		errAttrs := map[string]string{"Error": err.Error()}
		logger.Error("Could not set up migrations.", errAttrs)
	}
	// Run migrations
	err = m.Up()
	if err != nil {
		errAttrs := map[string]string{"Error": err.Error()}
		logger.Error("Could not run migrations.", errAttrs)
	}
	logger.Info("Running ran successfully!")
	// Return a connection ready to be used
	return sqlx.MustConnect("postgres", connWithDB)
}

func spawnTestApp() *TestApp {
	// Load config
	config := application.LoadConfig()
	// Change database name
	config.Database.Name = fmt.Sprint(rand.Uint64())
	// Build and return application
	app := application.New(config)
	// Run app in the background
	go app.Start(context.Background())
	// Configure database
	db := configureDatabase(&config.Database)
	// Build a test client
	client := http.Client{CheckRedirect: nil, Timeout: time.Duration(2) * time.Second}
	host := fmt.Sprintf("localhost:%d", config.ServerPort)
	// Beautiful complete app
	return &TestApp{
		DBPool: db,
		Host:   host,
		Client: client,
	}
}
