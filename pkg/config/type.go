package config

type Config struct {
	GitHub *GitHubConfig `yaml:"github"`
	Hugo   *HugoConfig   `yaml:"hugo"`
}

type GitHubConfig struct {
	Username   string `yaml:"username"`
	Repository string `yaml:"repository"`
}

type HugoConfig struct {
	Bundle    string               `yaml:"bundle"`
	Direcotry *HugoDirectoryConfig `yaml:"directory"`
	Url       *HugoURLConfig       `yaml:"url"`
}

type HugoDirectoryConfig struct {
	Articles string `yaml:"articles"`
}

type HugoURLConfig struct {
	AppendSlash bool   `yaml:"appendSlash"`
	Images      string `yaml:"images"`
}

func (c *GitHubConfig) RepositoryURL() string {
	return "https://github.com/" + c.Username + "/" + c.Repository
}
