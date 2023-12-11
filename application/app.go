package application

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
	config Config
}

func New(cfg Config) *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisAddress,
		}),
		config: cfg,
	}

	app.loadRouter()

	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.ServerPort),
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

	fmt.Printf("Starting server app at port: %d", app.config.ServerPort)

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
