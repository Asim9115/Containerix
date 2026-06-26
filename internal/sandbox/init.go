package sandbox

import (
	"fmt"
)
//create a sandbox type environment to run containers with given resources
func Init(name string, cpu int, memory string) (*Sandbox, error) {
	sb = &Sandbox{
		Name: name,
		CPU: cpu,
		Memory: memory,
	}

}