package messaging

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

type MessageHandler func(context.Context, []byte) error

type Consumer struct {
	group  sarama.ConsumerGroup
	topics []string
}

func NewConsumer(brokers []string, groupID string, topics ...string) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if groupID == "" {
		return nil, errors.New("consumer group ID is required")
	}
	if len(topics) == 0 {
		return nil, errors.New("at least one Kafka topic is required")
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}
	return &Consumer{group: group, topics: topics}, nil
}

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) error {
	if handler == nil {
		return errors.New("message handler is required")
	}

	h := consumerGroupHandler{handler: handler}
	for ctx.Err() == nil {
		if err := c.group.Consume(ctx, c.topics, h); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("consume Kafka messages: %w", err)
		}
	}
	return nil
}

func (c *Consumer) Errors() <-chan error { return c.group.Errors() }
func (c *Consumer) Close() error         { return c.group.Close() }

type consumerGroupHandler struct{ handler MessageHandler }

func (consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := h.handler(session.Context(), msg.Value); err != nil {
				return fmt.Errorf("handle topic %s partition %d offset %d: %w", msg.Topic, msg.Partition, msg.Offset, err)
			}
			session.MarkMessage(msg, "")
		}
	}
}
