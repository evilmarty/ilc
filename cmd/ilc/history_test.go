package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordHistory(t *testing.T) {
	path := "some/path/to/config"
	command := "echo 'hello world'"
	record := strings.Join([]string{path, command}, ": ")

	// Test case 1: Default history file in a temporary home directory
	t.Run("default history file", func(t *testing.T) {
		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)
		os.Unsetenv("ILC_HISTFILE") // Ensure it's not set

		err := recordHistory(path, command)
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}

		histFile := filepath.Join(tempHome, ".ilc_history")
		content, err := os.ReadFile(histFile)
		if err != nil {
			t.Fatalf("could not read history file: %v", err)
		}

		if !strings.Contains(string(content), record) {
			t.Errorf("history file should contain '%s', but it doesn't. Content: %s", command, content)
		}
	})

	// Test case 2: Custom history file
	t.Run("custom history file", func(t *testing.T) {
		tempDir := t.TempDir()
		histFile := filepath.Join(tempDir, "my_history")
		t.Setenv("ILC_HISTFILE", histFile)

		err := recordHistory(path, command)
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}

		content, err := os.ReadFile(histFile)
		if err != nil {
			t.Fatalf("could not read history file: %v", err)
		}

		if !strings.Contains(string(content), record) {
			t.Errorf("history file should contain '%s', but it doesn't. Content: %s", command, content)
		}
	})

	// Test case 3: History disabled with "-"
	t.Run("history disabled with dash", func(t *testing.T) {
		t.Setenv("ILC_HISTFILE", "-")
		// We can't easily check that a file was *not* created system-wide,
		// so we just check that no error is returned.
		if err := recordHistory(path, command); err != nil {
			t.Fatalf("recordHistory() with ILC_HISTFILE='-' should not return an error, but got %v", err)
		}
	})

	// Test case 5: History disabled with ILC_HISTSIZE=0
	t.Run("history disabled with histsize 0", func(t *testing.T) {
		tempDir := t.TempDir()
		histFile := filepath.Join(tempDir, "my_history")
		t.Setenv("ILC_HISTFILE", histFile)
		t.Setenv("ILC_HISTSIZE", "0")

		err := recordHistory(path, command)
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}

		if _, err := os.Stat(histFile); !os.IsNotExist(err) {
			t.Errorf("history file should not have been created, but it was")
		}
	})

	// Test case 6: History truncated
	t.Run("history truncated", func(t *testing.T) {
		tempDir := t.TempDir()
		histFile := filepath.Join(tempDir, "my_history")
		t.Setenv("ILC_HISTFILE", histFile)
		t.Setenv("ILC_HISTSIZE", "2")

		recordHistory(path, "command 1")
		recordHistory(path, "command 2")
		recordHistory(path, "command 3")

		content, err := os.ReadFile(histFile)
		if err != nil {
			t.Fatalf("could not read history file: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 lines in history file, got %d", len(lines))
		}
		if !strings.Contains(lines[0], "command 2") {
			t.Errorf("expected first line to contain 'command 2', got '%s'", lines[0])
		}
		if !strings.Contains(lines[1], "command 3") {
			t.Errorf("expected second line to contain 'command 3', got '%s'", lines[1])
		}
	})
}
