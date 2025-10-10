package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// InvalidatorHandler est appelée lorsqu’une invalidation est reçue.
type InvalidatorHandler[K comparable] func(ctx context.Context, key K) error

type Invalidator[K comparable] interface {
	PublishInvalidate(ctx context.Context, key K) error
	SubscribeInvalidations(handler InvalidatorHandler[K]) error
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}

// RedisInvalidator implémente un système d’invalidation basé sur Redis Pub/Sub.
// Une seule goroutine écoute Redis et notifie tous les handlers enregistrés.
type RedisInvalidator[K comparable] struct {
	client  *redis.Client
	channel string

	mu       sync.RWMutex
	handlers []InvalidatorHandler[K]

	startOnce sync.Once
	cancel    context.CancelFunc
}

// NewRedisInvalidator crée un nouvel invalidator Redis.
func NewRedisInvalidator[K comparable](client *redis.Client, channel string) *RedisInvalidator[K] {
	return &RedisInvalidator[K]{client: client, channel: channel}
}

func (r *RedisInvalidator[K]) PublishInvalidate(ctx context.Context, key K) error {
	data, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}
	return r.client.Publish(ctx, r.channel, data).Err()
}

func (r *RedisInvalidator[K]) SubscribeInvalidations(handler InvalidatorHandler[K]) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, handler)
	return nil
}

// Start lance UNE SEULE goroutine pour écouter Redis et propager aux handlers.
func (r *RedisInvalidator[K]) Start(ctx context.Context) error {
	var err error
	r.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		r.cancel = cancel

		go func() {
			pubsub := r.client.Subscribe(ctx, r.channel)
			ch := pubsub.Channel()
			for {
				select {
				case <-ctx.Done():
					_ = pubsub.Close()
					return
				case msg, ok := <-ch:
					if !ok {
						return
					}

					var key K
					if err := json.Unmarshal([]byte(msg.Payload), &key); err != nil {
						continue
					}

					r.mu.RLock()
					for _, h := range r.handlers {
						// on ne bloque pas les autres handlers
						go h(ctx, key)
					}
					r.mu.RUnlock()
				}
			}
		}()
	})
	return err
}

func (r *RedisInvalidator[K]) Close(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.client.Close()
}
