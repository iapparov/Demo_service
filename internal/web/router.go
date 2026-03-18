package web

import (
	_ "demoservice/docs"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(r chi.Router, webHandler *WebHandler) {
	r.Get("/order/{orderUID}", webHandler.GetOrderByUID)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
}
