package application

import (
	"vcbiotech/microservice/domain/file"
	"vcbiotech/microservice/domain/user"

	"github.com/labstack/echo/v4"
)

func (a *App) loadUserRoutes(g *echo.Group) {
	userHandler := &user.UserRepo{
		Repo: &user.SQLRepo{
			Client: a.db,
		},
	}

	g.GET("/:cursor", userHandler.List)

	// g.POST("/", userHandler.Create)

	// g.GET("/:id", userHandler.GetByID)

	// g.PUT("/:id", userHandler.UpdateById)

	// g.DELETE("/:id", userHandler.DeleteById)
}

func (a *App) loadFileRoutes(g *echo.Group) {
	fileHandler := &file.FileHandler{
		Repo: &file.SQLRepo{
			Client: a.db,
		},
	}

	g.GET("/:id", fileHandler.FindById)

	g.POST("/render-template", fileHandler.Insert)

	// g.POST("/", userHandler.Create)

	// g.GET("/:id", userHandler.GetByID)

	// g.PUT("/:id", userHandler.UpdateById)

	// g.DELETE("/:id", userHandler.DeleteById)
}
