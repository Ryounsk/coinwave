package database

import (
	"coin-wave/models"
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	Ctx = context.Background()
)

func InitDB() {
	var err error

	// MySQL Connection
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Fallback for local development if not using docker-compose env
		dsn = "root:rootpassword@tcp(localhost:3306)/coinwave?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Retry connection loop (for docker startup timing)
	for i := 0; i < 30; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying in 2s... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(&models.User{}, &models.Article{}, &models.Bookmark{}, &models.WalletLog{}, &models.Purchase{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("MySQL database connected and migrated successfully.")

	// Redis Connection
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	RDB = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := RDB.Ping(Ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Redis connected successfully.")
	}
}
