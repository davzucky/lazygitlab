package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	testLogDir := filepath.Join(os.TempDir(), "lazygitlab-test")
	os.Setenv("HOME", testLogDir)
	defer os.Unsetenv("HOME")

	if err := InitLogger(true); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()

	if GetLogger() == nil {
		t.Error("Logger should not be nil after initialization")
	}

	logPath := filepath.Join(testLogDir, ".local", "share", "lazygitlab", "debug.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should be created after initialization")
	}

	os.RemoveAll(testLogDir)
}

func TestLoggerDoesNotCrash(t *testing.T) {
	testLogDir := filepath.Join(os.TempDir(), "lazygitlab-test-nocrash")
	os.Setenv("HOME", testLogDir)

	if err := InitLogger(true); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Debug("test debug message: %s", "value")
	Info("test info message: %s", "value")
	Error("test error message: %s", "value")

	Close()
	os.RemoveAll(testLogDir)
}
