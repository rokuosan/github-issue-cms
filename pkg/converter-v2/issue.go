package converter_v2

import (
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Article interface {
	Title() string
	Content() string
	Date() time.Time
	Author() string
	Tags() []string
	Category() string
	IsDraft() bool
	Images() []Image
}

type IssueArticle struct {
	*github.Issue

	// Parsed AST document and source
	doc    ast.Node
	source []byte

	frontMatter *FrontMatter
	images      []Image
}

var _ Markdownable = (*IssueArticle)(nil)
var _ Article = (*IssueArticle)(nil)

func NewIssueArticle(markdown goldmark.Markdown, issue *github.Issue) (*IssueArticle, error) {
	source := []byte(issue.GetBody())
	doc := markdown.Parser().Parse(text.NewReader(source))
	fm, err := ExtractFrontMatter(doc, source)
	if err != nil {
		return nil, err
	}

	return &IssueArticle{
		Issue: issue,

		doc:    doc,
		source: source,

		frontMatter: fm,
		images:      FindImages(doc, source),
	}, nil
}

func (a *IssueArticle) Title() string {
	// Title is determined by following priority:
	// 1. front matter `title`
	// 2. issue title
	if a.frontMatter != nil {
		if title, ok := a.frontMatter.ParseYAML()["title"].(string); ok {
			return title
		}
	}
	return a.GetTitle()
}

func (a *IssueArticle) Markdown() string {
	return a.GetBody()
}

// Author returns the author of the article.
func (a *IssueArticle) Author() string {
	// Author is determined by following priority:
	// 1. front matter `author`
	// 2. issue user login
	if a.frontMatter != nil {
		if author, ok := a.frontMatter.ParseYAML()["author"].(string); ok {
			return author
		}
	}
	return a.GetUser().GetLogin()
}

// Category returns the category of the article.
func (a *IssueArticle) Category() string {
	// Category is determined by following priority:
	// 1. front matter `category`
	// 2. Milestone of the issue
	if a.frontMatter != nil {
		d := a.frontMatter.ParseYAML()
		if category, ok := d["category"].(string); ok {
			return category
		}
		if categories, ok := d["categories"].([]interface{}); ok && len(categories) > 0 {
			return categories[0].(string)
		}
	}
	if a.GetMilestone() != nil {
		return a.GetMilestone().GetTitle()
	}
	return ""
}

// Content returns the content of the article without front matter.
func (a *IssueArticle) Content() string {
	// Remove front matter from the content if present
	if a.frontMatter != nil {
		fm := a.frontMatter.StringWithBackQuotes()
		return string(a.source[len(fm):])
	}
	return a.GetBody()
}

// Date implements Article.
func (a *IssueArticle) Date() time.Time {
	// Date is determined by following priority:
	// 1. front matter `date`
	// 2. issue closed date
	// 3. issue created date
	if a.frontMatter != nil {
		if date, ok := a.frontMatter.ParseYAML()["date"].(time.Time); ok {
			return date
		}
		if dateStr, ok := a.frontMatter.ParseYAML()["date"].(string); ok {
			if date, err := time.Parse(time.DateOnly, dateStr); err == nil {
				return date
			}
			if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
				return date
			}
		}
	}
	if !a.GetClosedAt().IsZero() {
		return a.GetClosedAt().Time
	}
	return a.GetCreatedAt().Time
}

func (a *IssueArticle) Images() []Image {
	// Images are extracted during conversion and stored in the Images field.
	return a.images
}

// IsDraft implements Article.
func (a *IssueArticle) IsDraft() bool {
	// Draft status is determined by following priority:
	// 1. front matter `draft`
	// 2. issue closed status. if closed, it's not draft.
	if a.frontMatter != nil {
		if draft, ok := a.frontMatter.ParseYAML()["draft"].(bool); ok {
			return draft
		}
	}
	return a.State != nil && *a.State == "closed"
}

// Tags implements Article.
func (a *IssueArticle) Tags() []string {
	// Tags are determined by issue labels and front matter `tags`.
	tags := map[string]string{}
	for _, label := range a.Labels {
		tags[label.GetName()] = label.GetName()
	}
	if a.frontMatter != nil {
		if fmTags, ok := a.frontMatter.ParseYAML()["tags"].([]interface{}); ok {
			for _, tag := range fmTags {
				if tagStr, ok := tag.(string); ok {
					tags[tagStr] = tagStr
				}
			}
		}
	}
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	return tagList
}
