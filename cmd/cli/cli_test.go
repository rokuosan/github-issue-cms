package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "github-issue-cms", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate articles from GitHub issues")

	// Ensure subcommands are registered.
	commands := cmd.Commands()
	assert.GreaterOrEqual(t, len(commands), 2, "Should have at least 2 subcommands")

	var hasGenerate, hasInit bool
	for _, subCmd := range commands {
		switch subCmd.Use {
		case "generate":
			hasGenerate = true
		case "init":
			hasInit = true
		}
	}

	assert.True(t, hasGenerate, "Should have 'generate' subcommand")
	assert.True(t, hasInit, "Should have 'init' subcommand")
}

func TestRootCommand_Flags(t *testing.T) {
	cmd := NewRootCommand()

	// Verify the debug flag.
	debugFlag := cmd.PersistentFlags().Lookup("debug")
	assert.NotNil(t, debugFlag)
	assert.Equal(t, "d", debugFlag.Shorthand)
	assert.Equal(t, "false", debugFlag.DefValue)
}

func TestRootCommand_Help(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCommand_Version(t *testing.T) {
	cmd := NewRootCommand()

	// Version information is not available yet, but may be added later.
	// For now, just ensure the command is constructed correctly.
	assert.NotNil(t, cmd)
}

func TestRootCommand_InvalidSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"invalid-command"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRootCommand_DebugFlag(t *testing.T) {
	cmd := NewRootCommand()

	// Ensure the debug flag can be set.
	cmd.SetArgs([]string{"--debug", "--help"})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Ensure the debug flag is parsed.
	debugFlag := cmd.PersistentFlags().Lookup("debug")
	assert.NotNil(t, debugFlag)
}
