package ports

import (
	"fmt"
	"net"
	"sync"
)

const (
	StartPort = 40000
	EndPort   = 50000
)

type Manager struct {
	usedPorts map[int]struct{}
	mu        sync.Mutex
}

type PortManager interface {
	GetFreePortAndReserve() (int, error)
	MarkAsUsed(hostPort int) 
	MarkFree(hostPort int)
	ReleaseAll()
	IsUsed(hostPort int) bool
}

func New() *Manager {
	return &Manager{
		usedPorts: make(map[int]struct{}),
	}
}


func (m *Manager) GetFreePortAndReserve() (int , error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for port := StartPort;  port <= EndPort; port++ {
		if _, exists := m.usedPorts[port]; exists {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue //os is using this port
		}
		ln.Close()
		m.usedPorts[port] = struct{}{}
		return port, nil
	}
	return 0, fmt.Errorf("no free ports avilable in range %d-%d", StartPort, EndPort)
}

func (m *Manager) MarkFree(hostPort int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.usedPorts, hostPort)
}

func (m *Manager) MarkAsUsed(hostPort int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usedPorts[hostPort] = struct{}{}
}

func (m *Manager) IsUsed(hostPort int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.usedPorts[hostPort]
	return exists
}

func (m *Manager) ReleaseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.usedPorts)
}

