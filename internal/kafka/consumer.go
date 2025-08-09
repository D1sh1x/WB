package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	groupID  string
	handler  MessageHandler
}

// MessageHandler получает сырые байты сообщения из Kafka.
// Ожидается JSON заказа, который будет разобран в обработчике на уровне приложения.
type MessageHandler func(data []byte) error

func NewConsumer(brokers []string, topic, groupID string, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		groupID:  groupID,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := []string{c.topic}

	for {
		err := c.consumer.Consume(ctx, topics, c)
		if err != nil {
			log.Printf("Error from consumer: %v", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Передаем сырые данные в обработчик
		if err := c.handler(message.Value); err != nil {
			log.Printf("Failed to process message: %v", err)
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
