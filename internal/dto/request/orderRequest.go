package request

import (
	"WB2/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateOrderRequest - DTO для создания заказа
type CreateOrderRequest struct {
	TrackNumber       string                `json:"track_number" validate:"required"`
	Entry             string                `json:"entry" validate:"required"`
	Delivery          CreateDeliveryRequest `json:"delivery" validate:"required"`
	Payment           CreatePaymentRequest  `json:"payment" validate:"required"`
	Items             []CreateItemRequest   `json:"items" validate:"required,min=1"`
	Locale            string                `json:"locale"`
	InternalSignature string                `json:"internal_signature"`
	CustomerID        string                `json:"customer_id" validate:"required"`
	DeliveryService   string                `json:"delivery_service"`
	ShardKey          string                `json:"shardkey"`
	SmID              int                   `json:"sm_id"`
	OofShard          string                `json:"oof_shard"`
}

// CreateDeliveryRequest - DTO для данных доставки
type CreateDeliveryRequest struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required"`
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region"`
	Email   string `json:"email" validate:"email"`
}

// CreatePaymentRequest - DTO для данных платежа
type CreatePaymentRequest struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" validate:"required"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,min=1"`
	PaymentDt    int64  `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" validate:"min=0"`
}

// CreateItemRequest - DTO для товара в заказе
type CreateItemRequest struct {
	ChrtID      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"required,min=1"`
	Rid         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"min=0"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price" validate:"required,min=1"`
	NmID        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}

// Validate валидация CreateOrderRequest
func (req *CreateOrderRequest) Validate() error {
	if req.TrackNumber == "" || req.Entry == "" || req.CustomerID == "" {
		return errors.New("track_number, entry, customer_id are required")
	}
	if req.Delivery.Name == "" || req.Delivery.Phone == "" || req.Delivery.Zip == "" || req.Delivery.City == "" || req.Delivery.Address == "" {
		return errors.New("delivery.name, phone, zip, city, address are required")
	}
	if req.Payment.Transaction == "" || req.Payment.Currency == "" || req.Payment.Provider == "" {
		return errors.New("payment.transaction, currency, provider are required")
	}
	if req.Payment.Amount < 1 {
		return fmt.Errorf("payment.amount must be >= 1")
	}
	if req.Payment.PaymentDt <= 0 {
		return fmt.Errorf("payment.payment_dt must be > 0")
	}
	if req.Payment.DeliveryCost < 0 || req.Payment.GoodsTotal < 0 || req.Payment.CustomFee < 0 {
		return fmt.Errorf("payment costs must be >= 0")
	}
	if len(req.Items) == 0 {
		return errors.New("items must contain at least 1 item")
	}
	for i, it := range req.Items {
		if it.ChrtID == 0 || it.TrackNumber == "" || it.Price < 1 || it.Rid == "" || it.Name == "" || it.TotalPrice < 1 || it.NmID == 0 || it.Brand == "" || it.Status == 0 {
			return fmt.Errorf("invalid item at index %d", i)
		}
		if it.Sale < 0 {
			return fmt.Errorf("item.sale must be >= 0 at index %d", i)
		}
	}
	return nil
}

// UpdateOrderRequest - DTO для обновления заказа
type UpdateOrderRequest struct {
	OrderUID          string                 `json:"order_uid" validate:"required"`
	TrackNumber       string                 `json:"track_number"`
	Entry             string                 `json:"entry"`
	Delivery          *UpdateDeliveryRequest `json:"delivery,omitempty"`
	Payment           *UpdatePaymentRequest  `json:"payment,omitempty"`
	Items             []UpdateItemRequest    `json:"items,omitempty"`
	Locale            string                 `json:"locale"`
	InternalSignature string                 `json:"internal_signature"`
	CustomerID        string                 `json:"customer_id"`
	DeliveryService   string                 `json:"delivery_service"`
	ShardKey          string                 `json:"shardkey"`
	SmID              int                    `json:"sm_id"`
	OofShard          string                 `json:"oof_shard"`
}

// UpdateDeliveryRequest - DTO для обновления доставки
type UpdateDeliveryRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email" validate:"omitempty,email"`
}

// UpdatePaymentRequest - DTO для обновления платежа
type UpdatePaymentRequest struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount" validate:"omitempty,min=1"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" validate:"min=0"`
}

// UpdateItemRequest - DTO для обновления товара
type UpdateItemRequest struct {
	ID          uint   `json:"id"`
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price" validate:"omitempty,min=1"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale" validate:"min=0"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price" validate:"omitempty,min=1"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

// Validate валидация UpdateOrderRequest (частичная)
func (req *UpdateOrderRequest) Validate() error {
	if req.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if req.Payment != nil {
		if req.Payment.Amount != 0 && req.Payment.Amount < 1 {
			return errors.New("payment.amount must be >= 1 if provided")
		}
		if req.Payment.DeliveryCost < 0 || req.Payment.GoodsTotal < 0 || req.Payment.CustomFee < 0 {
			return errors.New("payment costs must be >= 0")
		}
	}
	if len(req.Items) > 0 {
		for i, it := range req.Items {
			if it.Price != 0 && it.Price < 1 {
				return fmt.Errorf("item.price must be >= 1 at index %d", i)
			}
			if it.TotalPrice != 0 && it.TotalPrice < 1 {
				return fmt.Errorf("item.total_price must be >= 1 at index %d", i)
			}
			if it.Sale < 0 {
				return fmt.Errorf("item.sale must be >= 0 at index %d", i)
			}
		}
	}
	return nil
}

// ToOrderModel преобразует CreateOrderRequest в models.Order
func (req *CreateOrderRequest) ToOrderModel() *models.Order {
	order := &models.Order{
		OrderUID:          uuid.NewString(),
		TrackNumber:       req.TrackNumber,
		Entry:             req.Entry,
		Locale:            req.Locale,
		InternalSignature: req.InternalSignature,
		CustomerID:        req.CustomerID,
		DeliveryService:   req.DeliveryService,
		ShardKey:          req.ShardKey,
		SmID:              req.SmID,
		DateCreated:       time.Now(),
		OofShard:          req.OofShard,
	}

	// Преобразуем доставку
	order.Delivery = models.Delivery{
		Name:    req.Delivery.Name,
		Phone:   req.Delivery.Phone,
		Zip:     req.Delivery.Zip,
		City:    req.Delivery.City,
		Address: req.Delivery.Address,
		Region:  req.Delivery.Region,
		Email:   req.Delivery.Email,
	}

	// Преобразуем платеж
	order.Payment = models.Payment{
		Transaction:  req.Payment.Transaction,
		RequestID:    req.Payment.RequestID,
		Currency:     req.Payment.Currency,
		Provider:     req.Payment.Provider,
		Amount:       req.Payment.Amount,
		PaymentDt:    req.Payment.PaymentDt,
		Bank:         req.Payment.Bank,
		DeliveryCost: req.Payment.DeliveryCost,
		GoodsTotal:   req.Payment.GoodsTotal,
		CustomFee:    req.Payment.CustomFee,
	}

	// Преобразуем товары
	order.Items = make([]models.Item, len(req.Items))
	for i, item := range req.Items {
		order.Items[i] = models.Item{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		}
	}

	return order
}

// UpdateOrderModel обновляет models.Order из UpdateOrderRequest
func (req *UpdateOrderRequest) UpdateOrderModel(order *models.Order) {
	if req.TrackNumber != "" {
		order.TrackNumber = req.TrackNumber
	}
	if req.Entry != "" {
		order.Entry = req.Entry
	}
	if req.Locale != "" {
		order.Locale = req.Locale
	}
	if req.InternalSignature != "" {
		order.InternalSignature = req.InternalSignature
	}
	if req.CustomerID != "" {
		order.CustomerID = req.CustomerID
	}
	if req.DeliveryService != "" {
		order.DeliveryService = req.DeliveryService
	}
	if req.ShardKey != "" {
		order.ShardKey = req.ShardKey
	}
	if req.SmID != 0 {
		order.SmID = req.SmID
	}
	if req.OofShard != "" {
		order.OofShard = req.OofShard
	}

	// Обновляем доставку
	if req.Delivery != nil {
		if req.Delivery.Name != "" {
			order.Delivery.Name = req.Delivery.Name
		}
		if req.Delivery.Phone != "" {
			order.Delivery.Phone = req.Delivery.Phone
		}
		if req.Delivery.Zip != "" {
			order.Delivery.Zip = req.Delivery.Zip
		}
		if req.Delivery.City != "" {
			order.Delivery.City = req.Delivery.City
		}
		if req.Delivery.Address != "" {
			order.Delivery.Address = req.Delivery.Address
		}
		if req.Delivery.Region != "" {
			order.Delivery.Region = req.Delivery.Region
		}
		if req.Delivery.Email != "" {
			order.Delivery.Email = req.Delivery.Email
		}
	}

	// Обновляем платеж
	if req.Payment != nil {
		if req.Payment.Transaction != "" {
			order.Payment.Transaction = req.Payment.Transaction
		}
		if req.Payment.RequestID != "" {
			order.Payment.RequestID = req.Payment.RequestID
		}
		if req.Payment.Currency != "" {
			order.Payment.Currency = req.Payment.Currency
		}
		if req.Payment.Provider != "" {
			order.Payment.Provider = req.Payment.Provider
		}
		if req.Payment.Amount != 0 {
			order.Payment.Amount = req.Payment.Amount
		}
		if req.Payment.PaymentDt != 0 {
			order.Payment.PaymentDt = req.Payment.PaymentDt
		}
		if req.Payment.Bank != "" {
			order.Payment.Bank = req.Payment.Bank
		}
		if req.Payment.DeliveryCost >= 0 {
			order.Payment.DeliveryCost = req.Payment.DeliveryCost
		}
		if req.Payment.GoodsTotal >= 0 {
			order.Payment.GoodsTotal = req.Payment.GoodsTotal
		}
		if req.Payment.CustomFee >= 0 {
			order.Payment.CustomFee = req.Payment.CustomFee
		}
	}

	// Обновляем товары (если переданы)
	if len(req.Items) > 0 {
		order.Items = make([]models.Item, len(req.Items))
		for i, item := range req.Items {
			order.Items[i] = models.Item{
				Model:       models.Item{}.Model,
				OrderID:     order.ID,
				ChrtID:      item.ChrtID,
				TrackNumber: item.TrackNumber,
				Price:       item.Price,
				Rid:         item.Rid,
				Name:        item.Name,
				Sale:        item.Sale,
				Size:        item.Size,
				TotalPrice:  item.TotalPrice,
				NmID:        item.NmID,
				Brand:       item.Brand,
				Status:      item.Status,
			}
			if item.ID != 0 {
				order.Items[i].ID = item.ID
			}
		}
	}
}
