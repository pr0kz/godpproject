package main

import (
	"ai-review-system/pkg/config"
	"ai-review-system/pkg/database"
	"ai-review-system/pkg/httpx"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/user/internal/handler"
	"ai-review-system/services/user/internal/migrate"
	"ai-review-system/services/user/internal/repository"
	"ai-review-system/services/user/internal/service"
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c, e := config.Load("user")
	if e != nil {
		log.Fatal(e)
	}
	db, e := database.Open(database.Config{User: c.DBUser, Password: c.DBPassword, Host: c.DBHost, Port: c.DBPort, Name: c.DBName})
	if e != nil {
		log.Fatal(e)
	}
	defer database.Close(db)
	if e = migrate.Run(db); e != nil {
		log.Fatal(e)
	}
	tokens, e := jwtpkg.NewManager(c.JWTSecret, 24*time.Hour)
	if e != nil {
		log.Fatal(e)
	}
	r := handler.New(service.New(repository.New(db)), tokens).Router()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if e = httpx.Serve(ctx, c.Port, r); e != nil {
		log.Fatal(e)
	}
}
