package convert

import "github.com/rokuosan/github-issue-cms/internal/api"

type Converter[T interface{}] interface {
	Convert() ([]*T, error)
}

type IssueToArticleConverter struct {
	GitHub api.GitHub
}

func (c *IssueToArticleConverter) Convert() ([]*Article, error) {
	var articles []*Article
	issues := c.GitHub.GetIssues()
	for _, iss := range issues {
		article, err := (&issue{iss}).ConvertToArticle()
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, nil
}
