package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataParser_Parse(t *testing.T) {
	parser := newMetadataParser()

	tests := []struct {
		name       string
		body       string
		wantRaw    string
		wantParsed map[string]any
		wantErr    string
	}{
		{
			name:       "yaml code fence",
			body:       "```yaml\nauthor: Test User\ncustom: value\n```\n\nBody",
			wantRaw:    "```yaml\nauthor: Test User\ncustom: value\n```",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:       "yaml front matter",
			body:       "---\nauthor: Test User\ncustom: value\n---\n\nBody",
			wantRaw:    "---\nauthor: Test User\ncustom: value\n---",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:       "toml front matter",
			body:       "+++\nauthor = \"Test User\"\ncustom = \"value\"\n+++\n\nBody",
			wantRaw:    "+++\nauthor = \"Test User\"\ncustom = \"value\"\n+++",
			wantParsed: map[string]any{"author": "Test User", "custom": "value"},
		},
		{
			name:    "invalid yaml code fence",
			body:    "```\nauthor: [broken\n```\n\nBody",
			wantErr: "failed to parse front matter",
		},
		{
			name:    "no metadata",
			body:    "Body only",
			wantErr: "front matter not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parser.Parse(tt.body)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assertEqualCmp(t, tt.wantRaw, block.Raw)
			assertEqualCmp(t, tt.wantParsed, block.Values.Values())
		})
	}
}
