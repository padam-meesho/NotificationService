package dao

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisDaoImpl struct {
	client *redis.Client
}

func NewRedisDaoImpl(client *redis.Client) *RedisDaoImpl {
	return &RedisDaoImpl{
		client: client,
	}
}

func (r *RedisDaoImpl) GetAllBlacklistedNumbers(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, "blacklisted_numbers").Result()
}

func (r *RedisDaoImpl) CheckNumberInBlacklistedSet(ctx context.Context, number string) (bool, error) {
	return r.client.SIsMember(ctx, "blacklisted_numbers", number).Result()
}

func (r *RedisDaoImpl) AddNumberToBlacklistedSet(ctx context.Context, numbers []string) error {
	return r.client.SAdd(ctx, "blacklisted_numbers", numbers).Err()
}

func (r *RedisDaoImpl) RemoveFromBlacklistedSet(ctx context.Context, number string) (int64, error) {
	return r.client.SRem(ctx, "blacklisted_numbers", number).Result()
}
