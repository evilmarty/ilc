package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

const (
	DefaultHistoryFile      = ".ilc_history"
	DefaultHistoryFilePerms = 0644
)

var ErrInvalidHistoryFile = errors.New("invalid history file")

type History struct {
	Records map[string][][]string
	Path    string
}

func (h *History) Lookup(filepath string, args []string) ([]string, bool) {
	entries, found := h.Records[filepath]
	if !found {
		return []string{}, false
	}
	argsLen := len(args)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		entryLen := len(entry)
		// Nothing to match so the latest record should be returned
		// DeepEqual will match this case too but just saving some cycles
		if entryLen == 0 {
			return entry, true
		}
		// Skip if the entry is shorter than the args as it won't match
		if entryLen < argsLen {
			continue
		}
		subentry := entry[0:argsLen]
		if reflect.DeepEqual(subentry, args) {
			return entry, true
		}
	}
	return []string{}, false
}

func (h *History) Append(filepath string, args []string) {
	logger.Printf("Adding history entry for config %s\n", filepath)
	entries := h.Records[filepath]
	if !h.isLatestRecord(filepath, args) {
		h.Records[filepath] = append(entries, args)
	}
}

func (h *History) isLatestRecord(filepath string, args []string) bool {
	entries := h.Records[filepath]
	if entriesLen := len(entries); entriesLen > 0 {
		return reflect.DeepEqual(entries[entriesLen-1], args)
	}
	return false
}

func (h *History) Save() error {
	logger.Printf("Saving history to %s\n", h.Path)
	content, err := yaml.Marshal(h.Records)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(h.Path, os.O_CREATE|os.O_WRONLY, DefaultHistoryFilePerms)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(content)
	return nil
}

func (h *History) load() error {
	logger.Printf("Loading history from %s\n", h.Path)
	if content, err := os.ReadFile(h.Path); err != nil {
		return err
	} else if err := yaml.Unmarshal(content, &h.Records); err != nil {
		return err
	} else {
		return nil
	}
}

func LoadHistory(path string) (History, error) {
	if path == "" {
		if userHome, err := os.UserHomeDir(); err != nil {
			path = DefaultHistoryFile
		} else {
			path = filepath.Join(userHome, DefaultHistoryFile)
		}
	}
	history := History{
		Path:    path,
		Records: make(map[string][][]string),
	}
	err := history.load()
	return history, err
}
