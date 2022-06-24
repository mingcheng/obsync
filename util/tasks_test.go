package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTasksByPath(t *testing.T) {
	// has := prefixPath("..")
	// assert.True(t, has)

	tasks, err := TasksByPath("..")
	assert.NoError(t, err)
	assert.NotEmpty(t, tasks)

	assert.Equal(t, len(tasks) > 0, true)
}
