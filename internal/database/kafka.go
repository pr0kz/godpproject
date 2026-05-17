package database

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

const TopicSeckillOrder = "seckill_order"

var KafkaProducer sarama.SyncProducer

// InitKafka initializes the Kafka producer
func InitKafka() error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	brokerList := strings.Split(brokers, ",")

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return err
	}

	KafkaProducer = producer
	log.Println("[kafka] Producer connected successfully")
	return nil
}

// SendOrderMessage sends a seckill order message to Kafka
func SendOrderMessage(ctx context.Context, userID, couponID uint) error {
	type OrderMsg struct {
		UserID   uint  `json:"user_id"`
		CouponID uint  `json:"coupon_id"`
		Time     int64 `json:"time"`
	}

	msg := OrderMsg{
		UserID:   userID,
		CouponID: couponID,
		Time:     time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, _, err = KafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicSeckillOrder,
		Value: sarama.StringEncoder(data),
	})
	return err
}

// StartOrderConsumer starts a Kafka consumer that creates orders from messages
func StartOrderConsumer(handler func(userID, couponID uint)) error {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	brokerList := strings.Split(brokers, ",")

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		return err
	}

	partitionConsumer, err := consumer.ConsumePartition(TopicSeckillOrder, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}

	go func() {
		defer consumer.Close()
		defer partitionConsumer.Close()
		log.Println("[kafka] Consumer started, waiting for seckill order messages...")
		for msg := range partitionConsumer.Messages() {
			var payload struct {
				UserID   uint `json:"user_id"`
				CouponID uint `json:"coupon_id"`
			}
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Printf("[kafka] Failed to unmarshal message: %v", err)
				continue
			}
			handler(payload.UserID, payload.CouponID)
		}
	}()

	return nil
}
