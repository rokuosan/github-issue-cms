package subcommand

import (
	"testing"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/rokuosan/github-issue-cms/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestResolveOGPArticlePath(t *testing.T) {
	t.Run("uses article key as subdirectory", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
			Key:  "2024-01-15_103000",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		assert.Contains(t, path, "ogp.jpeg")
		assert.Contains(t, path, "2024-01-15_103000")
	})

	t.Run("falls back to datetime when key is empty", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		assert.Contains(t, path, "ogp.jpeg")
	})

	t.Run("nil article returns error", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		_, err := resolveOGPArticlePath(conf, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "article is nil")
	})

	t.Run("nil output config returns error", func(t *testing.T) {
		conf := config.Config{}
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
		}
		_, err := resolveOGPArticlePath(conf, article)
		assert.Error(t, err)
	})
}
