//go:generate mockgen -source=$GOFILE -package=${GOPACKAGE}_mock -destination=./mock/$GOFILE
package converter

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v67/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"gopkg.in/yaml.v3"
)

// Converter interface defines the methods for converting GitHub issues to articles
type Converter interface {
	GetIssues() []*github.Issue
	IssueToArticle(issue *github.Issue) *Article
}

type converterImpl struct {
	*github.Client
	token  string
	config Config
}

type Config struct {
	config.Config
	Token string
}

type ImageDescriptor struct {
	Url  string
	Time string
	Id   int
}

var (
	errFrontMatterNotFound = func() error { return fmt.Errorf("front matter not found") }
	errInvalidFrontMatter  = func(err error) error { return fmt.Errorf("failed to parse front matter: %w", err) }
)

var (
	regexMarkdownImage = regexp.MustCompile(`!\[image*]\((.*)\)`)
	regexHTMLImage     = regexp.MustCompile(`<img width="\d+" alt="(\w+)" src="(\S+)">`)
)

func NewConverter(config Config, token string) Converter {
	slog.Debug("Setting up GitHub Client...")
	if token == "" {
		slog.Error("Failed to initialize GitHub Client due to the Token is empty.")
		return nil
	}

	client := github.NewClient(nil).WithAuthToken(token)
	if client == nil {
		slog.Error("Failed to initialize GitHub Client due to the Token is invalid.")
		return nil
	}

	slog.Debug("Successfully created GitHub Client")
	return &converterImpl{
		Client: client,
		token:  token,
		config: config,
	}
}

func (c *converterImpl) GetIssues() []*github.Issue {
	username := c.config.GitHub.Username
	repository := c.config.GitHub.Repository
	if username == "" || repository == "" {
		slog.Error("Please set github username and repository name in gic.config.yaml")
		return nil
	}

	client := c.Client
	if client == nil {
		slog.Error("Client is nil")
		return nil
	}

	slog.Debug("Collecting Issues...")
	var issues []*github.Issue
	var rate github.Rate
	nextPage := 1
	for nextPage != 0 {
		issuesAndPRs, resp, err := client.Issues.ListByRepo(
			context.Background(),
			username,
			repository,
			&github.IssueListByRepoOptions{
				State: "all",
				ListOptions: github.ListOptions{
					PerPage: 200,
					Page:    nextPage,
				},
			})

		if err != nil {
			if strings.Contains(err.Error(), "401 Bad credentials") {
				slog.Error("Invalid API Token; Please check your GitHub token.")
				return nil
			} else {
				panic(err)
			}
		}

		var list []*github.Issue
		for _, item := range issuesAndPRs {
			if item.IsPullRequest() {
				continue
			}
			list = append(list, item)
		}
		issues = append(issues, list...)

		nextPage = resp.NextPage
		rate = resp.Rate
	}
	slog.Debug("Get issues - " + strconv.Itoa(len(issues)))
	slog.Debug(fmt.Sprintf("Remaining Rate Limit: %d/%d (Reset: %s)", rate.Remaining, rate.Limit, rate.Reset))

	return issues
}

func (c *converterImpl) removeCR(content string) string {
	return strings.ReplaceAll(content, "\r", "")
}

func (c *converterImpl) insertTrailingNewline(content string) string {
	if !strings.HasSuffix(content, "\n") {
		return content + "\n"
	}
	return content
}

// IssueToArticle converts an issue into article. Returns an Article object and array of ImageDescriptor.
func (c *converterImpl) IssueToArticle(issue *github.Issue) *Article {
	if issue.IsPullRequest() {
		return nil
	}

	if c.config.GitHub != nil && len(c.config.GitHub.AllowedAuthors) > 0 {
		author := ""
		if issue.User != nil {
			author = issue.GetUser().GetLogin()
		}
		allowed := slices.Contains(c.config.GitHub.AllowedAuthors, author)
		if !allowed {
			slog.Debug(fmt.Sprintf("Author '%s' is not allowed. Skipping issue #%d", author, issue.GetNumber()))
			return nil
		}
	}

	num := strconv.Itoa(issue.GetNumber())
	slog.Debug("Converting #" + num + "...")

	// Get issue content
	content := issue.GetBody()
	content = c.removeCR(content)
	content = c.insertTrailingNewline(content)

	// Get front matter and remove it from original content
	frontMatter, err := c.extractFrontMatter(content)
	if err != nil {
		slog.Debug("Front matter not found: " + err.Error())
		frontMatter = definedOr(&frontMatter, []string{"", ""})
	}
	content = strings.Replace(content, frontMatter[0], "", 1)
	content = strings.TrimLeft(content, "\n")

	// Make image base url
	baseURL := config.CompileTimeTemplate(issue.GetCreatedAt().Time, c.config.Hugo.Url.Images)

	// Replace image URL of Markdown style
	var imageDescriptors []*ImageDescriptor
	time := issue.GetCreatedAt().Format("2006-01-02_150405")
	content, ids := c.replaceImageURL(replaceImageURLInput{
		re:      regexMarkdownImage,
		content: content,
		baseURL: baseURL,
		time:    time,
	})
	imageDescriptors = append(imageDescriptors, ids...)

	// Replace image URL of HTML style
	content, ids = c.replaceImageURL(replaceImageURLInput{
		re:      regexHTMLImage,
		content: content,
		baseURL: baseURL,
		time:    time,
		offset:  len(imageDescriptors),
	})
	imageDescriptors = append(imageDescriptors, ids...)

	// Tags
	var tags []string
	for _, label := range issue.Labels {
		tags = append(tags, label.GetName())
	}

	return &Article{
		Author:           issue.GetUser().GetLogin(),
		Title:            issue.GetTitle(),
		Date:             issue.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		Category:         issue.GetMilestone().GetTitle(),
		Draft:            issue.GetState() == "open",
		Content:          content,
		ExtraFrontMatter: frontMatter[1],
		Tags:             tags,
		Key:              time,
		Images:           imageDescriptors,
	}
}

func (c *converterImpl) extractFrontMatter(body string) ([]string, error) {
	re := regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```")
	match := re.FindStringSubmatch(body)

	if len(match) < 2 {
		return nil, errFrontMatterNotFound()
	}

	// Check if the front matter is valid YAML
	content := []byte(match[1])
	t := make(map[interface{}]interface{})
	err := yaml.Unmarshal(content, &t)
	if err != nil {
		return nil, errInvalidFrontMatter(err)
	}

	return match, nil
}

func (c *converterImpl) createMarkdownImageExpression(path, alt string, id int) string {
	name := strings.ReplaceAll(c.config.Hugo.Filename.Images, "[:id]", strconv.Itoa(id))
	p := filepath.Join(path, name)
	return fmt.Sprintf("![%s](%s)", alt, p)
}

type replaceImageURLInput struct {
	re      *regexp.Regexp
	content string
	baseURL string
	time    string
	offset  int
}

func (c *converterImpl) replaceImageURL(input replaceImageURLInput) (string, []*ImageDescriptor) {
	var ids []*ImageDescriptor

	match := input.re.FindAllStringSubmatch(input.content, -1)
	for i, m := range match {
		n := i + input.offset

		replaced := c.createMarkdownImageExpression(input.baseURL, m[1], n)
		ids = append(ids, &ImageDescriptor{Url: m[1], Time: input.time, Id: n})

		slog.Debug(fmt.Sprintf("Found: (ID:%d) %s %s", n, input.time, m[1]))
		input.content = strings.ReplaceAll(input.content, m[0], replaced)
	}

	return input.content, ids
}
