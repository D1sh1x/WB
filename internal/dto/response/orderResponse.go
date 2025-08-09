package response

import (
	"WB2/internal/models"
	"time"
)

// OrderResponse - DTO для ответа с заказом
type OrderResponse struct {
	ID                uint             `json:"id"`
	OrderUID          string           `json:"order_uid"`
	TrackNumber       string           `json:"track_number"`
	Entry             string           `json:"entry"`
	Delivery          DeliveryResponse `json:"delivery"`
	Payment           PaymentResponse  `json:"payment"`
	Items             []ItemResponse   `json:"items"`
	Locale            string           `json:"locale"`
	InternalSignature string           `json:"internal_signature"`
	CustomerID        string           `json:"customer_id"`
	DeliveryService   string           `json:"delivery_service"`
	ShardKey          string           `json:"shardkey"`
	SmID              int              `json:"sm_id"`
	DateCreated       time.Time        `json:"date_created"`
	OofShard          string           `json:"oof_shard"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// DeliveryResponse - DTO для ответа с данными доставки
type DeliveryResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

// PaymentResponse - DTO для ответа с данными платежа
type PaymentResponse struct {
	ID           uint   `json:"id"`
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// ItemResponse - DTO для ответа с данными товара
type ItemResponse struct {
	ID          uint   `json:"id"`
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

// GetAllOrdersResponse - DTO для ответа со списком заказов
type GetAllOrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
}

// ErrorResponse - DTO для ошибок
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse - DTO для успешных операций
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DeleteOrderRequest - DTO для удаления заказа
type DeleteOrderRequest struct {
	OrderUID string `json:"order_uid" validate:"required"`
}

// ToOrderResponse преобразует models.Order в OrderResponse
func ToOrderResponse(order *models.Order) *OrderResponse {
	response := &OrderResponse{
		ID:                order.ID,
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		ShardKey:          order.ShardKey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}

	// Преобразуем доставку
	response.Delivery = DeliveryResponse{
		ID:      order.Delivery.ID,
		Name:    order.Delivery.Name,
		Phone:   order.Delivery.Phone,
		Zip:     order.Delivery.Zip,
		City:    order.Delivery.City,
		Address: order.Delivery.Address,
		Region:  order.Delivery.Region,
		Email:   order.Delivery.Email,
	}

	// Преобразуем платеж
	response.Payment = PaymentResponse{
		ID:           order.Payment.ID,
		Transaction:  order.Payment.Transaction,
		RequestID:    order.Payment.RequestID,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDt:    order.Payment.PaymentDt,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
		CustomFee:    order.Payment.CustomFee,
	}

	// Преобразуем товары
	response.Items = make([]ItemResponse, len(order.Items))
	for i, item := range order.Items {
		response.Items[i] = ItemResponse{
			ID:          item.ID,
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

	return response
}

// ToOrderResponseList преобразует список models.Order в GetAllOrdersResponse
func ToOrderResponseList(orders []models.Order, total, page, limit int) *GetAllOrdersResponse {
	response := &GetAllOrdersResponse{
		Orders: make([]OrderResponse, len(orders)),
		Total:  total,
	}

	for i, order := range orders {
		response.Orders[i] = *ToOrderResponse(&order)
	}

	return response
}
