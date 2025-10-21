package cache

import (
	"context"
	"time"
)

type Local[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	SetWithTTL(key K, value V, ttl time.Duration)
	Delete(key K)
}

type Invalidator[K comparable] interface {
	PublishEvict(ctx context.Context, keys []K) error
	SubscribeEvict(ctx context.Context, handler func(keys []K))
}
