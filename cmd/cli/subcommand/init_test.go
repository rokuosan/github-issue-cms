package subcommand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "init", cmd.Use)
	assert.Contains(t, cmd.Short, "Generate config")

	// Verify the flags.
	usernameFlag := cmd.Flags().Lookup("username")
	assert.NotNil(t, usernameFlag)
	assert.Equal(t, "u", usernameFlag.Shorthand)

	repositoryFlag := cmd.Flags().Lookup("repository")
	assert.NotNil(t, repositoryFlag)
	assert.Equal(t, "r", repositoryFlag.Shorthand)
}

func TestInitCommand_Flags(t *testing.T) {
	cmd := NewInitCommand()

	// Username flag.
	usernameFlag := cmd.Flags().Lookup("username")
	assert.NotNil(t, usernameFlag)
	assert.Equal(t, "u", usernameFlag.Shorthand)

	// Repository flag.
	repositoryFlag := cmd.Flags().Lookup("repository")
	assert.NotNil(t, repositoryFlag)
	assert.Equal(t, "r", repositoryFlag.Shorthand)
}

func TestInitCommand_Help(t *testing.T) {
	cmd := NewInitCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestInitCommand_CreateConfigFile(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir := t.TempDir()

	// Change the working directory to the temporary directory.
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWd))
	})

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Run the init command.
	cmd := NewInitCommand()
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Ensure the config file was created.
	configPath := filepath.Join(tempDir, config.ConfigFileName+"."+config.ConfigFileType)
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Config file should be created")
}

func TestInitCommand_WithUsernameAndRepository(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir := t.TempDir()

	// Change the working directory to the temporary directory.
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWd))
	})

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Run the init command with username and repository.
	cmd := NewInitCommand()
	cmd.SetArgs([]string{
		"--username", "testuser",
		"--repository", "testrepo",
	})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Ensure the config file was created.
	configPath := filepath.Join(tempDir, config.ConfigFileName+"."+config.ConfigFileType)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify the config contents by reading the file.
	data, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "testuser")
	assert.Contains(t, content, "testrepo")
}

func TestInitCommand_ShortFlags(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWd))
	})

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Run the command with shorthand flags.
	cmd := NewInitCommand()
	cmd.SetArgs([]string{
		"-u", "shortuser",
		"-r", "shortrepo",
	})

	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify the config file contents.
	configPath := filepath.Join(tempDir, config.ConfigFileName+"."+config.ConfigFileType)
	data, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "shortuser")
	assert.Contains(t, content, "shortrepo")
}

func TestInitCommand_Examples(t *testing.T) {
	cmd := NewInitCommand()

	// Ensure the examples are present.
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, "Examples:")
}
