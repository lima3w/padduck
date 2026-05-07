package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL      string
	ServerPort       string
	Environment      string
	MFAEncryptionKey string
}

func Load() *Config {
	env := getEnv("ENVIRONMENT", "development")
	mfaKey := getEnv("MFA_ENCRYPTION_KEY", "")
	if len(mfaKey) != 64 {
		if env != "development" && env != "test" {
			log.Fatalf("FATAL: MFA_ENCRYPTION_KEY must be set to a 64-character hex string in %s environment; refusing to start.", env)
		}
		// Development/test fallback — never use in production
		mfaKey = "0000000000000000000000000000000000000000000000000000000000000000"
		log.Println("WARNING: MFA_ENCRYPTION_KEY not set or invalid; using insecure default. Set a 64-char hex key in production.")
	}
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://ipam:ipam@localhost:5432/ipam"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		MFAEncryptionKey: mfaKey,
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
