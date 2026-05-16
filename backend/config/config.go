package config

import (
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL      string
	ServerPort       string
	Environment      string
	MFAEncryptionKey string
}

func Load() *Config {
	env := getEnv("ENVIRONMENT", "development")
	isProd := env != "development" && env != "test"

	mfaKey := validateMFAKey(getEnv("MFA_ENCRYPTION_KEY", ""), isProd)
	dbURL := getEnv("DATABASE_URL", "postgres://ipam:ipam@localhost:5432/ipam")

	if isProd {
		warnWeakDBCredentials(dbURL)
	}

	return &Config{
		DatabaseURL:      dbURL,
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		Environment:      env,
		MFAEncryptionKey: mfaKey,
	}
}

const devFallbackKey = "0000000000000000000000000000000000000000000000000000000000000000"

// validateMFAKey checks that the key is a valid 64-char hex string with sufficient
// entropy. In production, it fatals on any violation. In development/test it falls
// back to an insecure all-zero key with a warning.
func validateMFAKey(key string, isProd bool) string {
	fail := func(msg string) string {
		if isProd {
			log.Fatalf("FATAL: %s", msg)
		}
		log.Printf("WARNING: %s Using insecure default — never use in production.", msg)
		return devFallbackKey
	}

	if len(key) != 64 {
		return fail("MFA_ENCRYPTION_KEY must be a 64-character hex string (32 bytes). Generate with: openssl rand -hex 32")
	}

	b, err := hex.DecodeString(key)
	if err != nil {
		return fail("MFA_ENCRYPTION_KEY is not valid hex.")
	}

	if isProd && isWeakKey(b) {
		return fail("MFA_ENCRYPTION_KEY appears to be a weak or default value (all identical bytes). Set a random key.")
	}

	return key
}

// isWeakKey returns true if all bytes are identical (e.g. all zeros).
func isWeakKey(b []byte) bool {
	for _, c := range b[1:] {
		if c != b[0] {
			return false
		}
	}
	return true
}

// warnWeakDBCredentials logs a fatal error when the DATABASE_URL contains
// the known-default credentials shipped in the example compose file.
func warnWeakDBCredentials(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	user := u.User.Username()
	password, hasPassword := u.User.Password()
	if !hasPassword {
		return
	}
	knownDefaults := []string{"ipam", "postgres", "password", "secret", "changeme", "demo"}
	for _, weak := range knownDefaults {
		if strings.EqualFold(password, weak) || strings.EqualFold(user, weak) && password == user {
			log.Fatalf(
				"FATAL: DATABASE_URL contains known-default credentials (%s:***). "+
					"Set strong credentials via DATABASE_URL or POSTGRES_PASSWORD env var before running in production.",
				user,
			)
		}
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
