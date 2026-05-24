package ilc

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerIsReplay(t *testing.T) {
	runner := Runner{}
	runner.Args = []string{"arg1", "arg2"}
	assert.False(t, runner.isReplay())
	runner.Args = []string{"!arg1", "arg2"}
	assert.True(t, runner.isReplay())
	runner.Args = []string{"!", "arg1", "arg2"}
	assert.True(t, runner.isReplay())
}

func TestRunner_ValidateOutput(t *testing.T) {
	// 1. Create a temporary valid configuration file content
	content := `
description: Test validation configuration
commands:
  hello:
    run: echo "Hello World"
`
	tempFile, err := os.CreateTemp("", "ilc-test-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err)
	tempFile.Close()

	// 2. Initialize Runner with a custom buffer for Output
	var buf bytes.Buffer
	r := Runner{
		Name:           "ILC",
		ValidateConfig: true,
		Output:         &buf,
	}

	// 3. Parse arguments
	args := []string{"ilc", "-validate", tempFile.Name()}
	err = r.Parse(args)
	assert.NoError(t, err)

	// 4. Run the validation
	err = r.Run()
	assert.NoError(t, err)

	// 5. Assert the custom Output buffer captures the correct validation logs
	assert.Contains(t, buf.String(), "configuration is valid")
}

type MockHistoryStore struct {
	History *History
	LoadErr error
	SaveErr error
	Saved   []*History
}

func (m *MockHistoryStore) Load(path string) (*History, error) {
	if m.LoadErr != nil {
		return nil, m.LoadErr
	}
	return m.History, nil
}

func (m *MockHistoryStore) Save(h *History) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Saved = append(m.Saved, h)
	return nil
}

func TestRunner_ReplayMocked(t *testing.T) {
	// Create mock history records
	records := make(map[string][][]string)
	records["config.yml"] = [][]string{
		{"hello", "world"},
	}

	history := &History{
		Path:    "dummy_path",
		Records: records,
	}

	mockStore := &MockHistoryStore{
		History: history,
	}

	// Initialize runner with mock store
	r := Runner{
		Name:           "ILC",
		HistoryStore:   mockStore,
		ConfigPath:     "config.yml",
		Config:         &Config{},
		Args:           []string{"!hello"},
		NonInteractive: true,
	}

	// Verify isReplay is true
	assert.True(t, r.isReplay())

	// Run replay
	err := r.replay()
	// Since NonInteractive is true, it immediately aborts on the non-runnable blank config and returns ErrInvalidCommand.
	assert.ErrorIs(t, err, ErrInvalidCommand)
	assert.Equal(t, []string{"hello", "world"}, r.Args)
}
