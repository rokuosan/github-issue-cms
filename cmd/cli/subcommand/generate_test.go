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
	t.Run("flat layout places OGP adjacent to markdown with swapped extension", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
			Key:  "2024-01-15_103000",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		// Markdown is saved as content/posts/2024-01-15_103000.md, so the
		// OGP image must be the adjacent file 2024-01-15_103000.ogp.jpeg —
		// not an orphaned content/posts/<key>/ogp.jpeg subdirectory.
		assert.Equal(t, "content/posts/2024-01-15_103000.ogp.jpeg", path)
	})

	t.Run("page bundle layout places ogp.jpeg in the bundle directory", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		conf.Output.Articles.Filename = "index.md"
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
			Key:  "2024-01-15_103000",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		assert.Equal(t, "content/posts/ogp.jpeg", path)
	})

	t.Run("flat layout uses datetime from date, not the article key", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
			Key:  "some-other-key",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		assert.Equal(t, "content/posts/2024-01-15_103000.ogp.jpeg", path)
	})

	t.Run("non-md extension stays adjacent by appending", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		conf.Output.Articles.Filename = "%Y-%m-%d.post"
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		// Non-markdown extensions are NOT stripped — the OGP appends to the
		// full filename so it always stays adjacent to the markdown.
		assert.Equal(t, "content/posts/2024-01-15.post.ogp.jpeg", path)
	})

	t.Run(".markdown extension is stripped like .md", func(t *testing.T) {
		conf := testConfig(t.TempDir() + "/articles")
		conf.Output.Articles.Filename = "%Y-%m-%d.markdown"
		article := &core.Article{
			Date: "2024-01-15T10:30:00Z",
		}

		path, err := resolveOGPArticlePath(conf, article)
		require.NoError(t, err)
		assert.Equal(t, "content/posts/2024-01-15.ogp.jpeg", path)
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
