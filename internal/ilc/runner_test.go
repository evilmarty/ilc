package ilc

import (
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
