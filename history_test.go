package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoryLookup(t *testing.T) {
	history := History{
		Records: map[string][][]string{
			"some file1": {
				[]string{"arg1", "arg2", "arg3"},
				[]string{"arg1"},
				[]string{"arg1", "arg2 arg3"},
			},
			"some file2": {
				[]string{"arg1", "arg2"},
			},
		},
	}
	t.Run("when has no config record", func(t *testing.T) {
		expected := []string{}
		actual, found := history.Lookup("some file3", []string{"arg1"})
		assert.False(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("when query args do not match", func(t *testing.T) {
		expected := []string{}
		actual, found := history.Lookup("some file1", []string{"arg1", "arg2", "arg3", "arg4"})
		assert.False(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("when query args do match all", func(t *testing.T) {
		expected := []string{"arg1", "arg2", "arg3"}
		actual, found := history.Lookup("some file1", []string{"arg1", "arg2", "arg3"})
		assert.True(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("when query args do match some", func(t *testing.T) {
		expected := []string{"arg1", "arg2", "arg3"}
		actual, found := history.Lookup("some file1", []string{"arg1", "arg2"})
		assert.True(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("when query args are empty", func(t *testing.T) {
		expected := []string{"arg1", "arg2 arg3"}
		actual, found := history.Lookup("some file1", []string{})
		assert.True(t, found)
		assert.Equal(t, expected, actual)
	})
}

func TestHistoryAppend(t *testing.T) {
	actual := History{
		Records: map[string][][]string{
			"some file1": {
				[]string{"arg1", "arg2", "arg3"},
				[]string{"arg1"},
			},
			"some file2": {
				[]string{"arg1", "arg2"},
			},
		},
	}
	expected := History{
		Records: map[string][][]string{
			"some file1": {
				[]string{"arg1", "arg2", "arg3"},
				[]string{"arg1"},
			},
			"some file2": {
				[]string{"arg1", "arg2"},
			},
			"some file3": {
				[]string{"arg1", "arg2"},
			},
		},
	}
	actual.Append("some file3", []string{"arg1", "arg2"})
	assert.Equal(t, expected, actual)
	actual.Append("some file3", []string{"arg1", "arg2"})
	assert.Equal(t, expected, actual)
}

func TestHistorySave(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "")
	assert.NoError(t, err, "Failed to create temp file")

	history := History{
		Path: tempFile.Name(),
		Records: map[string][][]string{
			"some file1": {
				[]string{"arg1", "arg2", "arg3"},
				[]string{"arg1"},
				[]string{"arg1", "arg2", "arg3"},
			},
			"some file2": {
				[]string{"arg1", "arg2"},
			},
		},
	}
	expected := `some file1:
    - - arg1
      - arg2
      - arg3
    - - arg1
    - - arg1
      - arg2
      - arg3
some file2:
    - - arg1
      - arg2
`
	err = history.Save()
	assert.NoError(t, err)
	actual, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}

func TestLoadHistory(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		tempFile, err := os.CreateTemp(t.TempDir(), "")
		assert.NoError(t, err, "Failed to create temp file")

		expected := History{
			Path:    tempFile.Name(),
			Records: make(map[string][][]string),
		}
		actual, err := LoadHistory(tempFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("invalid file", func(t *testing.T) {
		content := `
  arg1 arg2 arg3
some file:
  arg1 arg2
`
		tempFile, err := os.CreateTemp(t.TempDir(), "")
		assert.NoError(t, err, "Failed to create temp file")

		_, err = tempFile.Write([]byte(content))
		assert.NoError(t, err, "Failed to write config to temp file")

		expected := History{
			Path:    tempFile.Name(),
			Records: make(map[string][][]string),
		}
		actual, err := LoadHistory(tempFile.Name())
		assert.Error(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("valid file", func(t *testing.T) {
		content := `
# comment
some file1:
  - [arg1, arg2, arg3]
  - [arg1]

some file2:
  - [arg1, arg2]
`
		tempFile, err := os.CreateTemp(t.TempDir(), "")
		assert.NoError(t, err, "Failed to create temp file")

		_, err = tempFile.Write([]byte(content))
		assert.NoError(t, err, "Failed to write config to temp file")

		expected := History{
			Path: tempFile.Name(),
			Records: map[string][][]string{
				"some file1": {
					[]string{"arg1", "arg2", "arg3"},
					[]string{"arg1"},
				},
				"some file2": {
					[]string{"arg1", "arg2"},
				},
			},
		}
		actual, err := LoadHistory(tempFile.Name())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
