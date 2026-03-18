package web

import (
	"demoservice/internal/cache"
	"demoservice/internal/db"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type WebHandler struct {
	cache *cache.OrderCache
	repo  db.Repository
}

func NewWebHanlder(cache *cache.OrderCache, repo db.Repository) *WebHandler {
	return &WebHandler{
		cache: cache,
		repo:  repo,
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

	if order, ok := h.cache.Get(orderUID); ok {
		//log.Printf("Loaded from cache for %s\n", orderUID)
		h.writeJSON(w, order, http.StatusOK)
		return
	}

	order, err := h.repo.Load(orderUID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	//log.Printf("Loaded from database for %s\n", orderUID)
	h.cache.Set(order)
	h.writeJSON(w, order, http.StatusOK)

}

func (h *WebHandler) writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}
