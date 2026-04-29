package config

// Config package is a package for configuration.
// If you change the configuration, you also need to change ``config.Generate()`` (config.go).

type Config struct {
	GitHub *GitHubConfig `yaml:"github" mapstructure:"github"`
	Output *OutputConfig `yaml:"output" mapstructure:"output"`
	Hugo   *HugoConfig   `yaml:"hugo,omitempty" mapstructure:"hugo"`
}

type GitHubConfig struct {
	Username   string `yaml:"username" mapstructure:"username"`
	Repository string `yaml:"repository" mapstructure:"repository"`
}

type OutputConfig struct {
	Articles *OutputArticlesConfig `yaml:"articles" mapstructure:"articles"`
	Images   *OutputImagesConfig   `yaml:"images" mapstructure:"images"`
}

type OutputArticlesConfig struct {
	Directory string `yaml:"directory" mapstructure:"directory"`
	Filename  string `yaml:"filename" mapstructure:"filename"`
}

type OutputImagesConfig struct {
	Directory string  `yaml:"directory" mapstructure:"directory"`
	Filename  string  `yaml:"filename" mapstructure:"filename"`
	BaseURL   *string `yaml:"url" mapstructure:"url"`
}

type HugoConfig struct {
	Content   *HugoContentConfig   `yaml:"content,omitempty" mapstructure:"content"`
	Images    *HugoImagesConfig    `yaml:"images,omitempty" mapstructure:"images"`
	Bundle    string               `yaml:"bundle,omitempty" mapstructure:"bundle"`
	Directory *HugoDirectoryConfig `yaml:"directory,omitempty" mapstructure:"directory"`
	Filename  *HugoFilenameConfig  `yaml:"filename,omitempty" mapstructure:"filename"`
	Url       *HugoURLConfig       `yaml:"url,omitempty" mapstructure:"url"`
}

type HugoContentConfig struct {
	Directory string `yaml:"directory" mapstructure:"directory"`
	Filename  string `yaml:"filename" mapstructure:"filename"`
}

type HugoImagesConfig struct {
	Directory string `yaml:"directory" mapstructure:"directory"`
	Filename  string `yaml:"filename" mapstructure:"filename"`
	URL       string `yaml:"url" mapstructure:"url"`
}

type HugoDirectoryConfig struct {
	Articles string `yaml:"articles" mapstructure:"articles"`
	Images   string `yaml:"images" mapstructure:"images"`
}

type HugoFilenameConfig struct {
	Articles string `yaml:"articles" mapstructure:"articles"`
	Images   string `yaml:"images" mapstructure:"images"`
}

type HugoURLConfig struct {
	AppendSlash bool   `yaml:"appendSlash,omitempty" mapstructure:"appendSlash"`
	Images      string `yaml:"images" mapstructure:"images"`
}

func (c *GitHubConfig) RepositoryURL() string {
	return "https://github.com/" + c.Username + "/" + c.Repository
}

func NewConfig() *Config {
	return &Config{
		GitHub: NewGitHubConfig(),
		Output: NewOutputConfig(),
	}
}

func NewGitHubConfig() *GitHubConfig {
	return &GitHubConfig{
		Username:   "<YOUR_USERNAME>",
		Repository: "<YOUR_REPOSITORY>",
	}
}

func NewOutputConfig() *OutputConfig {
	return &OutputConfig{
		Articles: NewOutputArticlesConfig(),
		Images:   NewOutputImagesConfig(),
	}
}

func NewOutputArticlesConfig() *OutputArticlesConfig {
	return &OutputArticlesConfig{
		Directory: "content/posts",
		Filename:  "%Y-%m-%d_%H%M%S.md",
	}
}

func NewOutputImagesConfig() *OutputImagesConfig {
	url := "/images/%Y-%m-%d_%H%M%S"
	return &OutputImagesConfig{
		Directory: "static/images/%Y-%m-%d_%H%M%S",
		Filename:  "[:id].png",
		BaseURL:   &url,
	}
}

func (c *OutputImagesConfig) URL() string {
	if c == nil || c.BaseURL == nil {
		return ""
	}
	return *c.BaseURL
}

func (c *Config) normalize() {
	if c.GitHub == nil {
		c.GitHub = NewGitHubConfig()
	}

	if c.Output == nil {
		if c.Hugo != nil {
			c.Output = &OutputConfig{}
		} else {
			c.Output = NewOutputConfig()
		}
	}

	if c.Output.Articles == nil {
		c.Output.Articles = &OutputArticlesConfig{}
	}
	if c.Output.Images == nil {
		c.Output.Images = &OutputImagesConfig{}
	}

	if c.Hugo == nil {
		return
	}

	if c.Hugo.Content != nil {
		if c.Output.Articles.Directory == "" {
			c.Output.Articles.Directory = c.Hugo.Content.Directory
		}
		if c.Output.Articles.Filename == "" {
			c.Output.Articles.Filename = c.Hugo.Content.Filename
		}
	}
	if c.Hugo.Images != nil {
		if c.Output.Images.Directory == "" {
			c.Output.Images.Directory = c.Hugo.Images.Directory
		}
		if c.Output.Images.Filename == "" {
			c.Output.Images.Filename = c.Hugo.Images.Filename
		}
		if c.Output.Images.BaseURL == nil {
			url := c.Hugo.Images.URL
			c.Output.Images.BaseURL = &url
		}
	}

	if c.Hugo.Directory != nil {
		if c.Output.Articles.Directory == "" {
			c.Output.Articles.Directory = c.Hugo.Directory.Articles
		}
		if c.Output.Images.Directory == "" {
			c.Output.Images.Directory = c.Hugo.Directory.Images
		}
	}
	if c.Hugo.Filename != nil {
		if c.Output.Articles.Filename == "" {
			c.Output.Articles.Filename = c.Hugo.Filename.Articles
		}
		if c.Output.Images.Filename == "" {
			c.Output.Images.Filename = c.Hugo.Filename.Images
		}
	}
	if c.Hugo.Url != nil && c.Output.Images.BaseURL == nil {
		url := c.Hugo.Url.Images
		c.Output.Images.BaseURL = &url
	}
}
