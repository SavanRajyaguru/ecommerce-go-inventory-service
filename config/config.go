package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/SavanRajyaguru/ecommerce-go-config-service/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	AppPort      string
	JWTSecret    string
	DB           DBConfig
	Redis        RedisConfig
	Kafka        KafkaConfig
	FeatureFlags map[string]bool
	GrpcPort     string
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

type KafkaConfig struct {
	Brokers []string          `json:"brokers"`
	Topics  map[string]string `json:"topics"`
	GroupID string            `json:"group_id"`
}

var AppConfig *Config

func LoadConfig() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	AppConfig = &Config{
		AppPort:   getEnv("PORT", "8083"),
		JWTSecret: getEnv("JWT_SECRET", ""),
		GrpcPort:  getEnv("GRPC_PORT", "50052"),
		// Default internal Kafka config if not fetched remotely, but user said ALL service configs from ConfigService
		// However, sensitive creds must be local. Kafka brokers might be sensitive?
		// "Kafka broker credentials" -> .env
		// "Kafka topic names" -> Config Service
	}
	fmt.Printf("DEBUG: Loaded Config - Port: %s, GrpcPort: %s\n", AppConfig.AppPort, AppConfig.GrpcPort)
	if AppConfig.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required in environment")
	}

	// 2. Fetch from Config Service
	configServiceURL := getEnv("CONFIG_SERVICE_URL", "localhost:50051")
	log.Printf("Connecting to Config Service at: %s", configServiceURL)
	fetchRemoteConfig(configServiceURL)
}

func fetchRemoteConfig(url string) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	client := pb.NewConfigServiceClient(conn)

	// Retry logic for the RPC call
	var resp *pb.GetConfigResponse

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err = client.GetConfig(ctx, &pb.GetConfigRequest{
			ServiceName: "inventory-service",
		})
		cancel()

		if err == nil {
			break
		}

		log.Printf("Failed to fetch config (attempt %d/10): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to fetch config from %s after retries: %v", url, err)
	}

	var remoteCfg struct {
		DB    DBConfig    `json:"db"`
		Redis RedisConfig `json:"redis"`
		Kafka struct {
			Brokers []string          `json:"brokers"` // Brokers might come from env if sensitive, but topics from here
			Topics  map[string]string `json:"topics"`
		} `json:"kafka"`
		ServerPort   string          `json:"server_port"`
		FeatureFlags map[string]bool `json:"feature_flags"`
	}

	if err := json.Unmarshal([]byte(resp.ConfigJson), &remoteCfg); err != nil {
		log.Fatalf("Failed to unmarshal config json: %v", err)
	}

	// 3. Merge Strategies
	// DB & Redis: Strictly from Remote Config Service as per user requirement.
	AppConfig.DB = remoteCfg.DB
	AppConfig.Redis = remoteCfg.Redis

	AppConfig.Kafka.Topics = remoteCfg.Kafka.Topics
	// Kafka brokers -> Env
	brokersEnv := getEnv("KAFKA_BROKERS", "")
	if brokersEnv != "" {
		AppConfig.Kafka.Brokers = []string{brokersEnv} // Simplification
	} else {
		AppConfig.Kafka.Brokers = remoteCfg.Kafka.Brokers
	}
	AppConfig.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", "inventory-service-group")
	AppConfig.FeatureFlags = remoteCfg.FeatureFlags

	fmt.Println("Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
