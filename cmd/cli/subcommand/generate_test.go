package subcommand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGenerateCommand(t *testing.T) {
	cmd := NewGenerateCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "generate", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate articles")

	// Verify the token flag.
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag)
	assert.Equal(t, "t", tokenFlag.Shorthand)

	// Ensure the flag is marked as required.
	assert.Contains(t, cmd.Flags().Lookup("token").Annotations, "cobra_annotation_bash_completion_one_required_flag")
}

func TestGenerateCommand_Flags(t *testing.T) {
	cmd := NewGenerateCommand()

	// Test the token flag.
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag, "token flag should exist")
	assert.Equal(t, "t", tokenFlag.Shorthand, "token shorthand should be 't'")
}

func TestGenerateCommand_Help(t *testing.T) {
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGenerateCommand_MissingToken(t *testing.T) {
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{}) // No token provided.

	err := cmd.Execute()
	assert.Error(t, err, "Should error when token is missing")
}

func TestGenerateCommand_WithToken(t *testing.T) {
	// Skip because this requires an integration test.
	t.Skip("Integration test required - needs valid config file")
}

func TestGenerateCommand_Examples(t *testing.T) {
	cmd := NewGenerateCommand()

	// Ensure the examples are present.
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Examples:")
}
