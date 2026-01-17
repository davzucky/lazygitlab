package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// CopyToClipboard copies the given text to the system clipboard
func CopyToClipboard(text string) error {
	if text == "" {
		return fmt.Errorf("empty text to copy")
	}

	switch runtime.GOOS {
	case "linux":
		return copyToClipboardLinux(text)
	case "darwin":
		return copyToClipboardMac(text)
	case "windows":
		return copyToClipboardWindows(text)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func copyToClipboardLinux(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		if isCommandNotFound(err) {
			return fmt.Errorf("xclip not found. Please install xclip: sudo apt install xclip")
		}
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

func copyToClipboardMac(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		if isCommandNotFound(err) {
			return fmt.Errorf("pbcopy command failed")
		}
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

func copyToClipboardWindows(text string) error {
	cmd := exec.Command("cmd", "/c", "clip")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

func isCommandNotFound(err error) bool {
	if execErr, ok := err.(*exec.Error); ok {
		return execErr.Err == exec.ErrNotFound
	}
	return false
}
