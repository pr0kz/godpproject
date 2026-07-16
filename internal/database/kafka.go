package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

const TopicSeckillOrder = "seckill_order"
const TopicSeckillOrderDLQ = "seckill_order_dlq"
const OrderConsumerGroup = "ai-review-order-service"

var KafkaProducer sarama.SyncProducer

type OrderMessage struct {
	EventID   string `json:"event_id"`
	UserID    uint   `json:"user_id"`
	CouponID  uint   `json:"coupon_id"`
	CreatedAt int64  `json:"created_at"`
	Attempt   int    `json:"attempt"`
}

func InitKafka() error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = 10 * time.Second
	producer, err := sarama.NewSyncProducer(kafkaBrokers(), config)
	if err != nil {
		return err
	}
	KafkaProducer = producer
	return nil
}
func SendOrderMessage(ctx context.Context, userID, couponID uint) error {
	return publishOrderMessage(TopicSeckillOrder, OrderMessage{EventID: newEventID(), UserID: userID, CouponID: couponID, CreatedAt: time.Now().Unix()})
}
func publishOrderMessage(topic string, payload OrderMessage) error {
	if KafkaProducer == nil {
		return errors.New("Kafka producer is not initialized")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _, err = KafkaProducer.SendMessage(&sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(payload.EventID), Value: sarama.ByteEncoder(data)})
	return err
}
func newEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
func kafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	return strings.Split(brokers, ",")
}

type OrderConsumer struct {
	group   sarama.ConsumerGroup
	handler func(context.Context, uint, uint) error
}

func StartOrderConsumer(ctx context.Context, handler func(context.Context, uint, uint) error) (*OrderConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	group, err := sarama.NewConsumerGroup(kafkaBrokers(), OrderConsumerGroup, cfg)
	if err != nil {
		return nil, err
	}
	consumer := &OrderConsumer{group: group, handler: handler}
	go consumer.run(ctx)
	return consumer, nil
}
func (c *OrderConsumer) run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := c.group.Consume(ctx, []string{TopicSeckillOrder}, c); err != nil {
			log.Printf("[kafka] consumer error: %v", err)
			time.Sleep(time.Second)
		}
	}
}
func (c *OrderConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *OrderConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (c *OrderConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			var payload OrderMessage
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Printf("[kafka] malformed message: %v", err)
				session.MarkMessage(msg, "malformed")
				continue
			}
			if err := c.processWithRetry(session.Context(), payload); err != nil {
				payload.Attempt++
				if dlqErr := publishOrderMessage(TopicSeckillOrderDLQ, payload); dlqErr != nil {
					return fmt.Errorf("handler failed: %v; DLQ failed: %w", err, dlqErr)
				}
			}
			session.MarkMessage(msg, "processed")
		}
	}
}
func (c *OrderConsumer) processWithRetry(ctx context.Context, payload OrderMessage) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if err = c.handler(ctx, payload.UserID, payload.CouponID); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 200 * time.Millisecond):
		}
	}
	return err
}
func (c *OrderConsumer) Close() error { return c.group.Close() }
func CloseKafka() error {
	if KafkaProducer != nil {
		return KafkaProducer.Close()
	}
	return nil
}
