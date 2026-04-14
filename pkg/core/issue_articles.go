package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v67/github"
	"github.com/pelletier/go-toml/v2"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"gopkg.in/yaml.v3"
)

var (
	regexMarkdownImage = regexp.MustCompile(`!\[[^\]]*]\((.+?)\)`)
	regexHTMLImage     = regexp.MustCompile(`<img width="\d+" alt="(\w+)" src="(\S+)">`)
)

// ArticleService converts issues into articles.
type ArticleService struct {
	config         config.Config
	metadataParser metadataParser
}

type metadataBlock struct {
	Raw    string
	Values FrontMatter
}

type metadataBlockParser interface {
	Parse(body string) (metadataBlock, bool, error)
}

type metadataParser struct {
	parsers []metadataBlockParser
}

type metadataFormat string

const (
	metadataFormatYAML metadataFormat = "yaml"
	metadataFormatTOML metadataFormat = "toml"
)

type codeFenceMetadataParser struct{}

type delimitedMetadataParser struct {
	delimiter string
	format    metadataFormat
}

// NewArticleService creates a new ArticleService.
func NewArticleService(conf config.Config) *ArticleService {
	return &ArticleService{
		config:         conf,
		metadataParser: newMetadataParser(),
	}
}

// ConvertIssueToArticle converts a GitHub issue into an Article.
func (s *ArticleService) ConvertIssueToArticle(issue *github.Issue) *Article {
	if issue.IsPullRequest() {
		return nil
	}

	content := insertTrailingNewline(removeCR(issue.GetBody()))

	frontMatter, err := s.metadataParser.Parse(content)
	if err != nil {
		frontMatter = metadataBlock{Values: EmptyFrontMatter()}
	}
	content = strings.Replace(content, frontMatter.Raw, "", 1)
	content = strings.TrimLeft(content, "\n")

	time := issue.GetCreatedAt().Format("2006-01-02_150405")

	var images []*Image
	images = append(images, s.findImages(content, regexMarkdownImage, time, 0)...)
	images = append(images, s.findImages(content, regexHTMLImage, time, len(images))...)

	var tags []string
	for _, label := range issue.Labels {
		tags = append(tags, label.GetName())
	}

	return &Article{
		Author:      issue.GetUser().GetLogin(),
		Title:       issue.GetTitle(),
		Date:        issue.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		Category:    issue.GetMilestone().GetTitle(),
		Draft:       issue.GetState() == "open",
		Content:     content,
		FrontMatter: frontMatter.Values,
		Tags:        tags,
		Key:         time,
		Images:      images,
	}
}

func (s *ArticleService) findImages(content string, re *regexp.Regexp, time string, offset int) []*Image {
	var images []*Image

	matches := re.FindAllStringSubmatch(content, -1)
	for i, m := range matches {
		id := i + offset
		url := m[1]
		if len(m) > 2 {
			url = m[2]
		}
		images = append(images, NewImage(url, time, id))
	}

	return images
}

func newMetadataParser() metadataParser {
	return metadataParser{
		parsers: []metadataBlockParser{
			codeFenceMetadataParser{},
			delimitedMetadataParser{delimiter: "---", format: metadataFormatYAML},
			delimitedMetadataParser{delimiter: "+++", format: metadataFormatTOML},
		},
	}
}

func (p metadataParser) Parse(body string) (metadataBlock, error) {
	for _, parser := range p.parsers {
		block, matched, err := parser.Parse(body)
		if err != nil {
			return metadataBlock{}, err
		}
		if matched {
			return block, nil
		}
	}
	return metadataBlock{}, fmt.Errorf("front matter not found")
}

func (codeFenceMetadataParser) Parse(body string) (metadataBlock, bool, error) {
	prefix, trimmed := splitLeadingWhitespace(body)
	if !strings.HasPrefix(trimmed, "```") {
		return metadataBlock{}, false, nil
	}

	rest := strings.TrimPrefix(trimmed, "```")
	lineEnd := strings.IndexByte(rest, '\n')
	if lineEnd < 0 {
		return metadataBlock{}, false, nil
	}

	infoString := strings.TrimSpace(rest[:lineEnd])
	afterHeader := rest[lineEnd+1:]
	closingOffset := 0
	closingLength := len("```")
	if strings.HasPrefix(afterHeader, "```") {
		closingOffset = 0
	} else {
		closingOffset = strings.Index(afterHeader, "\n```")
		if closingOffset < 0 {
			return metadataBlock{}, false, nil
		}
		closingLength = len("\n```")
	}

	content := afterHeader[:closingOffset]
	rawLength := 3 + lineEnd + 1 + closingOffset + closingLength
	raw := prefix + trimmed[:rawLength]

	format := metadataFormatYAML
	switch strings.ToLower(infoString) {
	case "", "yaml", "yml":
		format = metadataFormatYAML
	case "toml":
		format = metadataFormatTOML
	}

	normalized, err := normalizeMetadata(content, format)
	if err != nil {
		return metadataBlock{}, false, fmt.Errorf("failed to parse front matter: %w", err)
	}

	values, err := parseNormalizedFrontMatter(normalized)
	if err != nil {
		return metadataBlock{}, false, fmt.Errorf("failed to parse front matter: %w", err)
	}

	return metadataBlock{Raw: raw, Values: values}, true, nil
}

func (p delimitedMetadataParser) Parse(body string) (metadataBlock, bool, error) {
	prefix, trimmed := splitLeadingWhitespace(body)
	if !strings.HasPrefix(trimmed, p.delimiter+"\n") {
		return metadataBlock{}, false, nil
	}

	content := trimmed[len(p.delimiter)+1:]
	endMarker := "\n" + p.delimiter
	end := strings.Index(content, endMarker)
	if end < 0 {
		return metadataBlock{}, false, nil
	}

	raw := prefix + trimmed[:len(p.delimiter)+1+end+len(endMarker)]
	normalized, err := normalizeMetadata(content[:end], p.format)
	if err != nil {
		return metadataBlock{}, false, fmt.Errorf("failed to parse front matter: %w", err)
	}

	values, err := parseNormalizedFrontMatter(normalized)
	if err != nil {
		return metadataBlock{}, false, fmt.Errorf("failed to parse front matter: %w", err)
	}

	return metadataBlock{Raw: raw, Values: values}, true, nil
}

func splitLeadingWhitespace(body string) (string, string) {
	trimmed := strings.TrimLeft(body, "\n\r\t ")
	return body[:len(body)-len(trimmed)], trimmed
}

func normalizeMetadata(content string, format metadataFormat) (string, error) {
	var values map[string]any

	switch format {
	case metadataFormatYAML:
		if err := yaml.Unmarshal([]byte(content), &values); err != nil {
			return "", err
		}
	case metadataFormatTOML:
		if err := toml.Unmarshal([]byte(content), &values); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported metadata format: %s", format)
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func parseNormalizedFrontMatter(content string) (FrontMatter, error) {
	values := map[string]any{}
	if strings.TrimSpace(content) == "" {
		return EmptyFrontMatter(), nil
	}
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		return EmptyFrontMatter(), err
	}
	return NewFrontMatter(values), nil
}

func removeCR(content string) string {
	return strings.ReplaceAll(content, "\r", "")
}

func insertTrailingNewline(content string) string {
	if !strings.HasSuffix(content, "\n") {
		return content + "\n"
	}
	return content
}
