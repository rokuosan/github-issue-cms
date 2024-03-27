package converter

import (
	"context"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/viper"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Client struct {
	*github.Client
	token string
}

func GetClient(token string) *Client {
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
	return &Client{client, token}
}

func (c *Client) GetIssues() []*github.Issue {
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

	issuesAndPRs, _, err := client.Issues.ListByRepo(
		context.Background(),
		username,
		repository,
		&github.IssueListByRepoOptions{
			State: "all",
		})
	if err != nil {
		panic(err)
	}
	var issues []*github.Issue
	for _, item := range issuesAndPRs {
		if item.IsPullRequest() {
			continue
		}
		issues = append(issues, item)
	}

	return issues
}

func (c *Client) downloadImage(url string, time string, number int) {
	imageUrl := viper.GetString("hugo.url.images")
	path := filepath.Join("./static", imageUrl)

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
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "token "+c.token)
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
}
