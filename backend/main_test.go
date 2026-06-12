package main

import (
	"log/slog"
	"testing"
)

func TestLogLevelFromEnv(t *testing.T) {
	tests := []struct {
		raw       string
		wantLevel slog.Level
		wantKnown bool
	}{
		{"", slog.LevelWarn, true},
		{"warn", slog.LevelWarn, true},
		{"warning", slog.LevelWarn, true},
		{"WARN", slog.LevelWarn, true},
		{"info", slog.LevelInfo, true},
		{"Info", slog.LevelInfo, true},
		{"debug", slog.LevelDebug, true},
		{"error", slog.LevelError, true},
		{"  info  ", slog.LevelInfo, true},
		{"verbose", slog.LevelWarn, false},
		{"3", slog.LevelWarn, false},
	}
	for _, tt := range tests {
		level, known := logLevelFromEnv(tt.raw)
		if level != tt.wantLevel || known != tt.wantKnown {
			t.Errorf("logLevelFromEnv(%q) = (%v, %v), want (%v, %v)",
				tt.raw, level, known, tt.wantLevel, tt.wantKnown)
		}
	}
}
