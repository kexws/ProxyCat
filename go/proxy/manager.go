package proxy

import (
	"sync"
)

// Proxy represents a single proxy server.
type Proxy struct {
	ID      string
	Address string
}

// Manager keeps track of active and failed proxies.
type Manager struct {
	mu      sync.RWMutex
	active  map[string]Proxy
	failed  map[string]Proxy
	checker func(string) bool
}

// NewManager creates a Manager that uses checker to verify proxies.
func NewManager(checker func(string) bool) *Manager {
	return &Manager{
		active:  make(map[string]Proxy),
		failed:  make(map[string]Proxy),
		checker: checker,
	}
}

// Add adds a proxy, placing it in active or failed based on checker.
func (m *Manager) Add(address string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := Proxy{ID: address, Address: address}
	if m.checker(address) {
		m.active[p.ID] = p
	} else {
		m.failed[p.ID] = p
	}
}

// Active returns all active proxies.
func (m *Manager) Active() []Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]Proxy, 0, len(m.active))
	for _, p := range m.active {
		res = append(res, p)
	}
	return res
}

// Failed returns all failed proxies.
func (m *Manager) Failed() []Proxy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]Proxy, 0, len(m.failed))
	for _, p := range m.failed {
		res = append(res, p)
	}
	return res
}

// Retry checks a failed proxy again. If successful, it moves to active.
func (m *Manager) Retry(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.failed[id]
	if !ok {
		return false
	}
	if m.checker(p.Address) {
		delete(m.failed, id)
		m.active[id] = p
		return true
	}
	return false
}

// Delete removes a failed proxy by id.
func (m *Manager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.failed, id)
}

// DeleteAllFailed clears all failed proxies.
func (m *Manager) DeleteAllFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failed = make(map[string]Proxy)
}
