package container

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
	usedPorts map[int]PortAllocation
	mu        sync.Mutex
}

type PortAllocation struct {
	ContainerId   string
	ContainerPort int
}

type PortManager interface {
	AllocatePort(containerId string, hostPort int, containerPort int) (int, error)
	GetFreePort() (int, error)
	Reserve(containerId string, hostPort int, containerPort int) error
	ReleasePort(hostPort int)
	ReleaseAll()
	IsUsed(hostPort int) bool
	GetAllocation(hostPort int) (PortAllocation, bool)
}

func New() *Manager {
	return &Manager{
		usedPorts: make(map[int]PortAllocation),
	}
}

func (m *Manager) AllocatePort(containerId string, hostPort int, containerPort int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//auto assign free port
	if hostPort == 0 {
		freePort, err := m.getFreePortLocked()
		if err != nil {
			return 0, err
		}
		hostPort = freePort
	}

	if _, exists := m.usedPorts[hostPort]; exists {
		return 0, fmt.Errorf("host port %d already allocated", hostPort)
	}

	address := fmt.Sprintf(":%d", hostPort)

	ln, err := net.Listen("tcp", address)
	if err != nil {
		return 0, fmt.Errorf("host port %d already in used", hostPort)
	}
	ln.Close()

	m.usedPorts[hostPort] = PortAllocation{
		ContainerId:   containerId,
		ContainerPort: containerPort,
	}
	return hostPort, nil
}

func (m *Manager) IsUsed(port int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.usedPorts[port]
	return exists
}

func (m *Manager) ReleasePort(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.usedPorts, port)
}

func (m *Manager) ReleaseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.usedPorts)
}

func (m *Manager) GetFreePort() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getFreePortLocked()
}

func (m *Manager) getFreePortLocked() (int, error) {
	for port := StartPort; port <= EndPort; port++ {
		if _, exists := m.usedPorts[port]; exists {
			continue
		}

		address := fmt.Sprintf(":%d", port)

		ln, err := net.Listen("tcp", address)
		if err != nil {
			continue
		}
		ln.Close()

		return port, nil
	}
	return 0, fmt.Errorf("NO free ports available")
}

func (m *Manager) Reserve(containerId string, hostPort int, containerPort int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.usedPorts[hostPort]; exists {
		return fmt.Errorf("host port %d already reserved", hostPort)
	}

	m.usedPorts[hostPort] = PortAllocation{
		ContainerId:   containerId,
		ContainerPort: containerPort,
	}

	return nil
}

func (m *Manager) GetAllocation(hostPort int) (PortAllocation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	allocation, exists := m.usedPorts[hostPort]
	return allocation, exists
}