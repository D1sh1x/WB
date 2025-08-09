package cache

import (
	"context"
	"sync"
	"time"

	"WB2/internal/models"
)

type entry struct {
	order     *models.Order
	expiresAt time.Time
}

// OrderCache хранит заказы в памяти с синхронизацией и TTL
type OrderCache struct {
	mu    sync.RWMutex
	byUID map[string]entry
	ttl   time.Duration
}

func NewOrderCache(ttl time.Duration) *OrderCache {
	return &OrderCache{byUID: make(map[string]entry), ttl: ttl}
}

func (c *OrderCache) Set(order *models.Order) {
	if order == nil || order.OrderUID == "" {
		return
	}
	c.mu.Lock()
	c.byUID[order.OrderUID] = entry{order: order, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Delete удаляет заказ из кэша
func (c *OrderCache) Delete(orderUID string) {
	c.mu.Lock()
	delete(c.byUID, orderUID)
	c.mu.Unlock()
}

func (c *OrderCache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	e, ok := c.byUID[orderUID]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		// lazy expiration
		c.Delete(orderUID)
		return nil, false
	}
	return e.order, true
}

func (c *OrderCache) GetAll() []*models.Order {
	now := time.Now()
	c.mu.RLock()
	result := make([]*models.Order, 0, len(c.byUID))
	for _, e := range c.byUID {
		if now.After(e.expiresAt) {
			// пометим, удалим после снятия RLock
			// сбор просроченных сделаем вне цикла для эффективности
		} else {
			result = append(result, e.order)
		}
	}
	c.mu.RUnlock()
	// лёгкая очистка просроченных элементов
	c.cleanupExpired()
	return result
}

func (c *OrderCache) Load(orders []models.Order) {
	c.mu.Lock()
	for i := range orders {
		o := orders[i]
		if o.OrderUID != "" {
			order := o // локальная копия
			c.byUID[o.OrderUID] = entry{order: &order, expiresAt: time.Now().Add(c.ttl)}
		}
	}
	c.mu.Unlock()
}

// StartCleaner запускает фоновый процесс периодической очистки просроченных записей
func (c *OrderCache) StartCleaner(ctx context.Context) {
	if c.ttl <= 0 {
		return
	}
	// чистим примерно раз в половину TTL, но не реже раза в минуту
	interval := c.ttl / 2
	if interval <= 0 || interval > time.Minute {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.cleanupExpired()
			}
		}
	}()
}

func (c *OrderCache) cleanupExpired() {
	if c.ttl <= 0 {
		return
	}
	now := time.Now()
	c.mu.Lock()
	for uid, e := range c.byUID {
		if now.After(e.expiresAt) {
			delete(c.byUID, uid)
		}
	}
	c.mu.Unlock()
}
