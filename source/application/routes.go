package application

import (
	"github.com/go-chi/chi/v5"

	"vcbiotech/microservice/domain/user"
)

func (a *App) loadUserRoutes(router chi.Router) {
	userHandler := &user.UserRepo{
		Repo: &user.SQLRepo{
			Client: a.db,
		},
	}

	router.Get("/", userHandler.List)
	router.Post("/", userHandler.Create)
	router.Get("/{id}", userHandler.GetByID)
	router.Put("/{id}", userHandler.UpdateById)
	router.Delete("/{id}", userHandler.DeleteById)
}
