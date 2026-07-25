package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestOutputImagesConfig_TrustedImageHosts(t *testing.T) {
	images := NewOutputImagesConfig()
	expected := []string{
		"github.com",
		"user-images.githubusercontent.com",
		"private-user-images.githubusercontent.com",
	}
	got := images.TrustedImageHosts()
	if len(got) != len(expected) {
		t.Fatalf("trusted hosts = %#v, want %#v", got, expected)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("trusted hosts = %#v, want %#v", got, expected)
		}
	}
}

func TestOutputImagesConfig_TrustedImageHosts_PreservesExplicitEmptyValue(t *testing.T) {
	images := &OutputImagesConfig{TrustedHosts: []string{}}
	if got := images.TrustedImageHosts(); len(got) != 0 {
		t.Fatalf("trusted hosts = %#v, want empty", got)
	}
}

func TestConfigValidate_RejectsInvalidTrustedImageHosts(t *testing.T) {
	conf := NewConfig()
	conf.Output.Images.TrustedHosts = []string{"https://github.com"}
	if err := conf.validate(); err == nil {
		t.Fatal("expected trusted host validation to fail")
	}
}

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
				Directory:    "static/images",
				Filename:     "[:id].png",
				BaseURL:      Ptr("/images"),
				Targets:      []string{},
				TrustedHosts: []string{},
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
	if !strings.Contains(string(data), "trusted_hosts: []") {
		t.Fatalf("expected explicit empty trusted hosts in config, got:\n%s", string(data))
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
	if reloaded.Output.Images.TrustedHosts == nil {
		t.Fatal("trusted hosts became nil after reload")
	}
	if len(reloaded.Output.Images.TrustedImageHosts()) != 0 {
		t.Fatalf("trusted hosts = %#v", reloaded.Output.Images.TrustedImageHosts())
	}
}
