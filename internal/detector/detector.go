package detector

import (
	"fmt"
	"log"
	"net"

	"time"
)

/*
// OLD STATIC DETECTION LOGIC (Deprecated in favor of Probe & Restart)
var DefaultPorts = map[string]int{
	"django":      8000,
	"flask":       5000,
	"fastapi":     8000,
	"express":     3000,
	"nestjs":      3000,
	"nexltjs":      3000,
	"nuxt":        3000,
	"react":       3000,
	"vue":         5173,
	"angular":     4200,
	"vite":        5173,
	"remix":       3000,
	"spring-boot": 8080,
	"laravel":     8000,
	"rails":       3000,
	"phoenix":     4000,
	"aspnet":      5000,
	"gin":         8080,
	"fiber":       3000,
	"echo":        1323,
	"actix":       8080,
	"rocket":      8000,
}

const (
	LangNode   = "node"
	LangPython = "python"
	LangGo     = "go"
)

type Result struct {
	Language  string
	Framework string
	Port      int
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetInternalPort(path string) (int, error) {

	result := Result{}
	indicators := []struct {
		file     string
		language string
	}{
		{"package.json", LangNode},
		{"requirements.txt", LangPython},
		{"pyproject.toml", LangPython},
		{"setup.py", LangPython},
		{"go.mod", LangGo},
	}
	for _, ind := range indicators {
		if fileExists(filepath.Join(path, ind.file)) {
			result.Language = ind.language
			break
		}
	}

	detectFramework(&result, path)

	framework := result.Framework
	if port, ok := DefaultPorts[framework];ok {
		return port, nil
	} else {
		return 8000, nil
	}
}

func detectFramework(result *Result, path string) {

	if result.Language == LangNode {
		detectNode()

	} else if result.Language == LangPython {
		framework := detectPython(path)
		result.Framework = framework
		return
	} else if result.Language == LangGo {
		result.Port = 8080
		return
	}
}

func detectPython(path string) string {
	if fileExists(filepath.Join(path, "manage.py")) {
		return "django"
	}

	files := []string {
		"requirements.txt", "pyproject.toml", "Pipfile", "setup.py",
	}

	frameworks := []string{
		"django",
		"flask",
		"fastapi",
		"streamlit",
		"dash",
		"tornado",
		"bottle",
		"falcon",
		"pyramid",
		"sanic",
		"quart",
	}
	for _, file := range files {
		newpath := filepath.Join(path, file)
		data, err := os.ReadFile(newpath)
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))

		for _, framework := range frameworks {
			if strings.Contains(content, framework){
				return framework
			}

		}
	}
	return "python"
}

func detectNode() {
	
}
*/


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
			address := fmt.Sprintf("%s:%d", ip, port)
			
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