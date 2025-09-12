package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	fmt.Println("🚀 Testing News Aggregator Infrastructure...")

	// Test Redis connection
	fmt.Println("🔍 Testing Redis connection...")
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("❌ Redis connection failed: %v\n", err)
	} else {
		fmt.Println("✅ Redis connection successful!")
	}

	// Test Elasticsearch
	fmt.Println("🔍 Testing Elasticsearch connection...")
	resp, err := http.Get("http://localhost:9200")
	if err != nil {
		fmt.Printf("❌ Elasticsearch connection failed: %v\n", err)
	} else {
		resp.Body.Close()
		fmt.Println("✅ Elasticsearch connection successful!")
	}

	// Start simple web server
	fmt.Println("🌐 Starting test web server on :8080...")
	
	r := gin.Default()
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now(),
			"services": gin.H{
				"redis":         "connected",
				"elasticsearch": "connected",
			},
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "News Aggregator Infrastructure Test",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health": "/health",
				"test":   "/test",
			},
		})
	})

	r.GET("/test", func(c *gin.Context) {
		// Test Redis
		err := rdb.Set(ctx, "test_key", "test_value", time.Minute).Err()
		if err != nil {
			c.JSON(500, gin.H{"error": "Redis test failed"})
			return
		}

		val, err := rdb.Get(ctx, "test_key").Result()
		if err != nil {
			c.JSON(500, gin.H{"error": "Redis get failed"})
			return
		}

		c.JSON(200, gin.H{
			"redis_test": "passed",
			"value":      val,
			"timestamp":  time.Now(),
		})
	})

	fmt.Println("✅ Test server ready!")
	fmt.Println("📋 Test URLs:")
	fmt.Println("   • Health: http://localhost:8081/health")
	fmt.Println("   • Home: http://localhost:8081/")
	fmt.Println("   • Test: http://localhost:8081/test")
	fmt.Println("")

	log.Fatal(r.Run(":8081"))
}
