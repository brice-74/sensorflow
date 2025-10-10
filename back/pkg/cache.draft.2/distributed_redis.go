package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache[K comparable, V any] struct {
	client    *redis.Client
	encode    func(V) ([]byte, error)
	decode    func([]byte) (V, error)
	formatKey func(K) string
	ttl       time.Duration
}

func NewRedisCache[K comparable, V any](client *redis.Client, encode func(V) ([]byte, error), decode func([]byte) (V, error)) *RedisCache[K, V] {
	return &RedisCache[K, V]{client: client, encode: encode, decode: decode}
}

func (r *RedisCache[K, V]) Get(ctx context.Context, key K) (V, error) {
	data, err := r.client.Get(ctx, r.formatKey(key)).Bytes()
	if err == redis.Nil {
		var zero V
		return zero, nil
	}
	if err != nil {
		var zero V
		return zero, err
	}
	return r.decode(data)
}

func (r *RedisCache[K, V]) MGet(ctx context.Context, keys ...K) (map[K]V, error) {
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = any(k).(string)
	}

	results, err := r.client.MGet(ctx, strKeys...).Result()
	if err != nil {
		return nil, err
	}

	out := make(map[K]V, len(keys))
	for i, val := range results {
		if val == nil {
			continue
		}
		v, err := r.decode([]byte(val.(string)))
		if err == nil {
			out[keys[i]] = v
		}
	}
	return out, nil
}

func (r *RedisCache[K, V]) Set(ctx context.Context, key K, value V) error {
	data, err := r.encode(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, any(key).(string), data, r.ttl).Err()
}

func (r *RedisCache[K, V]) SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error {
	data, err := r.encode(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, any(key).(string), data, ttl).Err()
}

func (r *RedisCache[K, V]) Delete(ctx context.Context, key K) error {
	return r.client.Del(ctx, any(key).(string)).Err()
}
