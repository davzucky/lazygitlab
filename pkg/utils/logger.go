package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger     *Logger
	instanceMu sync.RWMutex
)

type Logger struct {
	file  *os.File
	debug bool
	mu    sync.Mutex
}

func InitLogger(debug bool) error {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	if logger != nil {
		logger.mu.Lock()
		logger.debug = debug
		logger.mu.Unlock()
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	if homeDir == "" {
		return fmt.Errorf("failed to determine home directory")
	}

	logDir := filepath.Join(homeDir, ".local", "share", "lazygitlab")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "debug.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logger = &Logger{
		file:  file,
		debug: debug,
	}

	return nil
}

func GetLogger() *Logger {
	instanceMu.RLock()
	defer instanceMu.RUnlock()
	return logger
}

func Debug(format string, v ...interface{}) {
	l := GetLogger()
	if l == nil || !l.debug {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return
	}
	l.file.WriteString("[DEBUG] " + fmt.Sprintf(format, v...) + "\n")
	l.file.Sync()
}

func Info(format string, v ...interface{}) {
	l := GetLogger()
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return
	}
	l.file.WriteString("[INFO] " + fmt.Sprintf(format, v...) + "\n")
	l.file.Sync()
}

func Error(format string, v ...interface{}) {
	l := GetLogger()
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return
	}
	l.file.WriteString("[ERROR] " + fmt.Sprintf(format, v...) + "\n")
	l.file.Sync()
}

func Close() error {
	instanceMu.Lock()
	l := logger
	logger = nil
	instanceMu.Unlock()

	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}
