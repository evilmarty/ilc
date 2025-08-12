package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	// Test case 1: Basic help
	t.Run("basic help", func(t *testing.T) {
		output := captureOutput(func() {
			printHelp(nil, nil)
		})
		if !strings.Contains(output, "USAGE:") {
			t.Errorf("expected output to contain 'USAGE:', got '%s'", output)
		}
	})

	// Test case 2: Top-level help
	t.Run("top-level help", func(t *testing.T) {
		cfg := &Config{
			Description: "My awesome CLI",
			Commands: map[string]Command{
				"test": {Description: "Test command"},
			},
			Inputs: map[string]Input{
				"myinput": {Description: "My input"},
			},
		}
		output := captureOutput(func() {
			printHelp(cfg, nil)
		})
		if !strings.Contains(output, "My awesome CLI") {
			t.Errorf("expected output to contain 'My awesome CLI', got '%s'", output)
		}
		if !strings.Contains(output, "test") {
			t.Errorf("expected output to contain 'test', got '%s'", output)
		}
		if !strings.Contains(output, "myinput") {
			t.Errorf("expected output to contain 'myinput', got '%s'", output)
		}
	})

	// Test case 3: Command help
	t.Run("command help", func(t *testing.T) {
		cmd := &Command{
			Description: "My command",
			Commands: map[string]Command{
				"sub": {Description: "Sub command"},
			},
			Inputs: map[string]Input{
				"myinput": {Description: "My input", Type: "string", Default: "default"},
			},
		}
		output := captureOutput(func() {
			printHelp(&Config{}, cmd)
		})
		if !strings.Contains(output, "My command") {
			t.Errorf("expected output to contain 'My command', got '%s'", output)
		}
		if !strings.Contains(output, "sub") {
			t.Errorf("expected output to contain 'sub', got '%s'", output)
		}
		if !strings.Contains(output, "myinput") {
			t.Errorf("expected output to contain 'myinput', got '%s'", output)
		}
		if !strings.Contains(output, "string, default: default") {
			t.Errorf("expected output to contain 'string, default: default', got '%s'", output)
		}
	})
}

// captureOutput captures the stdout of a function.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
