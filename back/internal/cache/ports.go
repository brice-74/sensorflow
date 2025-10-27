package cache

import (
	"context"
	"time"
)

type Local[V any] interface {
	Get(key string) (V, bool)
	Set(key string, value V)
	SetWithTTL(key string, value V, ttl time.Duration)
	Del(key string)
}

type InvalidatorPublisher[K comparable] interface {
	PublishEvict(ctx context.Context, keys []K) error
}

type InvalidatorSubscriber[K comparable] interface {
	SubscribeEvict(ctx context.Context, handler func(keys []K))
}
