package main

import (
	"ai-review-system/api/events"
	"ai-review-system/pkg/config"
	"ai-review-system/pkg/database"
	"ai-review-system/pkg/httpx"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/pkg/messaging"
	"ai-review-system/services/review/internal/handler"
	"ai-review-system/services/review/internal/migrate"
	"ai-review-system/services/review/internal/model"
	"ai-review-system/services/review/internal/repository"
	"ai-review-system/services/review/internal/service"
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c, e := config.Load("review")
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
	p, e := messaging.NewProducer(c.KafkaBrokers)
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	publish := func(kind string, r *model.Review) error {
		id := fmt.Sprintf("%d-%d", r.ID, time.Now().UnixNano())
		return p.Publish("review.events", id, events.ReviewEvent{EventID: id, EventType: kind, OccurredAt: time.Now().Unix(), Source: "review-service", Payload: events.ReviewPayload{ReviewID: r.ID, ShopID: r.ShopID, Score: r.Score}})
	}
	tokens, e := jwtpkg.NewManager(c.JWTSecret, 24*time.Hour)
	if e != nil {
		log.Fatal(e)
	}
	router := handler.New(service.NewReviewService(repository.NewReviewRepository(db), publish), tokens).Router()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if e = httpx.Serve(ctx, c.Port, router); e != nil {
		log.Fatal(e)
	}
}
