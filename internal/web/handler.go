package web

import (
	"bytes"
	"demoservice/internal/cache"
	"demoservice/internal/db"
	//"encoding/json"
	json "github.com/goccy/go-json" // ускоряем работу с json
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
)

type WebHandler struct {
	cache   *cache.OrderCache
	repo    db.Repository
	bufPool *sync.Pool // Переиспользование буферов снижает аллокации и нагрузку на GC.
}

func NewWebHanlder(cache *cache.OrderCache, repo db.Repository) *WebHandler {
	return &WebHandler{
		cache: cache,
		repo:  repo,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 2048))
			},
		},
	}
}

// GetOrderByUID godoc
// @Summary      Получить заказ по UID
// @Description  Возвращает заказ из кэша или базы данных
// @Tags         orders
// @Param        orderUID path string true "Order UID"
// @Produce      json
// @Success      200 {object} app.Order
// @Failure      404 {object} map[string]string
// @Router       /order/{orderUID} [get]
func (h *WebHandler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	orderUID := chi.URLParam(r, "orderUID")
	if orderUID == "" {
		http.Error(w, "Missing order uid", http.StatusBadRequest)
		return
	}

	// Кэш возвращает уже сериализованый джсон
	if jsonData, ok := h.cache.GetJSON(orderUID); ok {
		log.Printf("Loaded from cache for %s\n", orderUID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
		return
	}

	order, err := h.repo.Load(orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		log.Printf("Loaded from database for %s\n", orderUID)
		return
	}

	h.cache.Set(order)
	h.writeJSON(w, order, http.StatusOK)
}

func (h *WebHandler) writeJSON(w http.ResponseWriter, data any, status int) {
	buf := h.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer h.bufPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		log.Printf("JSON encode error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}
