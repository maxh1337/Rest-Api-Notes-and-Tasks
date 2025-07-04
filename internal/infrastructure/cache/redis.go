package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"rest-api-notes/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	Client *redis.Client
}

type RedisClient interface {
	SetStruct(ctx context.Context, key string, value any, expiration time.Duration) error
	GetStruct(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	GetAllByKey(ctx context.Context, pattern string, dest any) error
}

func NewRedisClient(cfg *config.RedisConfig) (RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	log.Println("Redis connected successfully")

	return &redisClient{Client: rdb}, nil
}

func (c *redisClient) SetStruct(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Client.Set(ctx, key, data, expiration).Err()
}

func (r *redisClient) GetStruct(ctx context.Context, key string, dest any) error {
	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (r *redisClient) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *redisClient) GetAllByKey(ctx context.Context, pattern string, dest any) error {
	iter := r.Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		data, err := r.Client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		slicePtr := dest
		elemType := reflect.TypeOf(slicePtr).Elem().Elem()
		elemPtr := reflect.New(elemType)
		if err := json.Unmarshal([]byte(data), elemPtr.Interface()); err != nil {
			continue
		}
		reflect.ValueOf(slicePtr).Elem().Set(
			reflect.Append(reflect.ValueOf(slicePtr).Elem(), elemPtr.Elem()),
		)
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
