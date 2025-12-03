package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"shreshtasmg.in/jupyter/internal/auth"
	"shreshtasmg.in/jupyter/internal/company"
	"shreshtasmg.in/jupyter/internal/config"
	"shreshtasmg.in/jupyter/internal/user"
	"shreshtasmg.in/jupyter/internal/utils"
)

func main() {
	config.LoadEnv()
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	// Auto-migrate users & companies if needed (optional)
	if err := db.AutoMigrate(&user.User{}, &company.Company{}); err != nil {
		log.Fatalf("failed to automigrate: %v", err)
	}

	// ---- CONFIGURE THESE BEFORE RUNNING ----
	superadminEmail := "superadmin@example.com"
	superadminPassword := "SuperAdminPassword123!"

	adminEmail := "admin@example.com"
	adminPassword := "AdminPassword123!"

	userEmail := "user@example.com"
	userPassword := "UserPassword123!"

	defaultCompanyName := "Default Company"
	//-----------------------------------------

	var defaultCompany company.Company
	if err := db.Where("company_slug = ?", "default-company").First(&defaultCompany).Error; err != nil {
		log.Printf("default company not found, creating...")

		companyID := utils.GenerateID()
		defaultCompany = company.Company{
			ID:            companyID,
			CompanyName:   defaultCompanyName,
			CompanySlug:   utils.Slugify(defaultCompanyName),
			CompanyAPIKey: "bkps_" + companyID, // or use your GenerateAPIKey
			// You can set quotas / AWS fields if you want
			CreatedAt: time.Now(),
		}

		if err := db.Create(&defaultCompany).Error; err != nil {
			log.Fatalf("failed to create default company: %v", err)
		}
		log.Printf("Created default company with ID=%s", defaultCompany.ID)
	} else {
		log.Printf("Found existing default company with ID=%s", defaultCompany.ID)
	}

	// Helper to create a user if not already present
	createUserIfNotExists := func(email, password, role string, companyID *string) {
		var existing user.User
		if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
			log.Printf("User %s already exists with ID=%s, skipping", email, existing.ID)
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", email, err)
		}

		u := user.User{
			ID:           utils.GenerateID(),
			Email:        email,
			PasswordHash: hash,
			CompanyID:    companyID,
			Role:         role,
		}

		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("failed to create user %s: %v", email, err)
		}

		log.Printf("Created %s user: email=%s id=%s company_id=%v", role, email, u.ID, companyID)
	}

	// superadmin: no company
	createUserIfNotExists(superadminEmail, superadminPassword, "superadmin", nil)

	// admin: assigned to defaultCompany
	createUserIfNotExists(adminEmail, adminPassword, "admin", &defaultCompany.ID)

	// normal user: assigned to defaultCompany
	createUserIfNotExists(userEmail, userPassword, "user", &defaultCompany.ID)

	log.Println("Seed complete")
}
