package utils

import (
	"runtime"
	"testing"
)

func TestCopyToClipboard(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping clipboard test in short mode")
	}

	testText := "Test text for clipboard"

	err := CopyToClipboard(testText)
	if err != nil {
		if runtime.GOOS == "linux" {
			t.Skipf("Skipping test on Linux: %v", err)
		} else {
			t.Fatalf("CopyToClipboard failed: %v", err)
		}
	}

	t.Log("Clipboard copy test completed successfully")
}

func TestCopyToClipboardEmptyText(t *testing.T) {
	err := CopyToClipboard("")
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}
}
