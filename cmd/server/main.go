package main

import (
	"log"
	"os"
	"strings"

	"ai-review-system/internal/database"
	"ai-review-system/internal/handler"
	"ai-review-system/internal/service"
)

func main() {
	serviceRole := strings.ToLower(os.Getenv("SERVICE_ROLE"))

	// Initialize MySQL
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	if err := database.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Kafka only for monolith/order service roles.
	if serviceRole == "" || serviceRole == "order" || serviceRole == "seckill" {
		if err := database.InitKafka(); err != nil {
			log.Fatalf("Failed to initialize Kafka: %v", err)
		}

		if err := database.StartOrderConsumer(service.ProcessOrderMessage); err != nil {
			log.Fatalf("Failed to start Kafka consumer: %v", err)
		}
	}

	// Get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := handler.NewRouterForService(serviceRole)

	if serviceRole == "" {
		serviceRole = "monolith"
	}
	log.Printf("[ai-review-system] %s service starting on port %s", serviceRole, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
