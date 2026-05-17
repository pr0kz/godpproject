package database

import (
	"log"

	"ai-review-system/internal/model"
)

// Migrate runs all database migrations
func Migrate() error {
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Shop{},
		&model.Review{},
		&model.Like{},
		&model.Coupon{},
		&model.Order{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("[database] Migrations completed successfully")
	return nil
}
