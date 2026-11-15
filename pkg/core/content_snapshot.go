package core

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rokuosan/github-issue-cms/pkg/config"
)

// ContentArticle is the public JSON representation of an issue article.
type ContentArticle struct {
	ID       int            `json:"id"`
	Author   string         `json:"author"`
	Title    string         `json:"title"`
	Content  string         `json:"content"`
	Date     string         `json:"date"`
	Category string         `json:"category"`
	Tags     []string       `json:"tags"`
	Draft    bool           `json:"draft"`
	Metadata map[string]any `json:"metadata"`
}

// ArticleSummary is an article without its body or custom metadata.
type ArticleSummary struct {
	Summary  string   `json:"summary"`
	ID       int      `json:"id"`
	Author   string   `json:"author"`
	Title    string   `json:"title"`
	Date     string   `json:"date"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Draft    bool     `json:"draft"`
	URL      string   `json:"url"`
}

// Pagination describes the filtered result set.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// ArticlePage contains one page of article summaries.
type ArticlePage struct {
	Items      []ArticleSummary `json:"items"`
	Pagination Pagination       `json:"pagination"`
}

// ContentQuery selects and orders articles before pagination.
type ContentQuery struct {
	Page, PerPage         int
	Q                     string
	Tags                  []string
	Category, Sort, Order string
}

// ContentAsset is an attachment fetched while building the snapshot.
type ContentAsset struct {
	Data        []byte
	ContentType string
}

// ContentSnapshot holds the complete data needed to serve content offline.
type ContentSnapshot struct {
	Articles    []ContentArticle
	GeneratedAt time.Time
	PageSize    int
	Assets      map[string]ContentAsset
}

// NewContentSnapshot validates and copies articles into a deterministic snapshot.
func NewContentSnapshot(articles []ContentArticle, pageSize int) (*ContentSnapshot, error) {
	if pageSize < 1 || pageSize > 100 {
		return nil, fmt.Errorf("page size must be between 1 and 100")
	}
	s := &ContentSnapshot{Articles: make([]ContentArticle, 0, len(articles)), GeneratedAt: time.Now().UTC(), PageSize: pageSize, Assets: map[string]ContentAsset{}}
	seen := map[int]bool{}
	for _, article := range articles {
		if article.ID < 1 || seen[article.ID] {
			return nil, fmt.Errorf("invalid or duplicate article ID %d", article.ID)
		}
		seen[article.ID] = true
		date, err := parseContentDate(article.Date)
		if err != nil {
			return nil, fmt.Errorf("article %d: %w", article.ID, err)
		}
		article.Date = date.UTC().Format(time.RFC3339Nano)
		article.Tags = append([]string{}, article.Tags...)
		article.Metadata = NewFrontMatter(article.Metadata).Values()
		s.Articles = append(s.Articles, article)
	}
	sortArticles(s.Articles, "date", "desc")
	return s, nil
}

// LoadContentSnapshot fetches issues and attachments once, before serving requests.
func LoadContentSnapshot(ctx context.Context, store IssueStore, assets AssetFetcher, conf config.Config, pageSize int, includeDrafts bool) (*ContentSnapshot, error) {
	if pageSize < 1 || pageSize > 100 {
		return nil, fmt.Errorf("page size must be between 1 and 100")
	}
	if conf.GitHub == nil || conf.GitHub.Username == "" || conf.GitHub.Repository == "" {
		return nil, fmt.Errorf("GitHub username and repository are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	issues, err := store.ListIssues(ctx, IssueListQuery{Username: conf.GitHub.Username, Repository: conf.GitHub.Repository, Labels: conf.GitHub.Labels})
	if err != nil {
		return nil, err
	}
	service := NewArticleService(conf)
	articles := []ContentArticle{}
	downloaded := map[string]ContentAsset{}
	for _, issue := range issues {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		article := service.ConvertIssueToArticle(issue)
		if article == nil {
			continue
		}
		extra := article.FrontMatter.Values()
		ApplyFrontMatterOverrides(article, extra)
		FilterArticleTags(article, conf)
		if article.Draft && !includeDrafts {
			continue
		}
		replacements := []string{}
		for _, image := range article.Images {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if assets == nil {
				return nil, fmt.Errorf("asset fetcher is required for article %d", issue.GetNumber())
			}
			asset, err := assets.Fetch(ctx, image)
			if err != nil {
				return nil, fmt.Errorf("article %d attachment: %w", issue.GetNumber(), err)
			}
			data, readErr := io.ReadAll(asset.Body)
			closeErr := asset.Body.Close()
			if readErr != nil {
				return nil, fmt.Errorf("read attachment: %w", readErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close attachment: %w", closeErr)
			}
			name := "/images/" + strconv.Itoa(issue.GetNumber()) + "/" + strconv.Itoa(image.ID) + extensionFromContentType(asset.ContentType)
			downloaded[name] = ContentAsset{Data: data, ContentType: asset.ContentType}
			replacements = append(replacements, image.URL, name)
		}
		if len(replacements) > 0 {
			article.Content = strings.NewReplacer(replacements...).Replace(article.Content)
		}
		articles = append(articles, ContentArticle{ID: issue.GetNumber(), Author: article.Author, Title: article.Title, Content: article.Content, Date: article.Date, Category: article.Category, Tags: article.Tags, Draft: article.Draft, Metadata: extra})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s, err := NewContentSnapshot(articles, pageSize)
	if err != nil {
		return nil, err
	}
	s.Assets = downloaded
	return s, nil
}

// Query filters, sorts, and paginates the snapshot without network access.
func (s *ContentSnapshot) Query(q ContentQuery) (ArticlePage, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PerPage == 0 {
		q.PerPage = s.PageSize
	}
	if q.Page < 1 {
		return ArticlePage{}, fmt.Errorf("page must be positive")
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		return ArticlePage{}, fmt.Errorf("per_page must be between 1 and 100")
	}
	if q.Sort == "" {
		q.Sort = "date"
	}
	if q.Order == "" {
		q.Order = "desc"
	}
	if q.Sort != "date" && q.Sort != "title" {
		return ArticlePage{}, fmt.Errorf("sort must be date or title")
	}
	if q.Order != "asc" && q.Order != "desc" {
		return ArticlePage{}, fmt.Errorf("order must be asc or desc")
	}
	matches := []ContentArticle{}
	search := strings.ToLower(q.Q)
	for _, article := range s.Articles {
		if search != "" && !strings.Contains(strings.ToLower(article.Title), search) && !strings.Contains(strings.ToLower(article.Content), search) {
			continue
		}
		if q.Category != "" && article.Category != q.Category {
			continue
		}
		matched := true
		for _, tag := range q.Tags {
			found := false
			for _, actual := range article.Tags {
				if tag == actual {
					found = true
					break
				}
			}
			if !found {
				matched = false
				break
			}
		}
		if matched {
			matches = append(matches, article)
		}
	}
	sortArticles(matches, q.Sort, q.Order)
	totalPages := len(matches) / q.PerPage
	if len(matches)%q.PerPage != 0 {
		totalPages++
	}
	result := ArticlePage{Items: []ArticleSummary{}, Pagination: Pagination{Page: q.Page, PerPage: q.PerPage, TotalItems: len(matches), TotalPages: totalPages}}
	// Compare before multiplying to avoid integer overflow for arbitrary page values.
	if q.Page > totalPages {
		return result, nil
	}
	start := (q.Page - 1) * q.PerPage
	end := min(start+q.PerPage, len(matches))
	for _, article := range matches[start:end] {
		result.Items = append(result.Items, ArticleSummary{Summary: contentSummary(article.Content), ID: article.ID, Author: article.Author, Title: article.Title, Date: article.Date, Category: article.Category, Tags: append([]string{}, article.Tags...), Draft: article.Draft, URL: "/data/article/" + strconv.Itoa(article.ID) + ".json"})
	}
	return result, nil
}

// Find looks up an article by its repository-local issue number.
func (s *ContentSnapshot) Find(id int) (ContentArticle, bool) {
	for _, article := range s.Articles {
		if article.ID == id {
			article.Tags = append([]string{}, article.Tags...)
			article.Metadata = NewFrontMatter(article.Metadata).Values()
			return article, true
		}
	}
	return ContentArticle{}, false
}

func parseContentDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
		if date, err := time.Parse(layout, value); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid article date %q", value)
}

func sortArticles(articles []ContentArticle, field, order string) {
	sort.Slice(articles, func(i, j int) bool {
		a, b := articles[i], articles[j]
		comparison := 0
		if field == "title" {
			comparison = strings.Compare(a.Title, b.Title)
		} else {
			da, _ := parseContentDate(a.Date)
			db, _ := parseContentDate(b.Date)
			comparison = da.Compare(db)
		}
		if comparison == 0 {
			if order == "asc" {
				return a.ID < b.ID
			}
			return a.ID > b.ID
		}
		if order == "asc" {
			return comparison < 0
		}
		return comparison > 0
	})
}
