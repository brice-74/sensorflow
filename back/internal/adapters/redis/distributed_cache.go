package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type DistributedCache[K comparable, V any] struct {
	client    *redis.Client
	encode    func(V) ([]byte, error)
	decode    func([]byte) (V, error)
	formatKey func(K) string
	ttl       time.Duration
}

func NewDistributedCache[K comparable, V any](client *redis.Client, encode func(V) ([]byte, error), decode func([]byte) (V, error)) *DistributedCache[K, V] {
	return &DistributedCache[K, V]{client: client, encode: encode, decode: decode}
}

func (r *DistributedCache[K, V]) Get(ctx context.Context, key K) (V, error) {
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

func (r *DistributedCache[K, V]) MGet(ctx context.Context, keys ...K) (map[K]V, error) {
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = r.formatKey(k)
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

func (r *DistributedCache[K, V]) Set(ctx context.Context, key K, value V) error {
	data, err := r.encode(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.formatKey(key), data, r.ttl).Err()
}

func (r *DistributedCache[K, V]) SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error {
	data, err := r.encode(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.formatKey(key), data, ttl).Err()
}

func (r *DistributedCache[K, V]) Delete(ctx context.Context, key K) error {
	return r.client.Del(ctx, r.formatKey(key)).Err()
}
