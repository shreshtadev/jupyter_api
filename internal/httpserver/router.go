package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	authpkg "shreshtasmg.in/jupyter/internal/auth"
	"shreshtasmg.in/jupyter/internal/company"
	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/contact"
	"shreshtasmg.in/jupyter/internal/filemeta"
	"shreshtasmg.in/jupyter/internal/uploader"
)

func NewRouter(config *config.Config, contactHandler *contact.Handler,
	companyHandler *company.Handler,
	uploaderConfigHandler *uploader.Handler,
	filemetaHandler *filemeta.Handler, authHandler *authpkg.AuthHandler) http.Handler {
	r := chi.NewRouter()
	jwtValidator, _ := authpkg.NewJWTValidatorFromFile(config.PublicKeyPath)
	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Swagger UI route
	// Served at: /swagger/index.html
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/contact-us", contactHandler.CreateContact)
		r.Post("/companies", companyHandler.CreateCompany)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/uploader/files", uploaderConfigHandler.GenerateUploadURL)
		r.Post("/uploader/folders/delete", uploaderConfigHandler.DeleteFolder)
		r.Group(func(r chi.Router) {
			r.Use(authpkg.JWTMiddleware(jwtValidator))
			r.Get("/auth/me", authHandler.Me)
			r.Get("/companies/files/meta", uploaderConfigHandler.ListCompanyFiles)
			r.Get("/companies/folders", uploaderConfigHandler.ListFolders)
			r.Get("/companies/files", uploaderConfigHandler.ListFilesInFolder)
			r.Post("/uploader/config", uploaderConfigHandler.CreateUploaderConfig)
			r.Post("/uploader/files/delete", uploaderConfigHandler.DeleteFile)
			r.Group(func(r chi.Router) {
				r.Use(authpkg.RequireSuperadmin())
				r.Post("/admin/register", authHandler.Register)
				r.Post("/admin/files/meta", filemetaHandler.CreateFileMeta)
				r.Get("/admin/files/meta/{id}", filemetaHandler.GetFileMetaByID)
			})
		})
	})

	return r
}
