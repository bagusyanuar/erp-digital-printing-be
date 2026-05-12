package main

import (
	"log"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/database"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Initialize Database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 3. Hash Password
	password := "@Administrator1234"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	// 4. Prepare User Data
	user := domain.User{
		Username: "administrator",
		Password: string(hashedPassword),
	}

	// 5. Seed with Upsert (On Conflict)
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"password", "updated_at"}),
	}).Create(&user).Error

	if err != nil {
		log.Fatalf("failed to seed user: %v", err)
	}

	log.Println("✅ Seeding completed: User 'administrator' created/updated successfully")
}
