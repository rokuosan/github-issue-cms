package converter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/rokuosan/github-issue-cms/pkg/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Converter struct {
	*github.Client
	token string
}

type ImageDescriptor struct {
	Url  string
	Time string
	Id   int
}

func NewConverter() *Converter {
	slog.Debug("Setting up GitHub Client...")
	if config.GitHubToken == "" {
		slog.Error("Failed to initialize GitHub Client due to the Token is empty.")
		return nil
	}

	client := github.NewClient(nil).WithAuthToken(config.GitHubToken)
	if client == nil {
		slog.Error("Failed to initialize GitHub Client due to the Token is invalid.")
		return nil
	}
	slog.Debug("Successfully created GitHub Client")
	return &Converter{Client: client, token: config.GitHubToken}
}

func (c *Converter) GetIssues() []*github.Issue {
	username := viper.GetString("github.username")
	repository := viper.GetString("github.repository")
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

func (c *Converter) downloadImage(url string, time string, number int) {
	imageDirectory := config.Get().Hugo.Directory.Images
	path := filepath.Clean(imageDirectory)

	base := filepath.Join(path, time)
	dest := filepath.Join(base, strconv.Itoa(number)+".png")

	// Create directory
	if _, err := os.Stat(base); os.IsNotExist(err) {
		err := os.MkdirAll(base, 0777)
		if err != nil {
			panic(err)
		}
	}

	slog.Debug("Downloading: " + url)
	file, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	containsToken := true
request:
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	if containsToken {
		req.Header.Set("Authorization", "token "+c.token)
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	// Check response
	contentType := resp.Header.Get("Content-Type")

	if resp.StatusCode != 200 || contentType != "image/png" {
		slog.Error(fmt.Sprintf("Response: %d %s", resp.StatusCode, contentType))

		if resp.StatusCode == 400 && contentType == "application/xml" {
			data, _ := io.ReadAll(resp.Body)
			if strings.Contains(string(data), "Unsupported Authorization Type") {
				if containsToken {
					// Retry once
					containsToken = false
					goto request
				}
			}
		}

		_ = file.Close()
		err := os.Remove(dest)
		if err != nil {
			panic(err)
		}

		return
	}
	slog.Debug(fmt.Sprintf("Response: %d %s", resp.StatusCode, contentType))

	// Write the body to file
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		panic(err)
	}
	slog.Debug("Downloaded image: " + dest + " (" + fmt.Sprintf("%d", written) + " bytes)")
}

func (c *Converter) SaveImages(descriptors []*ImageDescriptor) {
	for _, d := range descriptors {
		c.downloadImage(d.Url, d.Time, d.Id)
	}
}

// IssueToArticle converts an issue into article. Returns an Article object and array of ImageDescriptor.
func (c *Converter) IssueToArticle(issue *github.Issue) *Article {
	if issue.IsPullRequest() {
		return nil
	}
	num := strconv.Itoa(issue.GetNumber())
	slog.Debug("Converting #" + num + "...")

	// Get issue content
	content := strings.Replace(issue.GetBody(), "\r", "", -1)
	if !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}

	// Get front matter and remove it from original content
	frontMatter := func() []string {
		re := regexp.MustCompile("(?s)^\\s*```\\n([^`]*)\\n```")
		match := re.FindStringSubmatch(content)

		if len(match) < 2 {
			return nil
		}
		content := []byte(match[1])
		t := make(map[interface{}]interface{})
		err := yaml.Unmarshal(content, &t)
		if err != nil {
			return nil
		}

		return match
	}()
	if frontMatter != nil {
		content = strings.Replace(content, frontMatter[0], "", 1)
	} else {
		frontMatter = []string{"", ""}
	}
	content = strings.TrimLeft(content, "\n")

	// Make image url
	imageUrlPath := config.Get().Hugo.Url.Images
	imageUrlPath = config.CompileTimeTemplate(issue.GetCreatedAt().Time, imageUrlPath)
	filename := config.Get().Hugo.Filename.Images
	imageUrl := func(alt string, id int) string {
		name := strings.ReplaceAll(filename, "[:id]", strconv.Itoa(id))
		path := filepath.Join(imageUrlPath, name)
		path = filepath.Clean(path)
		return "![" + alt + "](" + path + ")"
	}

	// Replace image URL of Markdown style
	var imageDescriptors []*ImageDescriptor
	offset := 0
	time := issue.GetCreatedAt().Format("2006-01-02_150405")
	re := regexp.MustCompile(`!\[image*]\((.*)\)`)
	match := re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		replaced := imageUrl(m[1], i)
		imageDescriptors = append(imageDescriptors, &ImageDescriptor{Url: m[1], Time: time, Id: i})
		slog.Debug("Found: (ID:" + strconv.Itoa(i) + ") " + time + " " + m[1])
		content = strings.Replace(content, m[0], replaced, -1)
		offset = i + 1
	}

	// Replace image URL of HTML style
	re = regexp.MustCompile(`<img width="\d+" alt="(\w+)" src="(\S+)">`)
	match = re.FindAllStringSubmatch(content, -1)
	for i, m := range match {
		offset += i
		replaced := imageUrl(m[1], offset)
		imageDescriptors = append(imageDescriptors, &ImageDescriptor{Url: m[2], Time: time, Id: offset})
		slog.Debug("Found: " + strconv.Itoa(offset) + " " + time + " " + m[2])
		content = strings.Replace(content, m[0], replaced, -1)
	}

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
