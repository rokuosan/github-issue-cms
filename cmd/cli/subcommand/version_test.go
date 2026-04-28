package subcommand

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionCommand(t *testing.T) {
	version := "v1.2.3"
	cmd := NewVersionCommand(&version)

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Contains(t, cmd.Short, "version")
}

func TestVersionCommand_PrintsVersion(t *testing.T) {
	version := "v1.2.3"
	cmd := NewVersionCommand(&version)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "v1.2.3\n", stdout.String())
}

func TestVersionCommand_WithoutVersion(t *testing.T) {
	cmd := NewVersionCommand(nil)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "version is not set")
}
