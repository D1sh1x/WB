package cache

import (
	"sync"

	"WB2/internal/models"
)

// OrderCache хранит заказы в памяти с синхронизацией
type OrderCache struct {
	mu    sync.RWMutex
	byUID map[string]*models.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{byUID: make(map[string]*models.Order)}
}

func (c *OrderCache) Set(order *models.Order) {
	if order == nil || order.OrderUID == "" {
		return
	}
	c.mu.Lock()
	c.byUID[order.OrderUID] = order
	c.mu.Unlock()
}

func (c *OrderCache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	order, ok := c.byUID[orderUID]
	c.mu.RUnlock()
	return order, ok
}

func (c *OrderCache) GetAll() []*models.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*models.Order, 0, len(c.byUID))
	for _, o := range c.byUID {
		result = append(result, o)
	}
	return result
}

func (c *OrderCache) Load(orders []models.Order) {
	c.mu.Lock()
	for i := range orders {
		o := orders[i]
		if o.OrderUID != "" {
			// захватываем адрес конкретного элемента среза
			order := o
			c.byUID[o.OrderUID] = &order
		}
	}
	c.mu.Unlock()
}
