package container

type Config struct {
	Name    string
	Image   string
	Cpu     float64
	Memory  string
	Ports   []PortMapping
	Env     map[string]string
	cmd     []string
	Volumes []VolumeMount
}

type PortMapping struct {
	HostPort      int
	ContainerPort int
}

type VolumeMount struct {
	HostPath      string
	ContainerPath string
}

type Container struct {
	ID     string
	Config Config
	Status string
}
