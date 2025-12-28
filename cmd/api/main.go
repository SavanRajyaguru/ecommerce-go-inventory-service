package main

import (
	"fmt"
	"log"
	"net"

	"github.com/savanrajyaguru/ecommerce-go-inventory-service/api"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/config"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/cache"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/database"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/grpc"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/internal/worker"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/migrations"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/pkg/logger"
	pb "github.com/savanrajyaguru/ecommerce-go-inventory-service/proto"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/repository"
	"github.com/savanrajyaguru/ecommerce-go-inventory-service/services"
	googleGrpc "google.golang.org/grpc"
)

func main() {
	// 1. Initialize Logger
	logger.InitLogger()
	defer logger.Log.Sync()

	// 2. Load Configuration
	config.LoadConfig()

	// 3. Initialize Database
	database.ConnectDB(database.DBConfig{
		Host:     config.AppConfig.DB.Host,
		User:     config.AppConfig.DB.User,
		Password: config.AppConfig.DB.Password,
		DBName:   config.AppConfig.DB.Name,
		Port:     config.AppConfig.DB.Port,
		SSLMode:  config.AppConfig.DB.SSLMode,
	})

	// 4. Run Migrations
	migrations.RunMigrations()

	// 5. Initialize Redis
	if config.AppConfig.Redis.Enabled && config.AppConfig.Redis.Host != "" {
		cache.InitRedis(config.AppConfig.Redis.Host, config.AppConfig.Redis.Port, config.AppConfig.Redis.Password)
	}

	// Dependencies
	invRepo := repository.NewInventoryRepository()
	invService := services.NewInventoryService(invRepo)

	// 6. Start Kafka Consumer
	consumer := worker.NewKafkaConsumer(invService)
	if consumer != nil {
		consumer.Start()
	}
	// 7. Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.AppConfig.GrpcPort))
		if err != nil {
			log.Printf("Failed to listen for gRPC: %v", err)
			return
		}

		s := googleGrpc.NewServer()
		pb.RegisterInventoryServiceServer(s, grpc.NewInventoryGrpcServer(invService))

		log.Printf("Starting gRPC server on %s", config.AppConfig.GrpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 8. Setup & Start HTTP Server
	r := api.SetupRouter()
	addr := fmt.Sprintf(":%s", config.AppConfig.AppPort)

	log.Printf("Starting Inventory Service HTTP on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
