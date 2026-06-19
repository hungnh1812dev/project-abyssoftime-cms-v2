package grpcclient

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Manager struct {
	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn
}

func NewManager() *Manager {
	return &Manager{conns: make(map[string]*grpc.ClientConn)}
}

func (m *Manager) Connect(ctx context.Context, name, address string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc client %s (%s): %w", name, address, err)
	}
	m.mu.Lock()
	m.conns[name] = conn
	m.mu.Unlock()
	return nil
}

func (m *Manager) GetConnection(name string) (*grpc.ClientConn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[name]
	if !ok {
		return nil, fmt.Errorf("grpc client %q not connected", name)
	}
	return conn, nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, conn := range m.conns {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("close grpc client %s: %w", name, err)
		}
		delete(m.conns, name)
	}
	return nil
}

func ParseServices(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}
