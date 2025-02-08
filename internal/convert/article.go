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

type ArticleImage struct {
	Source   string `yaml:"source"`
	Alt      string `yaml:"alt"`
	Original string `yaml:"original"`
}

// contentWithoutCodeBlocks returns the content of the article without code blocks
func (a *Article) contentWithoutCodeBlocks() string {
	content := a.Content

	content = regex.MarkdownCodeBlock.ReplaceAllString(content, "")
	content = regex.MarkdownInlineCodeBlock.ReplaceAllString(content, "")

	return content
}

func (a *Article) markdownLinks() []string {
	return regex.MarkdownLink.FindAllString(a.contentWithoutCodeBlocks(), -1)
}

func (a *Article) images() []ArticleImage {
	content := a.contentWithoutCodeBlocks()
	var matches [][]string
	matches = append(matches, regex.MarkdownImage.FindAllStringSubmatch(content, -1)...)
	matches = append(matches, regex.HTMLImage.FindAllStringSubmatch(content, -1)...)
	images := make([]ArticleImage, len(matches))

	for i, match := range matches {
		var src, alt string
		if src = match[2]; src == "" {
			src = match[3]
		}
		if alt = match[1]; alt == "" {
			alt = match[4]
		}

		images[i] = ArticleImage{
			Source:   src,
			Alt:      alt,
			Original: match[0],
		}
	}

	return images
}
