package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/jcuevas012/orders-api/application"
)

func main() {
	app := application.New(application.LoadConfig())

	cxt, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(cxt)

	if err != nil {
		fmt.Println("failed to start", err)
	}
}
