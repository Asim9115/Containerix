package sandbox

func (s *Sandbox) Stats() Stats {
	return Stats{
		Cpu : s.Cpu,
		UsedCpu: s.UsedCpu,
		Memory: s.Memory,
		UsedMemory: s.UsedMemory,
		Containers: len(s.Containers),
	}
}