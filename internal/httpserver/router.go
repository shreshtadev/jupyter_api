package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/uploader"
)

func NewRouter(config *config.Config,
	uploaderConfigHandler *uploader.Handler) http.Handler {
	r := chi.NewRouter()
	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger UI route
	// Served at: /swagger/index.html
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/uploader/files", uploaderConfigHandler.GenerateUploadURL)
		r.Post("/uploader/folders/delete", uploaderConfigHandler.DeleteFolder)
		r.Post("/uploader/files/delete", uploaderConfigHandler.DeleteFile)
	})

	return r
}
