package pgx

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolManager struct {
	mu   sync.Mutex
	pool map[string]*pgxpool.Pool
}

var manager *PoolManager
var once sync.Once

// Singleton
func New() *PoolManager {
	once.Do(func() {
		manager = &PoolManager{
			pool: make(map[string]*pgxpool.Pool),
		}
	})
	return manager
}

func (m *PoolManager) GetPool(ctx context.Context, connID string, dsn string) (*pgxpool.Pool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if exists — reuse
	if p, ok := m.pool[connID]; ok {
		return p, nil
	}

	// create new pool
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// performance: low idle
	cfg.MaxConns = 3
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Minute * 30
	cfg.MaxConnIdleTime = time.Minute * 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// store pool
	m.pool[connID] = pool

	return pool, nil
}

func (m *PoolManager) CloseConnID(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.pool[connID]; ok {
		p.Close()
		delete(m.pool, connID)
	}
}

func (m *PoolManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, p := range m.pool {
		p.Close()
		delete(m.pool, k)
	}
}
