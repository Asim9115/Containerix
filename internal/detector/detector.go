package detector

import (
	"fmt"
	"log"
	"net"

	"time"
)


var CommonPorts = []int{
	80,    // HTTP
	3000,  // React, Express, Next.js
	4000,  // Phoenix, GraphQL, various dev servers
	4173,  // Vite preview
	5000,  // Flask, ASP.NET, various apps
	5173,  // Vite dev server
	5500,  // Live Server (VS Code)
	6006,  // Storybook
	7000,  // Misc dev servers
	8000,  // Django, FastAPI, Python HTTP server
	8080,  // Spring Boot, Go, Java, Tomcat
	8081,  // Alternate HTTP/dev server
	8088,  // Misc web apps
	8888,  // Jupyter Notebook
	9000,  // PHP-FPM, SonarQube, dev servers
	9090,  // Prometheus, Go apps
	10000, // Render/default app port
}

func ScanActivePort(ip string) (int, error) {
	if ip == "" {
		return 0, fmt.Errorf("container IP is empty, cannot scan")
	}
	log.Printf("Starting port scan on IP: %s", ip)
	
	// 1. Give the app 2 seconds to boot up inside the container
	time.Sleep(2 * time.Second)
	
	// 2. Retry up to 5 times for slower booting frameworks
	for retries := 0; retries < 5; retries++ {
		for _, port := range CommonPorts {
			address := fmt.Sprintf("%v:%v", ip, port)
			
			// Attempt a TCP connection with a fast timeout
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)
			if err == nil {
				conn.Close()
				log.Printf("Successfully connected to %s", address)
				return port, nil // Found the active port!
			}
		}
		log.Printf("Retry %d: No ports open yet on %s. Waiting...", retries+1, ip)
		time.Sleep(2 * time.Second)
	}
	
	return 0, fmt.Errorf("could not detect active port after probing")
}