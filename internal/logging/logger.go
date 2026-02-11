package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func New(debug bool) (*log.Logger, func(), error) {
	if !debug {
		logger := log.New(io.Discard, "", log.LstdFlags)
		return logger, func() {}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	dir := filepath.Join(home, ".local", "share", "lazygitlab")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, nil, err
	}

	path := filepath.Join(dir, "debug.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, nil, err
	}

	logger := log.New(f, "", log.LstdFlags|log.Lmicroseconds)
	return logger, func() { _ = f.Close() }, nil
}
