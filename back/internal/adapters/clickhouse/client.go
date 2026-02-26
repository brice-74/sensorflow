package clickhouse

import (
	"context"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/heartbeat"
	"github.com/google/uuid"
)

type ClientAliveCapable struct{ clickhouse.Conn }

func (c *ClientAliveCapable) Alive(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

type HealthyClient = heartbeat.Watcher[*ClientAliveCapable]

func NewHealthyClient(cfg *config.Clickhouse) *HealthyClient {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr:                 cfg.Addrs,
		ReadTimeout:          cfg.ReadTimeout,
		DialTimeout:          cfg.DialTimeout,
		MaxOpenConns:         cfg.MaxOpenConns,
		MaxIdleConns:         cfg.MaxIdleConns,
		ConnMaxLifetime:      cfg.ConnMaxLifetime,
		HttpMaxConnsPerHost:  cfg.HttpMaxConnsPerHost,
		BlockBufferSize:      uint8(cfg.BlockBufferSize),
		MaxCompressionBuffer: cfg.MaxCompressionBuffer,
	})
	if err != nil {
		return nil
	}

	aliveCapable := &ClientAliveCapable{
		Conn: conn,
	}

	watcher := heartbeat.NewWatcher(aliveCapable, HandleError, &heartbeat.Options{
		RecoverTimeout: cfg.HealthRecoverTimeout,
		PingTimeout:    cfg.HealthPingTimeout,
		InitialBackoff: cfg.HealthInitialBackoff,
		MaxBackoff:     cfg.HealthMaxBackoff,
		Multiplier:     cfg.HealthMultiplier,
		JitterPct:      cfg.HealthJitterPct,
	}, false)

	return (*HealthyClient)(watcher)
}

type ClientTTL struct {
	*HealthyClient
	mu       sync.Mutex
	lastUsed time.Time
	ttl      time.Duration
}

func (c *ClientTTL) touch() {
	c.mu.Lock()
	c.lastUsed = time.Now()
	c.mu.Unlock()
}

func (c *ClientTTL) expired() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(c.lastUsed) > c.ttl
}

func NewClientTTL(healthyClient *HealthyClient, ttl time.Duration) *ClientTTL {
	return &ClientTTL{
		HealthyClient: healthyClient,
		lastUsed:      time.Now(),
		ttl:           ttl,
	}
}

type ClientManager struct {
	clients    sync.Map
	tick       time.Duration
	stopCh     chan struct{}
	onCloseErr func(error)
}

func NewClientManager(tick time.Duration) *ClientManager {
	if tick <= 0 {
		tick = 10 * time.Second
	}

	m := &ClientManager{
		tick:    tick,
		clients: sync.Map{},
		stopCh:  make(chan struct{}),
	}

	go m.cleanupTTL()
	return m
}

func (m *ClientManager) GetOrCreate(
	dbID uuid.UUID,
	factory func() (*ClientTTL, error),
) (*ClientTTL, error) {
	if val, ok := m.clients.Load(dbID); ok {
		c := val.(*ClientTTL)
		c.touch()
		return c, nil
	}

	newClient, err := factory()
	if err != nil {
		return nil, err
	}

	actual, loaded := m.clients.LoadOrStore(dbID, newClient)
	if loaded {
		if err := newClient.Client.Close(); err != nil && m.onCloseErr != nil {
			m.onCloseErr(err)
		}
		c := actual.(*ClientTTL)
		c.touch()
		return c, nil
	}

	return newClient, nil
}

func (m *ClientManager) cleanupTTL() {
	ticker := time.NewTicker(m.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.clients.Range(func(key, value any) bool {
				wrapper := value.(*ClientTTL)
				if wrapper.expired() {
					if err := wrapper.Client.Close(); err != nil && m.onCloseErr != nil {
						m.onCloseErr(err)
					}
					m.clients.Delete(key)
				}
				return true
			})

		case <-m.stopCh:
			return
		}
	}
}

func (m *ClientManager) CloseAll() error {
	close(m.stopCh)

	var errs []error
	m.clients.Range(func(key, value any) bool {
		wrapper := value.(*ClientTTL)
		if wrapper.Client != nil {
			if err := wrapper.Client.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		return true
	})

	m.clients = sync.Map{}
	return errors.JoinWrap(errs...)
}
