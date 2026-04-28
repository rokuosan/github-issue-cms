package cli

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "github-issue-cms", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate articles from GitHub issues")

	// Ensure subcommands are registered.
	commands := cmd.Commands()
	assert.GreaterOrEqual(t, len(commands), 3, "Should have at least 3 subcommands")

	var hasGenerate, hasInit, hasVersion bool
	for _, subCmd := range commands {
		switch subCmd.Use {
		case "generate":
			hasGenerate = true
		case "init":
			hasInit = true
		case "version":
			hasVersion = true
		}
	}

	assert.True(t, hasGenerate, "Should have 'generate' subcommand")
	assert.True(t, hasInit, "Should have 'init' subcommand")
	assert.True(t, hasVersion, "Should have 'version' subcommand")
}

func TestRootCommand_Flags(t *testing.T) {
	cmd := NewRootCommand()

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Equal(t, "0", verboseFlag.DefValue)
}

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCommand_Version(t *testing.T) {
	cmd := NewRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"version"})

	originalVersion := Version
	Version = "v9.9.9"
	t.Cleanup(func() {
		Version = originalVersion
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "v9.9.9\n", stdout.String())
}

func TestRootCommand_InvalidSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"invalid-command"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestLogLevelForVerbosity(t *testing.T) {
	tests := []struct {
		name      string
		verbosity int
		want      slog.Level
	}{
		{name: "default", verbosity: 0, want: slog.LevelError},
		{name: "verbose", verbosity: 1, want: slog.LevelInfo},
		{name: "very verbose", verbosity: 2, want: slog.LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, logLevelForVerbosity(tt.verbosity))
		})
	}
}
