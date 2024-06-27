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

func (g *GitHubConfig) RepositoryURL() string {
	return "https://github.com/" + g.Username + "/" + g.Repository
}
