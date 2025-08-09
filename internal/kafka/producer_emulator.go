package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// ProduceTestMessage отправляет в Kafka произвольный JSON заказа (как OrderResponse) для демонстрации
func ProduceTestMessage(ctx context.Context, brokers []string, topic string, payload any) error {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return err
	}
	defer producer.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", time.Now().UnixNano())),
		Value: sarama.ByteEncoder(data),
	}
	_, _, err = producer.SendMessage(msg)
	return err
}
