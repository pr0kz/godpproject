package main

import (
	"log"

	"ai-review-system/internal/config"
	"ai-review-system/internal/database"
)

func main() {
	if _, err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(); err != nil {
		log.Fatal(err)
	}
	sqlDB, err := database.GetDB().DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
