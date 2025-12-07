package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return db
}
