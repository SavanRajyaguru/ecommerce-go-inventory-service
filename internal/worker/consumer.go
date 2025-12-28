package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/savanrajyaguru/ecommerce-go-inventory-service/config"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/services"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	service *services.InventoryService
	reader  *kafka.Reader
}

func NewKafkaConsumer(service *services.InventoryService) *KafkaConsumer {
	if len(config.AppConfig.Kafka.Brokers) == 0 {
		return nil
	}

	brokers := config.AppConfig.Kafka.Brokers
	groupID := config.AppConfig.Kafka.GroupID

	// Determine topic
	// Ideally we want to listen to a specific topic or list.
	// kafka-go reader is per topic usually, or can take a list (GroupTopics) in Config.
	// But commonly NewReader takes Topic.
	topic := "order.cancelled"
	if t, ok := config.AppConfig.Kafka.Topics["order_cancelled"]; ok {
		topic = t
	}

	log.Printf("Initializing Kafka consumer for topic: %s, brokers: %v", topic, brokers)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset, // Start from new messages only? Or FirstOffset? Usually Group tracks it.
		// If GroupID is set, StartOffset is only used when no commit exists.
	})

	return &KafkaConsumer{
		service: service,
		reader:  reader,
	}
}

func (k *KafkaConsumer) Start() {
	if k == nil || k.reader == nil {
		log.Println("Kafka consumer not initialized, skipping")
		return
	}

	log.Println("Kafka consumer started")
	ctx := context.Background()

	go func() {
		defer k.reader.Close()

		for {
			m, err := k.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error while reading message: %v", err)
				// Break loop if reader is closed? or retry?
				// ReadMessage blocks. Error usually means IO error or context cancel.
				// Simple backoff
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("Received message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

			var event struct {
				ProductID uint `json:"product_id"`
				Quantity  int  `json:"quantity"`
			}

			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}

			// Process Release Stock
			// We handle context cancellation in real app, here background is fine for now
			processCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = k.service.ReleaseStock(processCtx, event.ProductID, event.Quantity)
			cancel()

			if err != nil {
				log.Printf("Failed to release stock: %v", err)
				// In kafka-go, ReadMessage auto-commits (if CommitInterval set) or manual CommitMessages needed.
				// ReaderConfig default CommitInterval is 0 (sync). But we set it to 1s.
				// So it will confirm eventually.
				// If strictly exactly-once or at-least-once with manual retry is needed, we should use FetchMessage + CommitMessages.
				// For now simple AutoCommit logic via ReadMessage is matching previous ConsumerGroup behavior.
			}
		}
	}()
}
