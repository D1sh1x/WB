package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"WB2/internal/dto/response"
	"WB2/internal/kafka"

	"github.com/google/uuid"
)

func main() {

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "orders"
	}
	brokers := strings.Split(brokersEnv, ",")

	payload := response.OrderResponse{
		OrderUID:        uuid.NewString(),
		TrackNumber:     "ANDREY",
		Entry:           "ASS",
		Locale:          "ru",
		CustomerID:      "test-customer",
		DeliveryService: "meest",
		ShardKey:        "9",
		SmID:            99,
		DateCreated:     time.Now(),
		OofShard:        "1",
		Delivery: response.DeliveryResponse{
			Name:    "Test User",
			Phone:   "+79000000000",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Tverskaya 1",
			Region:  "Moscow",
			Email:   "test@example.com",
		},
		Payment: response.PaymentResponse{
			Transaction:  uuid.NewString(),
			Currency:     "RUB",
			Provider:     "wbpay",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Sber",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []response.ItemResponse{
			{
				ChrtID:      123,
				TrackNumber: "WBILMTESTTRACK",
				Price:       900,
				Rid:         "ab123",
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  900,
				NmID:        1000,
				Brand:       "WB",
				Status:      202,
			},
		},
	}

	if err := kafka.ProduceTestMessage(context.Background(), brokers, topic, payload); err != nil {
		log.Fatalf("failed to produce message: %v", err)
	}
	log.Printf("produced test order to topic %s", topic)
}
