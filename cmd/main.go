// @title           Demo Service API
// @version         1.0
// @description     API для просмотра заказов
// @BasePath        /

package main

import (
	"go.uber.org/fx"
	"demoservice/internal/config"
	"demoservice/internal/web"
	"demoservice/internal/db"
	"demoservice/internal/cache"
	"demoservice/internal/di"
)

func main(){
	app := fx.New(
		
		fx.Provide(
			config.MustLoad, 
			db.ConnectDB,
			db.NewPostgresRepo,
			func (db *db.PostgresRepo) db.Repository{
				return db
			},
			cache.NewOrderCache,
			web.NewWebHanlder,
		),

		fx.Invoke(
			di.StartHTTPServer,
			di.StartKafkaConsumer,
			di.LoadCacheOnStart,
		),
	)

	app.Run()
}