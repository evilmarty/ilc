package main

import (
	"os"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	// Test case 1: from -f flag
	t.Run("from -f flag", func(t *testing.T) {
		path, err := getConfigPath([]string{"-f", "myconfig.yml"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if path != "myconfig.yml" {
			t.Errorf("expected path 'myconfig.yml', got '%s'", path)
		}
	})

	// Test case 2: from arg
	t.Run("from arg", func(t *testing.T) {
		f, err := os.Create("myconfig.yml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("myconfig.yml")
		f.Close()

		path, err := getConfigPath([]string{"myconfig.yml"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if path != "myconfig.yml" {
			t.Errorf("expected path 'myconfig.yml', got '%s'", path)
		}
	})

	// Test case 3: from env
	t.Run("from env", func(t *testing.T) {
		os.Setenv("ILC_FILE", "envconfig.yml")
		defer os.Unsetenv("ILC_FILE")
		path, err := getConfigPath([]string{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if path != "envconfig.yml" {
			t.Errorf("expected path 'envconfig.yml', got '%s'", path)
		}
	})

	// Test case 4: default files
	t.Run("default files", func(t *testing.T) {
		f, err := os.Create("ilc.yml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove("ilc.yml")
		f.Close()

		path, err := getConfigPath([]string{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if path != "ilc.yml" {
			t.Errorf("expected path 'ilc.yml', got '%s'", path)
		}
	})

	// Test case 5: no file
	t.Run("no file", func(t *testing.T) {
		_, err := getConfigPath([]string{})
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestLoadConfig(t *testing.T) {
	f, err := os.Create("test.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test.yml")

	f.WriteString(`
description: Test config
`)
	f.Close()

	cfg, err := loadConfig([]string{"-f", "test.yml"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Description != "Test config" {
		t.Errorf("expected description 'Test config', got '%s'", cfg.Description)
	}
}
