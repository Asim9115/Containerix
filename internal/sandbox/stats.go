package sandbox

import (
	"fmt"
	"strconv"
)

func (s *Sandbox) Stats() Stats {
	return Stats{
		Cpu : s.Cpu, 
		UsedCpu: s.UsedCpu,
		Memory: s.Memory,
		UsedMemory: s.UsedMemory,
		Containers: len(s.Containers),
	}
}

func (s *Sandbox) CheckResource(cpuNeeded float64, memory string) error {

	usedMemory, err := strconv.ParseInt(s.UsedMemory, 10, 64)
	if err != nil {
		return err
	}
	totalMemory , err := strconv.ParseInt(s.Memory, 10, 64)
	if err != nil {
		return err
	}

	memoryNeeded , err := strconv.ParseInt(memory, 10, 64)
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