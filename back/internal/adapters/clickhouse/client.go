package clickhouse

import (
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/brice-74/sensorflow/pkg/errors"
)

type ClientWrapper struct {
	Client   clickhouse.Conn
	lastUsed time.Time
	dsn      string
}

type ClientManager struct {
	clients sync.Map
	ttl     time.Duration
}

func NewClientManager(ttl time.Duration) *ClientManager {
	m := &ClientManager{
		ttl:     ttl,
		clients: sync.Map{},
	}

	go m.cleanupTTL()
	return m
}

func (m *ClientManager) GetClient(dsn string) (clickhouse.Conn, error) {
	if val, ok := m.clients.Load(dsn); ok {
		wrapper := val.(*ClientWrapper)
		wrapper.lastUsed = time.Now()
		return wrapper.Client, nil
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
	})
	if err != nil {
		return nil, err
	}

	m.clients.Store(dsn, &ClientWrapper{
		Client:   conn,
		lastUsed: time.Now(),
		dsn:      dsn,
	})

	return conn, nil
}

func (m *ClientManager) cleanupTTL() {
	ticker := time.NewTicker(m.ttl)
	defer ticker.Stop()

	for range ticker.C {
		m.clients.Range(func(key, value any) bool {
			wrapper := value.(*ClientWrapper)
			if time.Since(wrapper.lastUsed) > m.ttl {
				wrapper.Client.Close()
				m.clients.Delete(key)
			}
			return true
		})
	}
}

func (m *ClientManager) CloseAll() error {
	var errs []error
	m.clients.Range(func(key, value any) bool {
		wrapper := value.(*ClientWrapper)
		if err := wrapper.Client.Close(); err != nil {
			errs = append(errs, err)
		}
		return true
	})
	m.clients = sync.Map{}
	return errors.JoinWrap(errs...)
}
