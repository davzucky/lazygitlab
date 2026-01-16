package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger     *Logger
	instanceMu sync.Mutex
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
		logger.debug = debug
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
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return logger
}

func Debug(format string, v ...interface{}) {
	if logger == nil || !logger.debug || logger.file == nil {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.file.WriteString("[DEBUG] " + fmt.Sprintf(format, v...) + "\n")
	logger.file.Sync()
}

func Info(format string, v ...interface{}) {
	if logger == nil || logger.file == nil {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.file.WriteString("[INFO] " + fmt.Sprintf(format, v...) + "\n")
	logger.file.Sync()
}

func Error(format string, v ...interface{}) {
	if logger == nil || logger.file == nil {
		return
	}
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.file.WriteString("[ERROR] " + fmt.Sprintf(format, v...) + "\n")
	logger.file.Sync()
}

func Close() error {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	if logger == nil || logger.file == nil {
		logger = nil
		return nil
	}

	if err := logger.file.Close(); err != nil {
		logger = nil
		return err
	}
	logger = nil
	return nil
}
