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
		parsed:         true,
		HistoryStore:   mockStore,
		ConfigPath:     "config.yml",
		Config:         &Config{},
		Args:           []string{"!hello"},
		NonInteractive: true,
	}

	// Verify isReplay is true
	assert.True(t, r.isReplay())

	// Run replay via Run()
	err := r.Run()
	// Since NonInteractive is true, it immediately aborts on the non-runnable blank config and returns ErrInvalidCommand.
	assert.ErrorIs(t, err, ErrInvalidCommand)
	assert.Equal(t, []string{"hello", "world"}, r.Args)
}

func TestRunner_ParsedAndFlags(t *testing.T) {
	var buf bytes.Buffer
	r := Runner{
		Output: &buf,
	}

	assert.False(t, r.Parsed())

	// Test Printf
	r.Printf("Hello %d", 42)
	assert.Equal(t, "Hello 42", buf.String())
}

func TestRunner_ParseErrors(t *testing.T) {
	r := &Runner{}
	// Empty args
	err := r.Parse([]string{"ilc"})
	assert.ErrorIs(t, err, ErrConfigFileMissing)

	// Missing arguments on Run
	r.parsed = false
	err = r.Run()
	assert.ErrorIs(t, err, ErrMissingArguments)
}

func TestRunner_RunNonInteractiveMissingInputs(t *testing.T) {
	content := `
description: Test config
inputs:
  name:
    description: Name is required
commands:
  greet:
    run: echo "Hello"
`
	tempFile, err := os.CreateTemp("", "ilc-test-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err)
	tempFile.Close()

	r := Runner{
		Name:           "ILC",
		NonInteractive: true,
		Args:           []string{"ilc", "-non-interactive", tempFile.Name(), "greet"},
	}

	err = r.Parse(r.Args)
	assert.NoError(t, err)

	err = r.Run()
	// Should fail since "name" input has no default and is missing
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing inputs: name")
}

func TestRunner_ReplayCommandNotFound(t *testing.T) {
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

	r := Runner{
		Name:         "ILC",
		HistoryStore: mockStore,
		ConfigPath:   "config.yml",
		Args:         []string{"!notfound"},
	}

	err := r.replay()
	assert.ErrorIs(t, err, ErrInvalidReplay)
}

func TestRunner_RecordToHistory(t *testing.T) {
	records := make(map[string][][]string)
	history := &History{
		Path:    "dummy_path",
		Records: records,
	}
	mockStore := &MockHistoryStore{
		History: history,
	}

	r := Runner{
		HistoryStore: mockStore,
		ConfigPath:   "config.yml",
	}

	r.recordToHistory([]string{"test-arg"})
	assert.Len(t, mockStore.Saved, 1)
	assert.Contains(t, mockStore.Saved[0].Records["config.yml"][0], "test-arg")
}

func TestRunner_RunNonInteractiveSuccess(t *testing.T) {
	content := `
description: Test config
commands:
  greet:
    run: echo "Hello Successful Run"
`
	tempFile, err := os.CreateTemp("", "ilc-test-*.yml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err)
	tempFile.Close()

	records := make(map[string][][]string)
	history := &History{
		Path:    "dummy_path",
		Records: records,
	}
	mockStore := &MockHistoryStore{
		History: history,
	}

	var outBuf bytes.Buffer
	r := Runner{
		Name:           "ILC",
		NonInteractive: true,
		Stdout:         &outBuf,
		Stderr:         &outBuf,
		HistoryStore:   mockStore,
		Args:           []string{"ilc", "-non-interactive", tempFile.Name(), "greet"},
	}

	err = r.Parse(r.Args)
	assert.NoError(t, err)

	err = r.Run()
	assert.NoError(t, err)
	assert.Contains(t, outBuf.String(), "Hello Successful Run")
	// Verify history record is added
	assert.Len(t, mockStore.Saved, 1)
}

func TestRunner_RecordToHistory_SaveError(t *testing.T) {
	mockStore := &MockHistoryStore{
		History: &History{Records: make(map[string][][]string)},
		SaveErr: assert.AnError,
	}
	r := Runner{
		HistoryStore: mockStore,
		ConfigPath:   "config.yml",
	}

	// This should not panic or return error, but log it
	r.recordToHistory([]string{"test-arg"})
	assert.Nil(t, mockStore.Saved)
}

func TestRunner_RunHistoryFileFromEnv(t *testing.T) {
	r := Runner{
		parsed:         true,
		NonInteractive: true,
		Config:         &Config{},
		Env:            EnvMap{EnvHistFile: "my_env_hist_file"},
	}
	// We want it to fail because config path is missing, but check if HistoryFile got populated
	err := r.Run()
	assert.Error(t, err)
	assert.Equal(t, "my_env_hist_file", r.HistoryFile)
}

func TestRunner_ReplayLoadError(t *testing.T) {
	mockStore := &MockHistoryStore{
		LoadErr: assert.AnError,
	}
	r := Runner{
		HistoryStore: mockStore,
		Args:         []string{"!hello"},
	}
	err := r.replay()
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRunner_PrintVersion(t *testing.T) {
	oldExitFunc := exitFunc
	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}
	defer func() { exitFunc = oldExitFunc }()

	var buf bytes.Buffer
	r := Runner{
		Name:      "TestApp",
		Version:   "1.2.3",
		BuildDate: "2026-05-24",
		Commit:    "abcdef",
		Output:    &buf,
	}

	r.printVersion()
	assert.Equal(t, 0, exitCode)
	assert.Contains(t, buf.String(), "TestApp")
	assert.Contains(t, buf.String(), "Version: 1.2.3")
}

func TestRunner_UsageFlagSet(t *testing.T) {
	oldExitFunc := exitFunc
	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}
	defer func() { exitFunc = oldExitFunc }()

	var buf bytes.Buffer
	r := Runner{
		Name:   "TestApp",
		Output: &buf,
	}

	fs := r.flagSet()
	fs.Usage()
	assert.Equal(t, 0, exitCode)
}




