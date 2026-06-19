package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const persistentMFAKeyFile = "mfa-encryption-key"

type Config struct {
	DatabaseURL      string
	ServerPort       string
	Environment      string
	MFAEncryptionKey string
	// V1CompatSunset is an ISO 8601 date (YYYY-MM-DD) after which v1 API routes
	// are considered retired. When set, the server logs a warning at startup.
	// Configure via V1_COMPAT_SUNSET env var.
	V1CompatSunset string
}

func Load() *Config {
	env := getEnv("ENVIRONMENT", "development")
	isProd := env != "development" && env != "test"

	mfaKey := loadMFAEncryptionKey(getEnv("MFA_ENCRYPTION_KEY", ""), isProd)
	dbURL := getEnv("DATABASE_URL", "postgres://ipam:ipam@localhost:5432/ipam")

	if isProd {
		warnWeakDBCredentials(dbURL)
		warnInsecureDBSSL(dbURL)
	}

	return &Config{
		DatabaseURL:      dbURL,
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		Environment:      env,
		MFAEncryptionKey: mfaKey,
		V1CompatSunset:   strings.TrimSpace(getEnv("V1_COMPAT_SUNSET", "")),
	}
}

func loadMFAEncryptionKey(key string, isProd bool) string {
	key = strings.TrimSpace(key)
	if key != "" || !isProd {
		return validateMFAKey(key, isProd)
	}

	persisted, path, err := readOrCreatePersistentMFAKey()
	if err != nil {
		log.Fatalf("FATAL: failed to initialize persistent MFA encryption key: %v", err)
	}
	log.Printf("INFO: Using persistent MFA encryption key from %s.", path)
	return validateMFAKey(persisted, isProd)
}

// validateMFAKey checks that the key is a valid 64-char hex string with sufficient
// entropy. In production, it fatals on any violation. In development/test it falls
// back to a random per-process key with a warning.
func validateMFAKey(key string, isProd bool) string {
	fail := func(msg string) string {
		if isProd {
			log.Fatalf("FATAL: %s", msg)
		}
		fallback, err := generateEphemeralMFAKey()
		if err != nil {
			log.Fatalf("FATAL: failed to generate development MFA key: %v", err)
		}
		log.Printf("WARNING: %s Generated an ephemeral development MFA key. Set MFA_ENCRYPTION_KEY to preserve MFA secrets across restarts.", msg)
		return fallback
	}

	if len(key) != 64 {
		return fail("MFA_ENCRYPTION_KEY must be a 64-character hex string (32 bytes). Generate with: openssl rand -hex 32")
	}

	b, err := hex.DecodeString(key)
	if err != nil {
		return fail("MFA_ENCRYPTION_KEY is not valid hex.")
	}

	if isWeakKey(b) {
		return fail("MFA_ENCRYPTION_KEY appears to be a weak or default value (all identical bytes). Set a random key.")
	}

	return key
}

func generateEphemeralMFAKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func readOrCreatePersistentMFAKey() (string, string, error) {
	dir, path := PersistentMFAKeyDir(), PersistentMFAKeyPath()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", path, fmt.Errorf("creating %s: %w", dir, err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", path, fmt.Errorf("opening %s: %w", dir, err)
	}
	defer root.Close()

	if data, err := root.ReadFile(persistentMFAKeyFile); err == nil {
		return strings.TrimSpace(string(data)), path, nil
	} else if !os.IsNotExist(err) {
		return "", path, fmt.Errorf("reading %s: %w", path, err)
	}

	key, err := generateEphemeralMFAKey()
	if err != nil {
		return "", path, err
	}
	if err := root.WriteFile(persistentMFAKeyFile, []byte(key+"\n"), 0600); err != nil {
		return "", path, fmt.Errorf("writing %s: %w", path, err)
	}
	return key, path, nil
}

func PersistentMFAKeyDir() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, "data")
}

func PersistentMFAKeyPath() string {
	return filepath.Join(PersistentMFAKeyDir(), persistentMFAKeyFile)
}

func HasPersistentMFAKey() bool {
	if strings.TrimSpace(os.Getenv("MFA_ENCRYPTION_KEY")) != "" {
		return true
	}
	root, err := os.OpenRoot(PersistentMFAKeyDir())
	if err != nil {
		return false
	}
	defer root.Close()
	_, err = root.Stat(persistentMFAKeyFile)
	return err == nil
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

// warnInsecureDBSSL fatals when DATABASE_URL explicitly disables SSL in production.
// sslmode=disable transmits credentials in plaintext.
func warnInsecureDBSSL(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if strings.EqualFold(u.Query().Get("sslmode"), "disable") {
		log.Fatalf(
			"FATAL: DATABASE_URL sets sslmode=disable. " +
				"Remove sslmode=disable or set sslmode=require (or stronger) before running in production.",
		)
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
