package di

import (
	"context"
	"demoservice/internal/web"
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"log"
	"demoservice/internal/config"
	"demoservice/internal/cache"
	"demoservice/internal/db"
	"demoservice/internal/kafka"
	"github.com/go-chi/cors"
)


func StartHTTPServer(lc fx.Lifecycle, webHandler *web.WebHandler, config *config.Config) {
	router := chi.NewRouter()
	
    corsMiddleware := cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"}, // или список разрешенных доменов
        AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        AllowCredentials: false,
        MaxAge:           300,
    })

    router.Use(corsMiddleware)

	web.RegisterRoutes(router, webHandler)

	addres := fmt.Sprintf(":%d", config.HttpPort)
	server := &http.Server{
		Addr:    addres,
		Handler: router, 
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server started")
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Printf("ListenAndServe error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Printf("Shutting down server...")
			return server.Close()
		},
	})
}

func LoadCacheOnStart(lc fx.Lifecycle, c *cache.OrderCache, repo *db.PostgresRepo, conf *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Loading cache from DB on startup...")
			if err := c.UploadFromDb(repo, conf); err != nil {
				log.Printf("Failed to load cache: %v", err)
				return err
			}

			log.Println("Cache loaded successfully")
			return nil
		},
	})
}

func StartKafkaConsumer(lc fx.Lifecycle, cache *cache.OrderCache, repo *db.PostgresRepo) {
	consumer := kafka.NewConsumer([]string{"localhost:9092"}, "orders", "order-consumer", cache, repo)

	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			log.Println("Starting Kafka consumer...")
			go consumer.Run(ctx)
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			log.Println("Stopping Kafka consumer...")
			cancel()
			return consumer.Close()
		},
	})
}