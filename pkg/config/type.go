package config

// Config package is a package for configuration.
// If you change the configuration, you also need to change ``config.Generate()`` (config.go).

type Config struct {
	GitHub *GitHubConfig `yaml:"github"`
	Hugo   *HugoConfig   `yaml:"hugo"`
}

type GitHubConfig struct {
	Username         string `yaml:"username"`
	Repository       string `yaml:"repository"`
	AllowedAuthors []string `yaml:"allowed_authors" mapstructure:"allowed_authors"`
}

type HugoConfig struct {
	Bundle    string               `yaml:"bundle,omitempty"`
	Directory *HugoDirectoryConfig `yaml:"directory"`
	Filename  *HugoFilenameConfig  `yaml:"filename"`
	Url       *HugoURLConfig       `yaml:"url"`
}

type HugoDirectoryConfig struct {
	Articles string `yaml:"articles"`
	Images   string `yaml:"images"`
}

type HugoFilenameConfig struct {
	Articles string `yaml:"articles"`
	Images   string `yaml:"images"`
}

type HugoURLConfig struct {
	AppendSlash bool   `yaml:"appendSlash,omitempty"`
	Images      string `yaml:"images"`
}

func (c *GitHubConfig) RepositoryURL() string {
	return "https://github.com/" + c.Username + "/" + c.Repository
}

func NewConfig() *Config {
	return &Config{
		GitHub: NewGitHubConfig(),
		Hugo:   NewHugoConfig(),
	}
}

func NewGitHubConfig() *GitHubConfig {
	return &GitHubConfig{
		Username:   "<YOUR_USERNAME>",
		Repository: "<YOUR_REPOSITORY>",
	}
}

func NewHugoConfig() *HugoConfig {
	return &HugoConfig{
		Directory: NewHugoDirectoryConfig(),
		Filename:  NewHugoFilenameConfig(),
		Url:       NewHugoURLConfig(),
	}
}

func NewHugoDirectoryConfig() *HugoDirectoryConfig {
	return &HugoDirectoryConfig{
		Articles: "content/posts",
		Images:   "static/images/%Y-%m-%d_%H%M%S",
	}
}

func NewHugoFilenameConfig() *HugoFilenameConfig {
	return &HugoFilenameConfig{
		Articles: "%Y-%m-%d_%H%M%S.md",
		Images:   "[:id].png",
	}
}

func NewHugoURLConfig() *HugoURLConfig {
	return &HugoURLConfig{
		Images: "/images/%Y-%m-%d_%H%M%S",
	}
}
