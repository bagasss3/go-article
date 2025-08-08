package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type MockCache struct {
	store map[string][]byte
	mu    sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		store: make(map[string][]byte),
	}
}

func (m *MockCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.store[key] = data
	return nil
}

func (m *MockCache) Get(_ context.Context, key string, target any) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.store[key]
	if !ok {
		return errors.New("cache miss")
	}

	return json.Unmarshal(data, target)
}

func (m *MockCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, key)
	return nil
}
