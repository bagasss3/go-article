package database

import (
	"context"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

func NewRedisConn(url string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: config.RedisPassword(),
		DB:       config.RedisDB(),
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.WithField("Addr", config.RedisHost()).Fatal("Failed to connect:", err)
	}

	log.Info("Success connect to redis")

	return rdb
}
