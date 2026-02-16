package clickhouse

import (
	"context"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/heartbeat"
)

type ClientAliveCapable struct{ clickhouse.Conn }

func (c *ClientAliveCapable) Alive(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

type ClientWrapper struct {
	*heartbeat.Watcher[*ClientAliveCapable]
	lastUsed time.Time
	ttl      time.Duration
}

func NewClientWrapper(managerTTL time.Duration, cfg *config.Clickhouse) *ClientWrapper {
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

	watcher := heartbeat.NewWatcher(aliveCapable, nil, &heartbeat.Options{
		PingTimeout:    cfg.HealthPingTimeout,
		InitialBackoff: cfg.HealthInitialBackoff,
		MaxBackoff:     cfg.HealthMaxBackoff,
		Multiplier:     cfg.HealthMultiplier,
		JitterPct:      cfg.HealthJitterPct,
	})

	return &ClientWrapper{
		Watcher: watcher,
		ttl:     managerTTL,
	}
}

type ClientManager struct {
	clients sync.Map
	tick    time.Duration
	stopCh  chan struct{}
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

/*
	func (m *ClientManager) GetClient(dbID uuid.UUID) (clickhouse.Conn, error) {
		val, ok := m.clients.Load(dbID)
		if !ok {
			return nil, errors.WrapMsg("unknown database")
		}

		w := val.(*ClientWrapper)

		if w.Client != nil {
			w.lastUsed = time.Now()
			return w.Client, nil
		}

		conn, err := clickhouse.Open(&clickhouse.Options{
			Addr: []string{w.dsn},
		})
		if err != nil {
			return nil, err
		}

		w.Client = conn
		w.lastUsed = time.Now()

		return conn, nil
	}
*/

func (m *ClientManager) cleanupTTL() {
	ticker := time.NewTicker(m.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.clients.Range(func(key, value any) bool {
				wrapper := value.(*ClientWrapper)
				if wrapper.Client != nil && time.Since(wrapper.lastUsed) > wrapper.ttl {
					wrapper.Client.Close()
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
		wrapper := value.(*ClientWrapper)
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
