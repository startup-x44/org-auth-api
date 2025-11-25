package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"auth-service/internal/config"
)

func main() {
	log.Println("Testing Upstash Redis connection...")

	// Load configuration
	cfg := config.Load()

	// Create Redis client
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Enable TLS for Upstash
	if cfg.Redis.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		log.Println("✅ TLS enabled for Redis connection")
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	// Test connection
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}

	log.Printf("✅ Successfully connected to Upstash Redis!")
	log.Printf("✅ PING response: %s", pong)

	// Test SET/GET
	err = rdb.Set(ctx, "test:key", "Hello Upstash!", 0).Err()
	if err != nil {
		log.Fatalf("❌ Failed to SET: %v", err)
	}
	log.Println("✅ SET test:key = 'Hello Upstash!'")

	val, err := rdb.Get(ctx, "test:key").Result()
	if err != nil {
		log.Fatalf("❌ Failed to GET: %v", err)
	}
	log.Printf("✅ GET test:key = '%s'", val)

	log.Println("")
	log.Println("🎉 Upstash Redis is working perfectly!")
}
