package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func recordHistory(path string, command string) error {
	histFile := os.Getenv("ILC_HISTFILE")
	if histFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get user home directory: %w", err)
		}
		histFile = filepath.Join(home, ".ilc_history")
	}

	if histFile == "-" {
		return nil
	}

	histSizeStr := os.Getenv("ILC_HISTSIZE")
	histSize, err := strconv.Atoi(histSizeStr)
	if err != nil {
		// Default to -1 (no limit) if not set or invalid
		histSize = -1
	}

	if histSize == 0 {
		return nil
	}

	f, err := os.OpenFile(histFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open history file: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s: %s\n", path, command); err != nil {
		return fmt.Errorf("could not write to history file: %w", err)
	}

	if histSize > 0 {
		return truncateHistory(histFile, histSize)
	}

	return nil
}

func truncateHistory(histFile string, histSize int) error {
	f, err := os.Open(histFile)
	if err != nil {
		return fmt.Errorf("could not open history file for truncation: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > histSize {
		lines = lines[len(lines)-histSize:]
		output := strings.Join(lines, "\n") + "\n"
		return os.WriteFile(histFile, []byte(output), 0644)
	}

	return nil
}
