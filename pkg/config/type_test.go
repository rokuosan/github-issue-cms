package config

import "testing"

func TestConfigNormalize_LegacyHugoDirectoryFields(t *testing.T) {
	conf := Config{
		GitHub: NewGitHubConfig(),
		Hugo: &HugoConfig{
			Directory: &HugoDirectoryConfig{
				Articles: "content/posts",
				Images:   "static/images/%Y-%m-%d",
			},
			Filename: &HugoFilenameConfig{
				Articles: "index.md",
				Images:   "[:id].png",
			},
			Url: &HugoURLConfig{
				Images: "/images/%Y-%m-%d",
			},
		},
	}

	conf.normalize()

	if conf.Output.Articles.Directory != "content/posts" {
		t.Fatalf("articles directory = %q", conf.Output.Articles.Directory)
	}
	if conf.Output.Articles.Filename != "index.md" {
		t.Fatalf("articles filename = %q", conf.Output.Articles.Filename)
	}
	if conf.Output.Images.Directory != "static/images/%Y-%m-%d" {
		t.Fatalf("images directory = %q", conf.Output.Images.Directory)
	}
	if conf.Output.Images.Filename != "[:id].png" {
		t.Fatalf("images filename = %q", conf.Output.Images.Filename)
	}
	if conf.Output.Images.URL != "/images/%Y-%m-%d" {
		t.Fatalf("images url = %q", conf.Output.Images.URL)
	}
}

func TestConfigNormalize_PrefersOutputOverLegacyHugo(t *testing.T) {
	conf := Config{
		GitHub: NewGitHubConfig(),
		Output: &OutputConfig{
			Articles: &OutputArticlesConfig{
				Directory: "content/custom",
				Filename:  "custom.md",
			},
			Images: &OutputImagesConfig{
				Directory: "static/custom",
				Filename:  "custom.png",
				URL:       "/custom",
			},
		},
		Hugo: &HugoConfig{
			Directory: &HugoDirectoryConfig{
				Articles: "content/posts",
				Images:   "static/images",
			},
			Filename: &HugoFilenameConfig{
				Articles: "index.md",
				Images:   "[:id].png",
			},
			Url: &HugoURLConfig{
				Images: "/images",
			},
		},
	}

	conf.normalize()

	if conf.Output.Articles.Directory != "content/custom" {
		t.Fatalf("articles directory = %q", conf.Output.Articles.Directory)
	}
	if conf.Output.Articles.Filename != "custom.md" {
		t.Fatalf("articles filename = %q", conf.Output.Articles.Filename)
	}
	if conf.Output.Images.Directory != "static/custom" {
		t.Fatalf("images directory = %q", conf.Output.Images.Directory)
	}
	if conf.Output.Images.Filename != "custom.png" {
		t.Fatalf("images filename = %q", conf.Output.Images.Filename)
	}
	if conf.Output.Images.URL != "/custom" {
		t.Fatalf("images url = %q", conf.Output.Images.URL)
	}
}
