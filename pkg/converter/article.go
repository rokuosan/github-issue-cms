package converter

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
	"gopkg.in/yaml.v3"
)

// Article is the article for Hugo.
type Article struct {
	// Author is the author of the article.
	Author string `yaml:"author"`

	// Title is the title of the article.
	Title string `yaml:"title"`

	// Content is the content of the article.
	Content string `yaml:"-"`

	// Date is the date of the article.
	Date string `yaml:"date"`

	// Category is the category of the article.
	Category string `yaml:"categories"`

	// Tags is the tags of the article.
	Tags []string `yaml:"tags"`

	// Draft is the draft of the article.
	// If it is true, the article will not be published.
	Draft bool `yaml:"draft"`

	// ExtraFrontMatter is the extra front matter of the article.
	// It must be a valid YAML string.
	ExtraFrontMatter string `yaml:"-"`

	Key string `yaml:"-"`

	// Images is the images of the article.
	Images []*ImageDescriptor `yaml:"-"`
}

func (a *Article) parseDateTime() (time.Time, error) {
	var err error
	if t, err := time.Parse("2006-01-02T15:04:05Z", a.Date); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", a.Date); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("failed to parse date %s: %w", a.Date, err)
}

func (a *Article) Export(conf config.Config) {
	// Build String
	text, err := a.Transform()
	if err != nil {
		panic(err)
	}

	// Parse Date
	datetime, err := a.parseDateTime()
	if err != nil {
		slog.Error("skip exporting for this article due to failed to parse datetime", "error", err.Error(), "article", a)
		return
	}

	dest := conf.Hugo.Directory.Articles
	if dest == "" {
		slog.Error("Hugo directory is not set")
		return
	}
	dest = config.CompileTimeTemplate(datetime, dest)
	dest = filepath.Clean(dest)

	// Prepare directory
	if err := createDirectoryIfNotExist(dest); err != nil {
		slog.Error(fmt.Sprintf("Failed to create directory: %s", dest))
		return
	}

	// Write
	filename := conf.Hugo.Filename.Articles
	if filename == "" {
		slog.Error("Hugo filename is not set")
		return
	}
	filename = config.CompileTimeTemplate(datetime, filename)
	path := filepath.Join(dest, filename)
	if err := createFileAndWrite(path, text); err != nil {
		slog.Error(fmt.Sprintf("Failed to write file: %s", path))
		return
	}

	// Save images
	imageDir := conf.Hugo.Directory.Images
	if imageDir == "" {
		slog.Error("Hugo image directory is not set")
		return
	}
	imageDir = config.CompileTimeTemplate(datetime, imageDir)
	filename = conf.Hugo.Filename.Images
	filename = config.CompileTimeTemplate(datetime, filename)
	for _, image := range a.Images {
		f := strings.ReplaceAll(filename, "[:id]", fmt.Sprintf("%d", image.Id))
		image.Save(imageDir, f)
	}
}

func createDirectoryIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	}
	return nil
}

func createFileAndWrite(path string, content string) error {
	// Create file
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Write
	writer := bufio.NewWriter(file)
	_, err = writer.Write([]byte(content))
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

// Transform transforms the article to the markdown format.
func (a *Article) Transform() (string, error) {
	extra := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(a.ExtraFrontMatter), &extra)
	if err != nil {
		return "", err
	}

	// Overwrite self if extra has the same key
	// and delete the key from extra
	if author, ok := extra["author"]; ok {
		a.Author = author.(string)
		delete(extra, "author")
	}
	if title, ok := extra["title"]; ok {
		a.Title = title.(string)
		delete(extra, "title")
	}
	if date, ok := extra["date"]; ok {
		a.Date = date.(string)
		delete(extra, "date")
	}
	if categories, ok := extra["categories"]; ok {
		a.Category = categories.(string)
		delete(extra, "categories")
	}
	if tags, ok := extra["tags"]; ok {
		a.Tags = tags.([]string)
		delete(extra, "tags")
	}
	if draft, ok := extra["draft"]; ok {
		a.Draft = draft.(bool)
		delete(extra, "draft")
	}

	extraFrontMatter, err := yaml.Marshal(extra)
	if err != nil {
		return "", err
	}

	partial, err := yaml.Marshal(a)
	if err != nil {
		panic(err)
	}
	frontmatter := string(partial)
	frontmatter += string(extraFrontMatter)

	return fmt.Sprintf("---\n%s---\n\n%s\n", string(frontmatter), a.Content), nil
}

// Download downloads the image.
// Expected path is "path/to/image/".
func (d *ImageDescriptor) Save(path string, filename string) {
	// Download image
	sendRequest := func(includeToken bool) io.ReadCloser {
		req, err := http.NewRequest("GET", d.Url, nil)
		if err != nil {
			panic(err)
		}
		if includeToken {
			req.Header.Set("Authorization", "token "+config.GitHubToken)
		}
		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		// Check response
		contentType := resp.Header.Get("Content-Type")
		if resp.StatusCode != 200 || contentType != "image/png" {
			slog.Error(fmt.Sprintf("Bad Response: %+v", resp))
			_, _ = io.ReadAll(resp.Body)
			return nil
		}

		return resp.Body
	}

	// Send request
	containsToken := config.GitHubToken != ""
	body := sendRequest(containsToken)
	if body == nil {
		body = sendRequest(!containsToken)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(body)

	if body == nil {
		slog.Error(fmt.Sprintf("Failed to download image: %s", d.Url))
		return
	}

	// Create directory
	if err := createDirectoryIfNotExist(path); err != nil {
		slog.Error(fmt.Sprintf("Failed to create directory: %s", path))
		return
	}

	// Write
	path = filepath.Join(path, filename)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	written, err := io.Copy(file, body)
	if err != nil {
		panic(err)
	}
	slog.Debug(fmt.Sprintf("Downloaded: %s (%d bytes)", path, written))
}
