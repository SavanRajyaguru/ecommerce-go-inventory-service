package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(host, port, password string) {
	addr := fmt.Sprintf("%s:%s", host, port)
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis at %s: %v", addr, err)
		return
	}

	log.Println("Connected to Redis successfully")
}
