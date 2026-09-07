package application

import (
	"log"
	"net/http"

	"file-manager/domain/file"
	"file-manager/storage"

	"github.com/labstack/echo/v4"
)

func (a *App) loadStorageManager() (*storage.StorageManager, error) {
	return storage.NewStorageManager(a.config)
}

func (a *App) loadFileRoutes(g *echo.Group) {
	fileRepo, err := file.NewFileRepo(a.storageManager, a.metadataStore, a.config)
	if err != nil {
		log.Printf("Failed to create file repository: %v", err)
		unavailable := func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "storage not configured")
		}
		g.POST("/render-template", unavailable)
		g.POST("/preview", unavailable)
		return
	}

	fileHandler := file.NewFileHandler(fileRepo)

	if a.storageManager == nil {
		g.POST("/render-template", func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "storage not configured")
		})
	} else {
		g.POST("/render-template", fileHandler.Insert)
	}
	g.POST("/preview", fileHandler.PreviewTemplate)
}
