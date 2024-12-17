package convert

type Article struct {
	Title            string   `yaml:"title"`
	Content          string   `yaml:"-"`
	Author           string   `yaml:"author"`
	Authors          []string `yaml:"authors"`
	Date             string   `yaml:"date"`
	Categories       []string `yaml:"categories"`
	Tags             []string `yaml:"tags"`
	Draft            bool     `yaml:"draft"`
	ExtraFrontMatter string   `yaml:"-"`
}
