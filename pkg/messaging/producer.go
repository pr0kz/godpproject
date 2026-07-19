package messaging

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer struct{ p sarama.SyncProducer }

func NewProducer(b []string) (*Producer, error) {
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	c.Producer.RequiredAcks = sarama.WaitForAll
	p, e := sarama.NewSyncProducer(b, c)
	if e != nil {
		return nil, e
	}
	return &Producer{p: p}, nil
}
func (p *Producer) Publish(topic, key string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	_, _, e = p.p.SendMessage(&sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(b)})
	return e
}
func (p *Producer) Close() error { return p.p.Close() }
