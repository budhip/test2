package cache

import (
	"context"
	"encoding/json"
	"time"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     20,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return &RedisCache{client: client}
}

func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, jsonData, ttl).Err()
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}

func (rc *RedisCache) GetOrSet(
	ctx context.Context,
	key string,
	ttl time.Duration,
	fetchFunc func() (interface{}, error),
) (interface{}, error) {

	cached, err := rc.client.Get(ctx, key).Bytes()
	if err == nil {
		var result interface{}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	data, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(data)
	if err == nil {
		go rc.client.Set(ctx, key, jsonData, ttl).Err()
	}

	return data, nil
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
