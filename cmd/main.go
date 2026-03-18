// @title           Demo Service API
// @version         1.0
// @description     API для просмотра заказов
// @BasePath        /

package main

import (
	"demoservice/internal/cache"
	"demoservice/internal/config"
	"demoservice/internal/db"
	"demoservice/internal/di"
	"demoservice/internal/web"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(

		fx.Provide(
			config.MustLoad,
			db.ConnectDB,
			db.NewPostgresRepo,
			func(db *db.PostgresRepo) db.Repository {
				return db
			},
			cache.NewOrderCache,
			web.NewWebHanlder,
		),

		fx.Invoke(
			di.StartHTTPServer,
			di.StartKafkaConsumer,
			di.LoadCacheOnStart,
			di.StartPprofServer,
		),
	)

	app.Run()
}
