package cache

import (
	"testing"
	"demoservice/internal/app"
	"demoservice/internal/config"
	"github.com/stretchr/testify/assert"
)

// Mock PostgresRepo for UploadFromDb
type mockRepo struct {
	orders []*app.Order
}

func (m *mockRepo) CacheLoad(conf *config.Config) ([]*app.Order, error) {
	return m.orders, nil
}

func (m *mockRepo) Save(order *app.Order) error {
    return nil
}

func (m *mockRepo) Load(uid string) (*app.Order, error) {
    return nil, nil
}

func TestNewOrderCache(t *testing.T) {
	conf := &config.Config{CacheSize: 2}
	cache := NewOrderCache(conf)
	assert.NotNil(t, cache)
	assert.Equal(t, 2, cache.maxsize)
	assert.Equal(t, 0, cache.orderQueue.Len())
}

func TestSetAndGet(t *testing.T) {
	conf := &config.Config{CacheSize: 2}
	cache := NewOrderCache(conf)

	order1 := &app.Order{OrderUid: "uid1"}
	order2 := &app.Order{OrderUid: "uid2"}
	order3 := &app.Order{OrderUid: "uid3"}

	cache.Set(order1)
	cache.Set(order2)

	o, ok := cache.Get("uid1")
	assert.True(t, ok)
	assert.Equal(t, order1, o)
	o, ok = cache.Get("uid2")
	assert.True(t, ok)
	assert.Equal(t, order2, o)

	cache.Set(order3)
	_, ok = cache.Get("uid1")
	assert.False(t, ok)
	_, ok = cache.Get("uid2")
	assert.True(t, ok)
	_, ok = cache.Get("uid3")
	assert.True(t, ok)
}

func TestSetEmptyOrderUid(t *testing.T) {
	conf := &config.Config{CacheSize: 2}
	cache := NewOrderCache(conf)
	order := &app.Order{OrderUid: ""}
	cache.Set(order)
	assert.Equal(t, 0, cache.orderQueue.Len())
	assert.Equal(t, 0, len(cache.orders))
}

func TestSetDuplicateOrder(t *testing.T) {
	conf := &config.Config{CacheSize: 2}
	cache := NewOrderCache(conf)
	order := &app.Order{OrderUid: "uid1"}
	cache.Set(order)
	cache.Set(order) // Should not add duplicate
	assert.Equal(t, 1, cache.orderQueue.Len())
	assert.Equal(t, 1, len(cache.orders))
}

func TestUploadFromDb(t *testing.T) {
	conf := &config.Config{CacheSize: 2}
	orders := []*app.Order{
		{OrderUid: "uid1"},
		{OrderUid: "uid2"},
	}
	repo := &mockRepo{orders: orders}
	cache := NewOrderCache(conf)
	err := cache.UploadFromDb(repo, conf)
	assert.NoError(t, err)
	o, ok := cache.Get("uid1")
	assert.True(t, ok)
	assert.Equal(t, "uid1", o.OrderUid)
	o, ok = cache.Get("uid2")
	assert.True(t, ok)
	assert.Equal(t, "uid2", o.OrderUid)
}