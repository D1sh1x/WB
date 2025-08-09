package storage

import (
	"WB2/internal/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	Db *gorm.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	newLogger := logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(postgres.Open(storagePath), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Важно: сначала создаем orders, затем сущности с FK на неё
	if err := db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Item{}); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{Db: db}, nil
}

// CreateOrder создает новый заказ со всеми связанными данными и возвращает UID
func (s *Storage) CreateOrder(order *models.Order) (string, error) {
	if err := s.Db.Create(order).Error; err != nil {
		return "", err
	}
	return order.OrderUID, nil
}

// GetAllOrders получает все заказы
func (s *Storage) GetAllOrders() ([]models.Order, error) {
    var orders []models.Order
    if err := s.Db.Preload("Delivery").Preload("Payment").Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrderByID получает заказ по order_uid
func (s *Storage) GetOrderByUID(orderUID string) (*models.Order, error) {
	var order models.Order
	if err := s.Db.Where("order_uid = ?", orderUID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrder обновляет заказ
func (s *Storage) UpdateOrder(order *models.Order) (string, error) {
	return order.OrderUID, s.Db.Save(order).Error
}

// DeleteOrder удаляет заказ по OrderUID
func (s *Storage) DeleteOrder(orderUID string) (string, error) {
	return orderUID, s.Db.Where("order_uid = ?", orderUID).Delete(&models.Order{}).Error
}
