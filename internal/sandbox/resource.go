package sandbox

import (
	"fmt"
	"strconv"
	"github.com/asim9115/containerix/internal/types"
)

func (s *SandboxManager)CanAllocate(cpuNeeded float64, memory string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	usedMemory, err := strconv.ParseInt(s.UsedMemory, 10, 64)
	if err != nil {
		return err
	}
	totalMemory, err := strconv.ParseInt(s.Memory, 10, 64)
	if err != nil {
		return err
	}

	memoryNeeded, err := strconv.ParseInt(memory, 10, 64)
	if err != nil {
		return err
	}

	cpuRemaining := s.Cpu - s.UsedCpu

	if cpuRemaining < cpuNeeded {
		return fmt.Errorf("insufficient CPU: requested %f, available %f", cpuNeeded, cpuRemaining)
	}

	memoryAvailable := totalMemory - usedMemory
	if memoryNeeded > memoryAvailable {
		return fmt.Errorf("insufficient memory: requested %d bytes, available %d bytes", memoryNeeded, memoryAvailable)
	}

	return nil
}

func (s *SandboxManager) Release(cpu float64, memory string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	newCpu := s.UsedCpu - cpu
	if newCpu < 0 {
		newCpu = 0
	}

	var freeMemory int64
	if memory != "" {
		memBytes, err := types.MemoryToBytes(memory)
		if err != nil {
			memBytes = memory
		}
		if parsed, err := strconv.ParseInt(memBytes, 10, 64); err == nil {
			freeMemory = parsed
		}
	}

	usedMemory, err := strconv.ParseInt(s.UsedMemory, 10, 64)
	if err != nil {
		usedMemory = 0
	}

	newMemory := usedMemory - freeMemory
	if newMemory < 0 {
		newMemory = 0
	}

	s.UsedCpu = newCpu
	s.UsedMemory = strconv.FormatInt(newMemory, 10)
	return nil
}

func (s *SandboxManager) Allocate(cpu float64, memory string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	newCpu := s.UsedCpu + cpu
	addMemory, err := strconv.Atoi(memory)
	if err != nil {
		return err
	}
	usedMemory, err := strconv.Atoi(s.UsedMemory)
	if err != nil {
		return err
	}
	newMemory := usedMemory + addMemory
	s.UsedCpu = newCpu
	s.UsedMemory = strconv.Itoa(newMemory)
	return nil
}

func (s *SandboxManager) Remaining() types.Stats {
	return s.Stats()
}
