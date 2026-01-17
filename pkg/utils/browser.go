package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenInBrowser opens the given URL in the default web browser
func OpenInBrowser(url string) error {
	if url == "" {
		return fmt.Errorf("empty URL to open")
	}

	switch runtime.GOOS {
	case "linux":
		return openInBrowserLinux(url)
	case "darwin":
		return openInBrowserMac(url)
	case "windows":
		return openInBrowserWindows(url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func openInBrowserLinux(url string) error {
	cmd := exec.Command("xdg-open", url)
	if err := cmd.Start(); err != nil {
		if isCommandNotFound(err) {
			return fmt.Errorf("xdg-open not found. Please install xdg-utils")
		}
		return fmt.Errorf("failed to open browser: %w", err)
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

func openInBrowserMac(url string) error {
	cmd := exec.Command("open", url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}

func openInBrowserWindows(url string) error {
	cmd := exec.Command("cmd", "/c", "start", "", url)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}
