package cache

import (
	"container/list"
	"demoservice/internal/app"
	"demoservice/internal/db"
	"demoservice/internal/config"
	"sync"
	"log"
)

type OrderCache struct{
	orders map[string]*app.Order
	orderQueue *list.List
	maxsize int
	mu sync.RWMutex
}

func NewOrderCache(conf *config.Config) *OrderCache{
	return &OrderCache{
		orders: make(map[string]*app.Order),
		orderQueue: list.New(),
		maxsize: conf.CacheSize,
	}
}

func (c *OrderCache) Get(uid string) (*app.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[uid]
	return order, ok
}

func (c *OrderCache) Set(order *app.Order) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if order.OrderUid == "" {
        log.Println("OrderCache: empty OrderUid, skipping")
        return
    }

    if _, exists := c.orders[order.OrderUid]; exists {
        return
    }

    c.orders[order.OrderUid] = order
    c.orderQueue.PushBack(order.OrderUid)

    if c.orderQueue.Len() > c.maxsize {
        oldest := c.orderQueue.Front()
        if oldest != nil {
            oldestValue, ok := oldest.Value.(string)
            if !ok {
                log.Printf("OrderCache: unexpected type in orderQueue: %T", oldest.Value)
            } else {
                delete(c.orders, oldestValue)
                c.orderQueue.Remove(oldest)
            }
        }
    }
}

func (c *OrderCache) UploadFromDb(repo db.Repository, conf *config.Config) error{
	CacheTmp, err := repo.CacheLoad(conf)

	if err !=nil{
		return err
	}
	for _, elem := range CacheTmp {
			c.Set(elem)
	}

	return nil
}