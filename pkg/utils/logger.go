package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger     *Logger
	once       sync.Once
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

	var initErr error
	once.Do(func() {
		logDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "lazygitlab")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		logPath := filepath.Join(logDir, "debug.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		logger = &Logger{
			file:  file,
			debug: debug,
		}
	})

	return initErr
}

func GetLogger() *Logger {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return logger
}

func Debug(format string, v ...interface{}) {
	if logger != nil && logger.debug {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.file.WriteString("[DEBUG] " + fmt.Sprintf(format, v...) + "\n")
		logger.file.Sync()
	}
}

func Info(format string, v ...interface{}) {
	if logger != nil {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.file.WriteString("[INFO] " + fmt.Sprintf(format, v...) + "\n")
		logger.file.Sync()
	}
}

func Error(format string, v ...interface{}) {
	if logger != nil {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.file.WriteString("[ERROR] " + fmt.Sprintf(format, v...) + "\n")
		logger.file.Sync()
	}
}

func Close() error {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	if logger != nil && logger.file != nil {
		return logger.file.Close()
	}
	return nil
}
