package main

import (
	"log"
	"os"

	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/feature"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	config.LoadEnv()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN env variable is required")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate only feature tables
	if err := db.AutoMigrate(&feature.Feature{}, &feature.CompanyFeature{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// ---- Define all platform-wide features here ----
	featuresToSeed := []feature.Feature{
		{
			Key:         "accounting",
			Name:        "Basic Accounts",
			Description: "Payments,Receipts,Bills,Purchases,Invoices, GST",
			IsActive:    false, // not live yet
		},
	}

	log.Println("Seeding features...")

	for _, f := range featuresToSeed {
		var existing feature.Feature
		err := db.Where("feature_key = ?", f.Key).First(&existing).Error

		switch err {
		case gorm.ErrRecordNotFound:
			// Insert new feature
			if err := db.Create(&f).Error; err != nil {
				log.Fatalf("failed to insert feature %s: %v", f.Key, err)
			}
			log.Printf("Inserted feature: %s (%s)", f.Key, f.Name)

		case nil:
			// Update name/description/is_active only (don’t modify removed features)
			existing.Name = f.Name
			existing.Description = f.Description
			existing.IsActive = f.IsActive
			if err := db.Save(&existing).Error; err != nil {
				log.Fatalf("failed to update feature %s: %v", f.Key, err)
			}
			log.Printf("Updated feature: %s", f.Key)

		default:
			log.Fatalf("db error while checking feature %s: %v", f.Key, err)
		}
	}

	log.Println("Feature seeding complete.")
}
