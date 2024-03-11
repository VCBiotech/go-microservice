package application

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"vcbiotech/microservice/domain/order"
	"vcbiotech/microservice/telemetry"
)

func (a *App) loadMiddleware() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(telemetry.Tracing)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/health"))
	router.Use(httprate.LimitByIP(500, 1*time.Minute))
	a.router = router
}

func (a *App) loadRoutes() {
	a.router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("VCBiotech Microservice."))
	})

	// App V1
	a.router.Route("/v1/orders", a.loadOrderRoutes)
}

func (a *App) loadOrderRoutes(router chi.Router) {
	orderHandler := &order.OrderRepo{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}

	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateById)
	router.Delete("/{id}", orderHandler.DeleteById)
}
