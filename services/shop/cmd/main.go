package main

import (
	"ai-review-system/pkg/config"
	"ai-review-system/pkg/database"
	"ai-review-system/pkg/httpx"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/services/shop/internal/handler"
	"ai-review-system/services/shop/internal/migrate"
	"ai-review-system/services/shop/internal/repository"
	"ai-review-system/services/shop/internal/service"
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c, e := config.Load("shop")
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
	rdb := redis.NewClient(&redis.Options{Addr: c.RedisHost + ":" + c.RedisPort, Password: c.RedisPassword})
	defer rdb.Close()
	tokens, e := jwtpkg.NewManager(c.JWTSecret, 24*time.Hour)
	if e != nil {
		log.Fatal(e)
	}
	router := handler.New(service.NewShopService(repository.NewShopRepository(db), rdb), tokens).Router()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if e = httpx.Serve(ctx, c.Port, router); e != nil {
		log.Fatal(e)
	}
}
