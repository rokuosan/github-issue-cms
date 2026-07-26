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

	// Verify the --with-ogimage flag exists.
	ogimageFlag := cmd.Flags().Lookup("with-ogimage")
	assert.NotNil(t, ogimageFlag, "--with-ogimage flag should exist")

	// Ensure the token flag is marked as required.
	assert.Contains(t, cmd.Flags().Lookup("token").Annotations, "cobra_annotation_bash_completion_one_required_flag")
}

func TestGenerateCommand_Flags(t *testing.T) {
	cmd := NewGenerateCommand()

	// Test the token flag.
	tokenFlag := cmd.Flags().Lookup("token")
	assert.NotNil(t, tokenFlag, "token flag should exist")
	assert.Equal(t, "t", tokenFlag.Shorthand, "token shorthand should be 't'")

	// Test the --with-ogimage flag.
	ogimageFlag := cmd.Flags().Lookup("with-ogimage")
	assert.NotNil(t, ogimageFlag, "--with-ogimage flag should exist")
}

func TestGenerateCommand_WithOGImageFlag(t *testing.T) {
	cmd := NewGenerateCommand()
	cmd.SetArgs([]string{"--token", "test-token", "--with-ogimage"})
	// This will fail because no config exists, but verifies the flag is parsed.
	err := cmd.Execute()
	assert.Error(t, err) // Missing config is expected.
}

func TestGenerateCommand_Examples(t *testing.T) {
	cmd := NewGenerateCommand()

	// Ensure the examples are present and mention --with-ogimage.
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Examples:")
	assert.Contains(t, cmd.Long, "--with-ogimage")
}
