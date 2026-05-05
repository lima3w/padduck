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
	mfaKey := getEnv("MFA_ENCRYPTION_KEY", "")
	if len(mfaKey) != 64 {
		// 32-byte dev-only default — never use in production
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
