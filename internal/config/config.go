package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr               string
	DBPath                   string
	SandboxName              string
	SandboxCPU               float64
	SandboxMemory            string
	AdminAPIKey              string
	AllowRegistration        bool
	MaxRequestBody           int64
	GlobalRateLimit          int
	GlobalRateWindow         time.Duration
	RegistrationRateLimit    int
	RegistrationRateWindow   time.Duration
}

func Load() *Config {
	return &Config{
		ListenAddr:             getEnv("CONTAINERIX_LISTEN", ":8080"),
		DBPath:                 getEnv("CONTAINERIX_DB_PATH", "data/containerix.db"),
		SandboxName:            getEnv("CONTAINERIX_SANDBOX_NAME", "containerix"),
		SandboxCPU:             getEnvFloat("CONTAINERIX_SANDBOX_CPU", 2),
		SandboxMemory:          getEnv("CONTAINERIX_SANDBOX_MEMORY", "3221225472"),
		AdminAPIKey:            os.Getenv("CONTAINERIX_ADMIN_API_KEY"),
		AllowRegistration:      getEnvBool("CONTAINERIX_ALLOW_REGISTRATION", true),
		MaxRequestBody:         int64(getEnvInt("CONTAINERIX_MAX_REQUEST_BODY", 1<<20)),
		GlobalRateLimit:        int(getEnvInt("CONTAINERIX_GLOBAL_RATE_LIMIT", 120)),
		GlobalRateWindow:       time.Duration(getEnvInt("CONTAINERIX_GLOBAL_RATE_WINDOW", 60)) * time.Second,
		RegistrationRateLimit:  int(getEnvInt("CONTAINERIX_REGISTRATION_RATE_LIMIT", 5)),
		RegistrationRateWindow: time.Duration(getEnvInt("CONTAINERIX_REGISTRATION_RATE_WINDOW", 3600)) * time.Second,
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