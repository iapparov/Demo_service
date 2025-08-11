package web

import(
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
    _ "demoservice/docs"
)

func RegisterRoutes(r chi.Router, webHandler *WebHandler){
	r.Get("/order/{orderUID}", webHandler.GetOrderByUID)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
}