package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Basic Invalidator interface (wrapping Redis pub/sub)
type RedisInvalidator[K comparable] struct {
	client  *redis.Client
	channel string
}

func NewRedisInvalidator[K comparable](client *redis.Client, channel string) *RedisInvalidator[K] {
	return &RedisInvalidator[K]{client: client, channel: channel}
}

func (i *RedisInvalidator[K]) PublishInvalidate(ctx context.Context, keys []K) error {
	data, err := encodeGob(keys)
	if err != nil {
		return err
	}
	return i.client.Publish(ctx, i.channel, data).Err()
}

func (i *RedisInvalidator[K]) SubscribeInvalidate(ctx context.Context, handler func(keys []K)) {
	sub := i.client.Subscribe(ctx, i.channel)
	ch := sub.Channel()

	go func() {
		for msg := range ch {
			var keys []K
			if err := decodeGob([]byte(msg.Payload), &keys); err == nil {
				handler(keys)
			}
		}
	}()
}

// ShardedRedisInvalidator wraps an Invalidator and only handles messages for its shard.
type ShardedRedisInvalidator[K comparable] struct {
	shardID   int
	numShards int
	inner     Invalidator[K]
	hashFn    func(K) uint32
}

func (s *ShardedRedisInvalidator[K]) PublishInvalidate(ctx context.Context, keys []K) error {
	var localKeys []K
	for _, key := range keys {
		if int(s.hashFn(key))%s.numShards == s.shardID {
			localKeys = append(localKeys, key)
		}
	}
	if len(localKeys) == 0 {
		return nil
	}
	return s.inner.PublishInvalidate(ctx, localKeys)
}

func (s *ShardedRedisInvalidator[K]) SubscribeInvalidate(ctx context.Context, handler func(keys []K)) {
	s.inner.SubscribeInvalidate(ctx, func(keys []K) {
		var localKeys []K
		for _, key := range keys {
			if int(s.hashFn(key))%s.numShards == s.shardID {
				localKeys = append(localKeys, key)
			}
		}
		if len(localKeys) > 0 {
			handler(localKeys)
		}
	})
}
