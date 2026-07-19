package main

import (
	"ai-review-system/pkg/config"
	"ai-review-system/pkg/database"
	"ai-review-system/pkg/httpx"
	jwtpkg "ai-review-system/pkg/jwt"
	"ai-review-system/pkg/messaging"
	"ai-review-system/services/order/internal/handler"
	"ai-review-system/services/order/internal/migrate"
	"ai-review-system/services/order/internal/repository"
	"ai-review-system/services/order/internal/service"
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c, e := config.Load("order")
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
	p, e := messaging.NewProducer(c.KafkaBrokers)
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	consumer, e := messaging.NewConsumer(c.KafkaBrokers, "order-service", "orders.created")
	if e != nil {
		log.Fatal(e)
	}
	defer consumer.Close()
	publish := func(ctx context.Context, u, coupon uint) error {
		return p.Publish("orders.created", time.Now().String(), service.OrderCreatedEvent{UserID: u, CouponID: coupon})
	}
	orders := repository.NewOrderRepository(db)
	s := service.NewSeckillService(repository.NewCouponRepository(db, rdb), orders, rdb, publish)
	tokens, e := jwtpkg.NewManager(c.JWTSecret, 24*time.Hour)
	if e != nil {
		log.Fatal(e)
	}
	router := handler.New(s, tokens).Router()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumerErr := make(chan error, 1)
	go func() { consumerErr <- consumer.Run(ctx, service.ProcessOrderMessage(orders)) }()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-consumer.Errors():
				if !ok {
					return
				}
				log.Printf("Kafka consumer error: %v", err)
			}
		}
	}()

	serverErr := make(chan error, 1)
	go func() { serverErr <- httpx.Serve(ctx, c.Port, router) }()
	select {
	case e = <-consumerErr:
		if e != nil {
			stop()
			log.Fatal(e)
		}
	case e = <-serverErr:
		if e != nil {
			stop()
			log.Fatal(e)
		}
	case <-ctx.Done():
	}
}
