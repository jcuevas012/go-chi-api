package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jcuevas012/orders-api/handler"
)

func loadRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", loadOrderRoutes)

	return router
}

func loadOrderRoutes(router chi.Router) {
	orderHandler := &handler.Order{}

	router.Get("/", orderHandler.List)
	router.Post("/", orderHandler.Create)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteById)
}