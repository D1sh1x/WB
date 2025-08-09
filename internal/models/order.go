package models

import (
	"time"

	"gorm.io/gorm"
)

// Order - основная модель заказа
type Order struct {
	gorm.Model
	OrderUID          string `gorm:"not null"`
	TrackNumber       string
	Entry             string
	Delivery          Delivery `gorm:"foreignKey:OrderID"`
	Payment           Payment  `gorm:"foreignKey:OrderID"`
	Items             []Item   `gorm:"foreignKey:OrderID"`
	Locale            string
	InternalSignature string
	CustomerID        string
	DeliveryService   string
	ShardKey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}

// Delivery - модель доставки
type Delivery struct {
	gorm.Model
	OrderID uint `gorm:"not null"`
	Name    string
	Phone   string
	Zip     string
	City    string
	Address string
	Region  string
	Email   string
}

// Payment - модель платежа
type Payment struct {
	gorm.Model
	OrderID      uint
	Transaction  string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDt    int64
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}

// Item - модель товара в заказе
type Item struct {
	gorm.Model
	OrderID     uint `gorm:"not null"`
	ChrtID      int
	TrackNumber string
	Price       int
	Rid         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmID        int
	Brand       string
	Status      int
}
