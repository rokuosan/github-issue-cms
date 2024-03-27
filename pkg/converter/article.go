package converter

// Article is the article for Hugo.
type Article struct {
	// Author is the author of the article.
	Author string `json:"author"`

	// Title is the title of the article.
	Title string `json:"title"`

	// Content is the content of the article.
	Content string `json:"content"`

	// Date is the date of the article.
	Date string `json:"date"`

	// Category is the category of the article.
	Category string `json:"category"`

	// Tags is the tags of the article.
	Tags []string `json:"tags"`

	// Draft is the draft of the article.
	// If it is true, the article will not be published.
	Draft bool `json:"draft"`

	// ExtraFrontMatter is the extra front matter of the article.
	// It must be a valid YAML string.
	ExtraFrontMatter string `json:"extra_front_matter"`
}
