package kafka

import (
	"context"
	"encoding/json"
	"demoservice/internal/cache"
	"demoservice/internal/db"
	"demoservice/internal/app"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type Consumer struct {
	reader *kafka.Reader
	cache  *cache.OrderCache
	repo   db.Repository
}

func NewConsumer(brokers []string, topic, groupID string, cache *cache.OrderCache, repo db.Repository) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
		CommitInterval: 0,
	})
	return &Consumer{
		reader: r,
		cache:  cache,
		repo:   repo,
	}
}

func (c *Consumer) Run(ctx context.Context) {
    for {
        m, err := c.reader.ReadMessage(ctx)
        if err != nil {
            if ctx.Err() != nil {
                return
            }
            log.Println("Kafka read error:", err)
            continue
        }

        var order app.Order
        if err := json.Unmarshal(m.Value, &order); err != nil {
            log.Println("Invalid message:", err)
            if err := c.reader.CommitMessages(ctx, m); err != nil {
                log.Println("Commit error after invalid JSON:", err)
            }
            continue
        }

        maxRetries := 5
        backoff := time.Second

        for attempt := 1; attempt <= maxRetries; attempt++ {
            err := c.repo.Save(&order)
            if err == nil {
                c.cache.Set(&order)
                if err = c.reader.CommitMessages(ctx, m); err != nil {
                    log.Println("Commit error:", err)
                } else {
                    log.Println("Committed message:", order.OrderUid)
                }
                break
            }

            log.Printf("DB error on attempt %d/%d: %v", attempt, maxRetries, err)

            if attempt < maxRetries {
                log.Printf("Retrying in %v...", backoff)
                time.Sleep(backoff)
                backoff *= 2
            } else {
                log.Println("Max retries reached")
            }
        }
    }
}
func (c *Consumer) Close() error {
	return c.reader.Close()
}