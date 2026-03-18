package web

import (
	"demoservice/internal/app"
	"demoservice/internal/cache"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ========================================================================================
// Бенчмарк 1: Mock (без реальной БД)
// ========================================================================================
// Зачем: измеряет чистую производительность Go-кода — роутинг chi, кэш, JSON-сериализацию.
// Результаты стабильные, не зависят от состояния БД/сети/диска.
// Именно этот бенчмарк используется для оптимизации CPU и аллокаций памяти через pprof.
// Запуск: go test -bench=BenchmarkGetOrder_CacheHit -benchmem ./internal/web/
// ========================================================================================

// setupBenchRouter создаёт chi-роутер с mock-зависимостями.
func setupBenchRouter() http.Handler {
	conf := testConfig() // подтянули из handler_test.go

	c := cache.NewOrderCache(conf)
	order := &app.Order{OrderUid: "b563feb7b2b84b6test"}
	c.Set(order)

	repo := &mockRepo{order: order} // аналогично подтянуто из hanler_test.go
	handler := NewWebHanlder(c, repo)

	r := chi.NewRouter()
	r.Get("/order/{orderUID}", handler.GetOrderByUID)

	return r
}

// BenchmarkGetOrder_CacheHit — запрос попадает в кэш, БД не трогается.
// Это основной бенчмарк для оптимизации: показывает аллокации и CPU в чистом Go-коде.
func BenchmarkGetOrder_CacheHit(b *testing.B) {
	router := setupBenchRouter()
	req := httptest.NewRequest(http.MethodGet, "/order/b563feb7b2b84b6test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetOrder_CacheMiss — запрос НЕ в кэше, идёт в mock-репозиторий.
// Показывает разницу между cache hit и cache miss (без реальной БД).
func BenchmarkGetOrder_CacheMiss(b *testing.B) {
	conf := testConfig()
	c := cache.NewOrderCache(conf)
	// НЕ кладём заказ в кэш — будет промах, handler пойдёт в repo.
	order := &app.Order{OrderUid: "miss-uid"}
	repo := &mockRepo{order: order}
	handler := NewWebHanlder(c, repo)

	r := chi.NewRouter()
	r.Get("/order/{orderUID}", handler.GetOrderByUID)

	req := httptest.NewRequest(http.MethodGet, "/order/miss-uid", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
