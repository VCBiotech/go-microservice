package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"victorcalderon/go-microservice/application"
)

func main() {
	app := application.New(application.LoadConfig())

	// Listen for SIGINT
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed to start app:", err)
	}
}
