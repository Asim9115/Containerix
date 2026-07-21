package sandbox

import "github.com/asim9115/containerix/internal/types"

func (s *SandboxManager) Stats() types.Stats {
	return types.Stats{
		Cpu:        s.Cpu,
		UsedCpu:    s.UsedCpu,
		Memory:     s.Memory,
		UsedMemory: s.UsedMemory,
		Containers: len(s.Containers),
	}
}

func (s *SandboxManager) GetState() types.Sandbox {
	s.mu.Lock()
	defer s.mu.Unlock()
	return *s.Sandbox
}
