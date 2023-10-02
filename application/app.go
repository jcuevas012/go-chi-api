lpackage application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		router: loadRouter(),
		rdb:    redis.NewClient(&redis.Options{}),
	}

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	err := app.rdb.Ping(ctx).Err()

	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Println("fail to close redis connection")
		}
	}()

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting server app")

	ch := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()

		if err != nil {
			ch <- fmt.Errorf("failed to start the server: %w ", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}

	return nil
}
