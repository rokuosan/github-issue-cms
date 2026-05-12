package core

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/google/go-github/v86/github"
	"github.com/pelletier/go-toml/v2"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"gopkg.in/yaml.v3"
)

var (
	regexURLCandidate = regexp.MustCompile(`https://[^\s<>"')\]]+`)
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
	images := extractTargetImages(content, time, s.config.Output.Images.TargetURLs())

	var tags []string
	excludedLabels := map[string]struct{}{}
	if s.config.GitHub != nil {
		excludedLabels = make(map[string]struct{}, len(s.config.GitHub.Labels))
		for _, label := range s.config.GitHub.Labels {
			excludedLabels[label] = struct{}{}
		}
	}
	for _, label := range issue.Labels {
		name := label.GetName()
		if _, ok := excludedLabels[name]; ok {
			continue
		}
		tags = append(tags, name)
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

func extractTargetImages(content string, time string, targetURLs []string) []*Image {
	var images []*Image
	seen := map[string]struct{}{}
	matches := regexURLCandidate.FindAllString(content, -1)
	for _, match := range matches {
		candidate := strings.TrimRight(match, ".,:;!?`")
		if !matchesTargetURL(candidate, targetURLs) {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		images = append(images, NewImage(candidate, time, len(images)))
	}
	return images
}

func matchesTargetURL(raw string, targetURLs []string) bool {
	parsedRaw, err := url.Parse(raw)
	if err != nil {
		return false
	}

	for _, targetURL := range targetURLs {
		if targetURL == "" {
			continue
		}
		if matchTargetPattern(raw, parsedRaw, targetURL) {
			return true
		}
	}
	return false
}

func matchTargetPattern(raw string, parsedRaw *url.URL, targetURL string) bool {
	if !strings.Contains(targetURL, "*") {
		return strings.HasPrefix(raw, targetURL)
	}

	parsedTarget, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	if parsedTarget.Scheme != "" && !strings.EqualFold(parsedTarget.Scheme, parsedRaw.Scheme) {
		return false
	}
	if parsedTarget.Host != "" {
		matched, err := path.Match(strings.ToLower(parsedTarget.Host), strings.ToLower(parsedRaw.Host))
		if err != nil || !matched {
			return false
		}
	}
	if !matchesPathPrefixPattern(parsedRaw.Path, parsedTarget.Path) {
		return false
	}

	return true
}

func matchesPathPrefixPattern(rawPath string, targetPath string) bool {
	if targetPath == "" {
		return true
	}
	if !strings.Contains(targetPath, "*") {
		return strings.HasPrefix(rawPath, targetPath)
	}

	for i := 1; i <= len(rawPath); i++ {
		matched, err := path.Match(targetPath, rawPath[:i])
		if err == nil && matched {
			return true
		}
	}

	return false
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
