package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"os"
	"log"
)

func main() {
	jsonData, _ := os.ReadFile("model.json")
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "orders",
	})
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("Writer close error: %v", err)
		}
	}()
	var msg map[string]any
	if err := json.Unmarshal(jsonData, &msg); err != nil {
    	log.Printf("Unmarshal error: %v", err)
	}

	err := w.WriteMessages(context.Background(), kafka.Message{
		Value: jsonData,
	})
	if err != nil {
		panic(err)
	}
}