package cache

import (
	"container/list"
	"demoservice/internal/app"
	"demoservice/internal/config"
	"demoservice/internal/db"
	"log"
	"sync"

	json "github.com/goccy/go-json" // ускоряем работу с json
)

// cacheEntry хранит и объект, и его готовый JSON — чтобы не сериализовать повторно.
type cacheEntry struct {
	order    *app.Order
	jsonData []byte // Gредварительно сериализованный JSON
}

type OrderCache struct {
	orders     map[string]*cacheEntry
	orderQueue *list.List
	maxsize    int
	mu         sync.RWMutex
}

func NewOrderCache(conf *config.Config) *OrderCache {
	return &OrderCache{
		orders:     make(map[string]*cacheEntry),
		orderQueue: list.New(),
		maxsize:    conf.CacheSize,
	}
}

func (c *OrderCache) Get(uid string) (*app.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.orders[uid]
	if !ok {
		return nil, false
	}
	return entry.order, true
}

func (c *OrderCache) GetJSON(uid string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.orders[uid]
	if !ok {
		return nil, false
	}
	return entry.jsonData, true
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

	// Сериализуем JSON один раз при добавлении в кэш
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("OrderCache: json marshal error for %s: %v", order.OrderUid, err)
		data = nil
	}

	c.orders[order.OrderUid] = &cacheEntry{order: order, jsonData: data}
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

func (c *OrderCache) UploadFromDb(repo db.Repository, conf *config.Config) error {
	CacheTmp, err := repo.CacheLoad(conf)

	if err != nil {
		return err
	}
	for _, elem := range CacheTmp {
		c.Set(elem)
	}

	return nil
}
