package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestWriteAndReload_PreservesExplicitEmptyImageTargets(t *testing.T) {
	tempDir := t.TempDir()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
		config = Config{}
		viper.Reset()
	})

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	config = Config{}
	viper.Reset()

	conf := Config{
		GitHub: NewGitHubConfig(),
		Output: &OutputConfig{
			Articles: NewOutputArticlesConfig(),
			Images: &OutputImagesConfig{
				Directory: "static/images",
				Filename:  "[:id].png",
				BaseURL:   Ptr("/images"),
				Targets:   []string{},
			},
		},
	}

	if err := Write(conf); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(GetConfigPath())
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "targets: []") {
		t.Fatalf("expected explicit empty targets in config, got:\n%s", string(data))
	}

	reloaded, err := Reload()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Output == nil || reloaded.Output.Images == nil {
		t.Fatalf("missing output images after reload")
	}
	if reloaded.Output.Images.Targets == nil {
		t.Fatalf("targets became nil after reload")
	}
	if len(reloaded.Output.Images.TargetURLs()) != 0 {
		t.Fatalf("target urls = %#v", reloaded.Output.Images.TargetURLs())
	}
}
