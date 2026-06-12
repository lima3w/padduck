package utils

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(passwordHash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}

// dummyHash is a bcrypt hash of a random value, computed once at startup with
// the same cost as real password hashes. No supplied password can match it.
var dummyHash = func() []byte {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		panic(err)
	}
	hash, err := bcrypt.GenerateFromPassword(random, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

// DummyVerifyPassword burns the same CPU time as a real bcrypt comparison and
// always returns false. Call it on login paths that fail before reaching the
// password check (unknown username, no password set) so response timing does
// not reveal whether an account exists.
func DummyVerifyPassword(password string) bool {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
	return false
}
