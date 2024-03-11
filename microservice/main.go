package main

import (
	"context"
	"os"
	"os/signal"

	"vcbiotech/microservice/application"
	"vcbiotech/microservice/telemetry"
)

func main() {
	app := application.New(application.LoadConfig())
	logger := telemetry.GetLogger()

	// Listen for SIGINT
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		logger.Error("Failed to start app: %s", err)
	}
}
