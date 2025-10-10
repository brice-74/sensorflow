package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Distributed[K comparable, V any] interface {
	Get(ctx context.Context, key K) (V, bool, error)
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	Del(ctx context.Context, key K) error
	Close(ctx context.Context) error
}

type Redis[K comparable, V any] struct {
	client  *redis.Client
	prefix  string
	keyFunc func(K) string
}

func NewRedis[K comparable, V any](client *redis.Client, prefix string, keyFunc func(K) string) *Redis[K, V] {
	return &Redis[K, V]{client: client, prefix: prefix, keyFunc: keyFunc}
}

func (r *Redis[K, V]) makeKey(k K) string {
	return fmt.Sprintf("%s:%v", r.prefix, r.keyFunc(k))
}

func (r *Redis[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	raw, err := r.client.Get(ctx, r.makeKey(key)).Result()
	if err == redis.Nil {
		var zero V
		return zero, false, nil
	}
	if err != nil {
		var zero V
		return zero, false, err
	}
	var val V
	if err := json.Unmarshal([]byte(raw), &val); err != nil {
		var zero V
		return zero, false, err
	}
	return val, true, nil
}

func (r *Redis[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.makeKey(key), data, ttl).Err()
}

func (r *Redis[K, V]) Del(ctx context.Context, key K) error {
	return r.client.Del(ctx, r.makeKey(key)).Err()
}

func (r *Redis[K, V]) Close(ctx context.Context) error {
	return r.client.Close()
}
