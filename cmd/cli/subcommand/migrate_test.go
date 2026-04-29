package subcommand

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMigrateCommand(t *testing.T) {
	cmd := NewMigrateCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "migrate", cmd.Use)
	assert.Contains(t, cmd.Short, "latest schema")
}

func TestMigrateCommand_Help(t *testing.T) {
	cmd := NewMigrateCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestMigrateCommand_RewritesLegacyHugoConfig(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalWd))
	})

	require.NoError(t, os.Chdir(tempDir))

	legacy := `github:
  username: "testuser"
  repository: "testrepo"
hugo:
  directory:
    articles: "content/posts"
    images: "static/images/%Y-%m-%d"
  filename:
    articles: "index.md"
    images: "[:id].png"
  url:
    images: "/images/%Y-%m-%d"
`
	configPath := filepath.Join(tempDir, config.ConfigFileName+"."+config.ConfigFileType)
	require.NoError(t, os.WriteFile(configPath, []byte(legacy), 0o644))

	cmd := NewMigrateCommand()
	err = cmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "output:")
	assert.Contains(t, content, "articles:")
	assert.Contains(t, content, "directory: content/posts")
	assert.Contains(t, content, "filename: index.md")
	assert.Contains(t, content, "url: /images/%Y-%m-%d")
	assert.NotContains(t, content, "\nhugo:")
}
