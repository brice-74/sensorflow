package clickhouse

import (
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/google/uuid"
)

type ClientWrapper struct {
	conn     clickhouse.Conn
	lastUsed time.Time
	dsn      string
}

type ClientManager struct {
	clients sync.Map
	ttl     time.Duration
	stopCh  chan struct{}
}

func NewClientManager(ttl time.Duration) *ClientManager {
	m := &ClientManager{
		ttl:     ttl,
		clients: sync.Map{},
		stopCh:  make(chan struct{}),
	}

	go m.cleanupTTL()
	return m
}

func (m *ClientManager) UpsertDatabase(dbID uuid.UUID, dsn string) {
	if val, ok := m.clients.Load(dbID); ok {
		w := val.(*ClientWrapper)
		if w.dsn != dsn {
			w.conn.Close()
			m.clients.Delete(dbID)
		}
		return
	}

	m.clients.Store(dbID, &ClientWrapper{
		dsn: dsn,
	})
}

func (m *ClientManager) GetClient(dbID uuid.UUID) (clickhouse.Conn, error) {
	val, ok := m.clients.Load(dbID)
	if !ok {
		return nil, errors.WrapMsg("unknown database")
	}

	w := val.(*ClientWrapper)

	if w.conn != nil {
		w.lastUsed = time.Now()
		return w.conn, nil
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{w.dsn},
	})
	if err != nil {
		return nil, err
	}

	w.conn = conn
	w.lastUsed = time.Now()

	return conn, nil
}

func (m *ClientManager) cleanupTTL() {
	ticker := time.NewTicker(m.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.clients.Range(func(key, value any) bool {
				wrapper := value.(*ClientWrapper)
				if wrapper.conn != nil && time.Since(wrapper.lastUsed) > m.ttl {
					wrapper.conn.Close()
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
		if wrapper.conn != nil {
			if err := wrapper.conn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		return true
	})

	m.clients = sync.Map{}
	return errors.JoinWrap(errs...)
}
