package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr             string
	DBPath                 string
	SandboxName            string
	SandboxCPU             float64
	SandboxMemory          string
	AdminAPIKey            string
	CORSOrigins            []string
	AllowRegistration      bool
	MaxRequestBody         int64
	GlobalRateLimit        int
	GlobalRateWindow       time.Duration
	RegistrationRateLimit  int
	RegistrationRateWindow time.Duration
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		ListenAddr:             getEnv("CONTAINERIX_LISTEN", ":8080"),
		DBPath:                 getEnv("CONTAINERIX_DB_PATH", "data/containerix.db"),
		SandboxName:            getEnv("CONTAINERIX_SANDBOX_NAME", "containerix"),
		SandboxCPU:             getEnvFloat("CONTAINERIX_SANDBOX_CPU", 2),
		SandboxMemory:          getEnv("CONTAINERIX_SANDBOX_MEMORY", "3221225472"),
		AdminAPIKey:            os.Getenv("CONTAINERIX_ADMIN_API_KEY"),
		CORSOrigins:            getEnvCSV("CONTAINERIX_CORS_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		AllowRegistration:      getEnvBool("CONTAINERIX_ALLOW_REGISTRATION", true),
		MaxRequestBody:         int64(getEnvInt("CONTAINERIX_MAX_REQUEST_BODY", 1<<20)),
		GlobalRateLimit:        int(getEnvInt("CONTAINERIX_GLOBAL_RATE_LIMIT", 120)),
		GlobalRateWindow:       time.Duration(getEnvInt("CONTAINERIX_GLOBAL_RATE_WINDOW", 60)) * time.Second,
		RegistrationRateLimit:  int(getEnvInt("CONTAINERIX_REGISTRATION_RATE_LIMIT", 5)),
		RegistrationRateWindow: time.Duration(getEnvInt("CONTAINERIX_REGISTRATION_RATE_WINDOW", 3600)) * time.Second,
	}
}

// loadDotEnv sets KEY=VALUE from a .env file when the key is not already in the environment.
// Existing env vars always win. Missing file is ignored.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	val, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvInt(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	val, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}

	return val
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return val
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvCSV(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
