package utils

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	ctx context.Context
	rdb *redis.Client
}

func CreateRedisClient(ctx context.Context, addr string, pass string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	return &RedisClient{
		ctx: ctx,
		rdb: rdb,
	}
}

func (r *RedisClient) Set(key string, value interface{}) error  {
	err := r.rdb.Set(r.ctx, key, value, 0).Err()

	return err
}

func (r *RedisClient) Get(key string) (string, error)  {
	return r.rdb.Get(r.ctx, key).Result()
}
