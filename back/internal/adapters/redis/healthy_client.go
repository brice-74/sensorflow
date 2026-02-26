package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/heartbeat"
	"github.com/redis/go-redis/v9"
)

type ClientAliveCapable redis.Client

func (c *ClientAliveCapable) Alive(ctx context.Context) bool {
	return c.Ping(ctx).Err() == nil
}

type HealthyClient struct {
	*heartbeat.Watcher[*ClientAliveCapable]
	*redis.Client
}

var _ Client = (*HealthyClient)(nil)

func NewHealthyClientFromCfg(cfg *config.Redis) *HealthyClient {
	client := redis.NewClient(&redis.Options{
		Network:               cfg.Network,
		Addr:                  cfg.Addr,
		ClientName:            cfg.ClientName,
		Username:              cfg.Username,
		Password:              cfg.Password,
		DB:                    cfg.DB,
		MaxRetries:            cfg.MaxRetries,
		MinRetryBackoff:       cfg.MinRetryBackoff,
		MaxRetryBackoff:       cfg.MaxRetryBackoff,
		DialTimeout:           cfg.DialTimeout,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		ContextTimeoutEnabled: cfg.ContextTimeoutEnabled,
		ReadBufferSize:        cfg.ReadBufferSize,
		WriteBufferSize:       cfg.WriteBufferSize,
		PoolFIFO:              cfg.PoolFIFO,
		PoolSize:              cfg.PoolSize,
		PoolTimeout:           cfg.PoolTimeout,
		MinIdleConns:          cfg.MinIdleConns,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxActiveConns:        cfg.MaxActiveConns,
		ConnMaxIdleTime:       cfg.ConnMaxIdleTime,
		ConnMaxLifetime:       cfg.ConnMaxLifetime,
	})

	aliveCapable := ClientAliveCapable(*client)

	watcher := heartbeat.NewWatcher(&aliveCapable, HandleError, &heartbeat.Options{
		RecoverTimeout: cfg.HealthRecoverTimeout,
		PingTimeout:    cfg.HealthPingTimeout,
		InitialBackoff: cfg.HealthInitialBackoff,
		MaxBackoff:     cfg.HealthMaxBackoff,
		Multiplier:     cfg.HealthMultiplier,
		JitterPct:      cfg.HealthJitterPct,
	}, false)

	healthyClient := HealthyClient{
		Watcher: watcher,
		Client:  client,
	}
	return &healthyClient
}

func (h *HealthyClient) Cmdable(ctx context.Context) redis.Cmdable {
	return h.Client
}

func (h *HealthyClient) NewPipeline(ctx context.Context) (redis.Pipeliner, context.Context) {
	pipe := h.Client.Pipeline()
	return pipe, WithPipeline(ctx, pipe)
}

func (h *HealthyClient) NewConn(ctx context.Context) (redis.Cmdable, context.Context) {
	conn := h.Client.Conn()
	return conn, WithConn(ctx, conn)
}
