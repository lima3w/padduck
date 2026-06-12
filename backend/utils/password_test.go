package utils

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Error("VerifyPassword rejected the correct password")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Error("VerifyPassword accepted a wrong password")
	}
}

func TestDummyVerifyPasswordAlwaysFalse(t *testing.T) {
	for _, password := range []string{"", "password", "padduck-timing-equalization-dummy"} {
		if DummyVerifyPassword(password) {
			t.Errorf("DummyVerifyPassword(%q) returned true", password)
		}
	}
}
