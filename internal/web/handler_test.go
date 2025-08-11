package web

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "demoservice/internal/app"
    "demoservice/internal/cache"
    "demoservice/internal/config"
    "github.com/go-chi/chi/v5"
    "github.com/stretchr/testify/assert"
    "context"
)

func testConfig() *config.Config {
    return &config.Config{
        CacheSize:  2,
    }
} 

type mockRepo struct {
    order *app.Order
    err   error
}

func (m *mockRepo) Load(uid string) (*app.Order, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.order, nil
}
func (m *mockRepo) Save(order *app.Order) error { return nil }
func (m *mockRepo) CacheLoad(conf *config.Config) ([]*app.Order, error) { return nil, nil }

func TestGetOrderByUID_FromCache(t *testing.T) {
    config := testConfig()
    cache := cache.NewOrderCache(config)
    order := &app.Order{OrderUid: "uid1"}
    cache.Set(order)
    handler := NewWebHanlder(cache, &mockRepo{order: order})

    r := httptest.NewRequest("GET", "/order/uid1", nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("orderUID", "uid1")
    ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
    r = r.WithContext(ctx)
    w := httptest.NewRecorder()

    handler.GetOrderByUID(w, r)
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "uid1")
}

func TestGetOrderByUID_FromDB(t *testing.T) {
    config := testConfig()
    cache := cache.NewOrderCache(config)
    order := &app.Order{OrderUid: "uid2"}
    handler := NewWebHanlder(cache, &mockRepo{order: order})

    r := httptest.NewRequest("GET", "/order/uid2", nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("orderUID", "uid2")
    ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
    r = r.WithContext(ctx)
    w := httptest.NewRecorder()

    handler.GetOrderByUID(w, r)
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "uid2")
}

func TestGetOrderByUID_NotFound(t *testing.T) {
    config := testConfig()
    cache := cache.NewOrderCache(config)
    handler := NewWebHanlder(cache, &mockRepo{order: nil, err: assert.AnError})

    r := httptest.NewRequest("GET", "/order/uid3", nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("orderUID", "uid3")
    ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
    r = r.WithContext(ctx)
    w := httptest.NewRecorder()

    handler.GetOrderByUID(w, r)
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetOrderByUID_EmptyUID(t *testing.T) {
    config := testConfig()
    cache := cache.NewOrderCache(config)
    handler := NewWebHanlder(cache, &mockRepo{order: nil})

    r := httptest.NewRequest("GET", "/order/", nil)
    rctx := chi.NewRouteContext()
    ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
    r = r.WithContext(ctx)
    w := httptest.NewRecorder()

    handler.GetOrderByUID(w, r)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}