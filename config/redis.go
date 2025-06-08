package config

import (
	"sync"

	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RedisCacheClient struct {
	RedisClient *redis.Client
}

var (
	redisCacheClient *RedisCacheClient
	redisCacheOnce   sync.Once
)

func NewRedisCache(config *models.AppConfig) *RedisCacheClient {
	redisCacheOnce.Do(func() {
		redisCacheClient = &RedisCacheClient{
			RedisClient: redis.NewClient(&redis.Options{
				Addr:     config.Redis.Addr,
				Password: config.Redis.Pwd,
				DB:       config.Redis.DB,
			})}
	})
	return redisCacheClient
}

func GetRedisClient() *RedisCacheClient {
	return redisCacheClient
}
