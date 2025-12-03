package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"shreshtasmg.in/jupyter/internal/contact"
)

func New(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&contact.Contact{}); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	return db
}
