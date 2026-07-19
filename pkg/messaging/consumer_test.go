package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

type fakeSession struct {
	ctx    context.Context
	marked []*sarama.ConsumerMessage
}

func (s *fakeSession) Claims() map[string][]int32               { return nil }
func (s *fakeSession) MemberID() string                         { return "test" }
func (s *fakeSession) GenerationID() int32                      { return 1 }
func (s *fakeSession) MarkOffset(string, int32, int64, string)  {}
func (s *fakeSession) Commit()                                  {}
func (s *fakeSession) ResetOffset(string, int32, int64, string) {}
func (s *fakeSession) MarkMessage(m *sarama.ConsumerMessage, _ string) {
	s.marked = append(s.marked, m)
}
func (s *fakeSession) Context() context.Context { return s.ctx }

type fakeClaim struct{ messages chan *sarama.ConsumerMessage }

func (c *fakeClaim) Topic() string                            { return "orders.created" }
func (c *fakeClaim) Partition() int32                         { return 0 }
func (c *fakeClaim) InitialOffset() int64                     { return 0 }
func (c *fakeClaim) HighWaterMarkOffset() int64               { return 1 }
func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.messages }

func TestConsumerMarksMessageOnlyAfterSuccessfulHandling(t *testing.T) {
	message := &sarama.ConsumerMessage{Topic: "orders.created", Value: []byte(`{"user_id":1,"coupon_id":2}`)}
	claim := &fakeClaim{messages: make(chan *sarama.ConsumerMessage, 1)}
	claim.messages <- message
	close(claim.messages)
	session := &fakeSession{ctx: context.Background()}

	handler := consumerGroupHandler{handler: func(context.Context, []byte) error { return nil }}
	if err := handler.ConsumeClaim(session, claim); err != nil {
		t.Fatal(err)
	}
	if len(session.marked) != 1 || session.marked[0] != message {
		t.Fatal("successful message was not committed")
	}
}

func TestConsumerDoesNotMarkFailedMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	claim := &fakeClaim{messages: make(chan *sarama.ConsumerMessage, 1)}
	claim.messages <- &sarama.ConsumerMessage{Topic: "orders.created", Value: []byte("bad")}
	session := &fakeSession{ctx: ctx}
	expected := errors.New("database unavailable")

	handler := consumerGroupHandler{handler: func(context.Context, []byte) error { return expected }}
	if err := handler.ConsumeClaim(session, claim); !errors.Is(err, expected) {
		t.Fatalf("expected handler error, got %v", err)
	}
	if len(session.marked) != 0 {
		t.Fatal("failed message must not be committed")
	}
}
