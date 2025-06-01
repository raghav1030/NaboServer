package connection

import (
	"sync"
	"github.com/gorilla/websocket"
)

type Manager struct {
	connections map[string]*websocket.Conn // userID -> WebSocket connection
	mu          sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*websocket.Conn),
	}
}

func (m *Manager) AddConnection(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[userID] = conn
}

func (m *Manager) RemoveConnection(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, userID)
}

func (m *Manager) GetConnection(userID string) *websocket.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[userID]
}
