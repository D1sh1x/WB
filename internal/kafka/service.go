package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"WB2/internal/config"
	reqdto "WB2/internal/dto/request"
	"WB2/internal/models"
	storagepkg "WB2/internal/storage/postgres"

	"gorm.io/gorm"
)

// cacheEntry хранит заказ и момент добавления в кэш для вычисления TTL.
type cacheEntry struct {
	order   *models.Order
	addedAt time.Time
}

// Cache хранит заказы в памяти по ключу order_uid и TTL для авто-протухания.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]cacheEntry
	ttl    time.Duration
}

// NewCache создаёт кэш с указанным TTL (0 = без протухания).
func NewCache(ttl time.Duration) *Cache {
	return &Cache{orders: make(map[string]cacheEntry), ttl: ttl}
}

type Service struct {
	producer *Producer
	consumer *Consumer
	config   *config.Config
	mu       sync.RWMutex

	storage *storagepkg.Storage
	cache   *Cache
}

func NewService(cfg *config.Config, storage *storagepkg.Storage) (*Service, error) {
	producer, err := NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Consumer будет передавать сырые данные в messageHandler (парсинг + сохранение + кэш)
	svc := &Service{
		producer: producer,
		config:   cfg,
		storage:  storage,
		cache:    NewCache(cfg.Cache.TTL),
	}

	consumer, err := NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, svc.messageHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	svc.consumer = consumer
	return svc, nil
}

func (s *Service) SendMessage(data string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Data:      data,
		Timestamp: time.Now(),
	}

	return s.producer.SendMessage(msg)
}

func (s *Service) StartConsumer(ctx context.Context) error {
	log.Printf("Starting Kafka consumer for topic: %s, group: %s", s.config.Kafka.Topic, s.config.Kafka.GroupID)
	return s.consumer.Start(ctx)
}

func (s *Service) Close() error {
	var errs []error

	if err := s.producer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close producer: %w", err))
	}

	if err := s.consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close consumer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing kafka service: %v", errs)
	}

	return nil
}

// messageHandler получает JSON заказа, валидирует, сохраняет в БД и обновляет кэш.
func (s *Service) messageHandler(data []byte) error {
	// Парсим входящее сообщение в DTO
	var dto reqdto.CreateOrderRequest
	if err := json.Unmarshal(data, &dto); err != nil {
		return fmt.Errorf("unmarshal dto: %w", err)
	}

	// Преобразуем в модель
	order := dto.ToOrderModel()

	// Транзакция: проверяем дубликат по order_uid и создаем новую запись, если её ещё нет
	err := s.storage.Db.Transaction(func(tx *gorm.DB) error {
		var existing models.Order
		if err := tx.Where("order_uid = ?", order.OrderUID).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Создаём новый заказ вместе с ассоциациями
				if err := tx.Create(order).Error; err != nil {
					return fmt.Errorf("create order: %w", err)
				}
				return nil
			}
			return fmt.Errorf("find existing order: %w", err)
		}

		// Дубликат — пока пропускаем (можно расширить до обновления/merge)
		return nil
	})
	if err != nil {
		return err
	}

	// Обновление кэша (пишем с отметкой времени)
	s.cache.mu.Lock()
	s.cache.orders[order.OrderUID] = cacheEntry{order: order, addedAt: time.Now()}
	s.cache.mu.Unlock()

	return nil
}

// Прогрев кэша из БД при старте
func (s *Service) WarmUpCache() error {
	var orders []models.Order
	if err := s.storage.Db.Preload("Delivery").Preload("Payment").Preload("Items").Find(&orders).Error; err != nil {
		return fmt.Errorf("load orders for cache: %w", err)
	}
	s.cache.mu.Lock()
	now := time.Now()
	for i := range orders {
		o := orders[i]
		s.cache.orders[o.OrderUID] = cacheEntry{order: &o, addedAt: now}
	}
	s.cache.mu.Unlock()
	return nil
}

// Вспомогательная настройка полной записи связей
var gormSessionFullSave = gorm.Session{FullSaveAssociations: true}

// GetOrderFromCache возвращает заказ из кэша по UID
func (s *Service) GetOrderFromCache(orderUID string) (*models.Order, bool) {
	s.cache.mu.RLock()
	entry, ok := s.cache.orders[orderUID]
	s.cache.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if s.isExpired(entry.addedAt) {
		s.cache.mu.Lock()
		delete(s.cache.orders, orderUID)
		s.cache.mu.Unlock()
		return nil, false
	}
	return entry.order, true
}

// GetAllFromCache возвращает все заказы из кэша (снимок на момент вызова)
func (s *Service) GetAllFromCache() []models.Order {
	now := time.Now()
	s.cache.mu.RLock()
	res := make([]models.Order, 0, len(s.cache.orders))
	for _, entry := range s.cache.orders {
		if s.cache.ttl > 0 && now.Sub(entry.addedAt) > s.cache.ttl {
			// пропустим протухшее; удалим ниже под lock на запись
			continue
		}
		res = append(res, *entry.order)
	}
	s.cache.mu.RUnlock()

	// Очистим протухшие записи в отдельной фазе
	s.evictExpired()
	return res
}

// StartCacheJanitor запускает фоновую очистку протухших записей с периодом.
func (s *Service) StartCacheJanitor(ctx context.Context) {
	if s.cache.ttl <= 0 {
		return
	}
	interval := s.cache.ttl / 2
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.evictExpired()
			}
		}
	}()
}

// evictExpired удаляет из кэша все записи, чьё время жизни истекло.
func (s *Service) evictExpired() {
	if s.cache.ttl <= 0 {
		return
	}
	now := time.Now()
	s.cache.mu.Lock()
	for uid, entry := range s.cache.orders {
		if now.Sub(entry.addedAt) > s.cache.ttl {
			delete(s.cache.orders, uid)
		}
	}
	s.cache.mu.Unlock()
}

// isExpired проверяет, истёк ли TTL для записи.
func (s *Service) isExpired(addedAt time.Time) bool {
	return s.cache.ttl > 0 && time.Since(addedAt) > s.cache.ttl
}
