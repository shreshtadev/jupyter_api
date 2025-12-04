package main

// @title           Jupyter API
// @version         1.0.0
// @description     API for Jupyter Platform (DMS, contact-us, etc).

// @host      localhost:9393
// @BasePath  /api/v1
import (
	"log"
	"net/http"
	"time"

	_ "shreshtasmg.in/jupyter/docs"
	"shreshtasmg.in/jupyter/internal/auth"
	"shreshtasmg.in/jupyter/internal/company"
	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/contact"
	"shreshtasmg.in/jupyter/internal/database"
	"shreshtasmg.in/jupyter/internal/feature"
	"shreshtasmg.in/jupyter/internal/filemeta"
	"shreshtasmg.in/jupyter/internal/httpserver"
	"shreshtasmg.in/jupyter/internal/uploader"
	"shreshtasmg.in/jupyter/internal/user"
)

func main() {

	//
	// import (
	// 	"crypto/rand"
	// 	"crypto/rsa"
	// 	"crypto/x509"
	// 	"encoding/pem"
	// 	"os"
	// )

	// priv, err := rsa.GenerateKey(rand.Reader, 4096)
	// if err != nil {
	// 	panic(err)
	// }

	// privBytes := x509.MarshalPKCS1PrivateKey(priv)
	// privFile, _ := os.Create("jwt_private.pem")
	// _ = pem.Encode(privFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	// privFile.Close()

	// pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	// pubFile, _ := os.Create("jwt_public.pem")
	// _ = pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	// pubFile.Close()

	config.LoadEnv()
	cfg := config.Load()

	db := database.New(cfg.DSN)

	contactRepo := contact.NewRepository(db)
	companyRepo := company.NewRepository(db)
	uploaderRepo := uploader.NewRepository(db)
	fileMetaRepo := filemeta.NewRepository(db)
	featureRepo := feature.NewRepository(db)
	userRepo := user.NewRepository(db)
	contactHandler := contact.NewHandler(contactRepo)
	companyHandler := company.NewHandler(companyRepo, featureRepo)
	s3Service := uploader.NewS3Service()
	uploaderConfigHandler := uploader.NewHandler(uploaderRepo, companyRepo, s3Service, fileMetaRepo)
	fileMetaHandler := filemeta.NewHandler(fileMetaRepo)
	signer, err := auth.NewJWTSignerFromFile(cfg.PrivateKeyPath, "jupyter-platform", "jupyter-platform-api", 24*time.Hour)
	if err != nil {
		log.Fatalf("failed to init JWT signer: %v", err)
	}
	if err := auth.InitJWKS(); err != nil {
		log.Fatalf("failed to init jwks: %v", err)
	}
	authHandler := auth.NewAuthHandler(userRepo, signer)

	router := httpserver.NewRouter(cfg, contactHandler, companyHandler, uploaderConfigHandler, fileMetaHandler, authHandler)

	log.Printf("starting HTTP server on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
