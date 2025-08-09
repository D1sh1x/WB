package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"WB2/internal/cache"
	"WB2/internal/dto/response"
	"WB2/internal/models"
	storage "WB2/internal/storage/postgres"

	"github.com/IBM/sarama"
)

type Consumer struct {
	log   *slog.Logger
	store *storage.Storage
	cache *cache.OrderCache
	group sarama.ConsumerGroup
	topic string
}

func NewConsumer(log *slog.Logger, store *storage.Storage, c *cache.OrderCache, brokers []string, groupID, topic string, version string) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	if v, err := sarama.ParseKafkaVersion(version); err == nil {
		cfg.Version = v
	}
	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		log:   log,
		store: store,
		cache: c,
		group: group,
		topic: topic,
	}, nil
}

func (c *Consumer) Close() error { return c.group.Close() }

func (c *Consumer) Run(ctx context.Context) error {
	handler := &consumerGroupHandler{log: c.log, store: c.store, cache: c.cache}
	for {
		if err := c.group.Consume(ctx, []string{c.topic}, handler); err != nil {
			c.log.Error("kafka consume error", slog.String("err", err.Error()))
			time.Sleep(2 * time.Second)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

type consumerGroupHandler struct {
	log   *slog.Logger
	store *storage.Storage
	cache *cache.OrderCache
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// ожидаем JSON с полями, соответствующими CreateOrderRequest (кроме автогенерируемого UID/DateCreated)
		var input response.OrderResponse
		if err := json.Unmarshal(msg.Value, &input); err != nil {
			h.log.Error("failed to decode kafka message", slog.String("err", err.Error()))
			sess.MarkMessage(msg, "bad_json")
			continue
		}

		order := toOrderModelFromResponse(&input)
		if err := h.store.Db.Create(order).Error; err != nil {
			h.log.Error("failed to save order", slog.String("err", err.Error()))
		} else {
			h.cache.Set(order)
			sess.MarkMessage(msg, "ok")
		}
	}
	return nil
}

func toOrderModelFromResponse(resp *response.OrderResponse) *models.Order {
	order := &models.Order{
		OrderUID:          resp.OrderUID,
		TrackNumber:       resp.TrackNumber,
		Entry:             resp.Entry,
		Locale:            resp.Locale,
		InternalSignature: resp.InternalSignature,
		CustomerID:        resp.CustomerID,
		DeliveryService:   resp.DeliveryService,
		ShardKey:          resp.ShardKey,
		SmID:              resp.SmID,
		DateCreated:       resp.DateCreated,
		OofShard:          resp.OofShard,
	}

	order.Delivery = models.Delivery{
		Name:    resp.Delivery.Name,
		Phone:   resp.Delivery.Phone,
		Zip:     resp.Delivery.Zip,
		City:    resp.Delivery.City,
		Address: resp.Delivery.Address,
		Region:  resp.Delivery.Region,
		Email:   resp.Delivery.Email,
	}

	order.Payment = models.Payment{
		Transaction:  resp.Payment.Transaction,
		RequestID:    resp.Payment.RequestID,
		Currency:     resp.Payment.Currency,
		Provider:     resp.Payment.Provider,
		Amount:       resp.Payment.Amount,
		PaymentDt:    resp.Payment.PaymentDt,
		Bank:         resp.Payment.Bank,
		DeliveryCost: resp.Payment.DeliveryCost,
		GoodsTotal:   resp.Payment.GoodsTotal,
		CustomFee:    resp.Payment.CustomFee,
	}

	order.Items = make([]models.Item, len(resp.Items))
	for i := range resp.Items {
		it := resp.Items[i]
		order.Items[i] = models.Item{
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			Rid:         it.Rid,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
			TotalPrice:  it.TotalPrice,
			NmID:        it.NmID,
			Brand:       it.Brand,
			Status:      it.Status,
		}
	}
	return order
}
