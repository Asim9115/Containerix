package runner

import (
	"fmt"
	"net"
)

func Run(imageTag string, slug string) (ContainerInfo, error) {
	port, err := getFreePort(8000, 8999)
	if err !=nil {
		return ContainerInfo, err
	}
	
}

func getFreePort(min, max int) (int, error) {
	for port := min; port <= max; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", address)

		if err != nil {
			//port already in use
			continue
		}
		ln.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", min, max)
}