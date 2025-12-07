package main

// @title           Jupyter API
// @version         1.0.0
// @description     API for Jupyter Platform (DMS, contact-us, etc).

// @host      localhost:9393
// @BasePath  /api/v1
import (
	"log"
	"net/http"

	_ "shreshtasmg.in/jupyter/docs"
	"shreshtasmg.in/jupyter/internal/company"
	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/database"
	"shreshtasmg.in/jupyter/internal/filemeta"
	"shreshtasmg.in/jupyter/internal/httpserver"
	"shreshtasmg.in/jupyter/internal/uploader"
)

func main() {

	config.LoadEnv()
	cfg := config.Load()

	db := database.New(cfg.DSN)

	companyRepo := company.NewRepository(db)
	uploaderRepo := uploader.NewRepository(db)
	fileMetaRepo := filemeta.NewRepository(db)
	s3Service := uploader.NewS3Service()
	uploaderConfigHandler := uploader.NewHandler(uploaderRepo, companyRepo, s3Service, fileMetaRepo)

	router := httpserver.NewRouter(cfg, uploaderConfigHandler)

	log.Printf("starting HTTP server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
