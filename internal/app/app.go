package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-review-system/internal/config"
	"ai-review-system/internal/database"
	"ai-review-system/internal/handler"
	"ai-review-system/internal/repository"
	"ai-review-system/internal/service"
	jwtpkg "ai-review-system/pkg/jwt"
)

func Run(role string) {
	if role != "" {
		_ = os.Setenv("SERVICE_ROLE", role)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}
	if err := database.InitRedis(); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var consumer *database.OrderConsumer
	if role == "order" || role == "seckill" || role == "" {
		if err := database.InitKafka(); err != nil {
			log.Fatal(err)
		}
		consumer, err = database.StartOrderConsumer(ctx, service.ProcessOrderMessage)
		if err != nil {
			log.Fatal(err)
		}
	}
	tokens, err := jwtpkg.NewManager(cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		log.Fatal(err)
	}
	userHandler := handler.NewUserHandler(service.NewUserService(repository.NewUserRepository(database.GetDB())), tokens)
	router := handler.NewRouterForService(role, handler.Dependencies{User: userHandler, Tokens: tokens})
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("[%s] listening on %s", role, server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	if consumer != nil {
		_ = consumer.Close()
	}
	_ = database.CloseKafka()
	_ = database.CloseRedis()
	if sqlDB, err := database.GetDB().DB(); err == nil {
		_ = sqlDB.Close()
	}
}
