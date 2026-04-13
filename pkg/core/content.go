package core

import (
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

// Article represents one generated content entry.
type Article struct {
	Author      string      `yaml:"author"`
	Title       string      `yaml:"title"`
	Content     string      `yaml:"-"`
	Date        string      `yaml:"date"`
	Category    string      `yaml:"categories"`
	Tags        []string    `yaml:"tags"`
	Draft       bool        `yaml:"draft"`
	FrontMatter FrontMatter `yaml:"-"`
	Key         string      `yaml:"-"`
	Images      []*Image    `yaml:"-"`
}

// FrontMatter stores normalized metadata values.
type FrontMatter struct {
	values map[string]any
}

// Image represents one image reference found in the issue body.
type Image struct {
	URL  string
	Time string
	ID   int
}

// ImageAsset is a streamed remote asset payload.
type ImageAsset struct {
	Body        io.ReadCloser
	ContentType string
}

func NewFrontMatter(values map[string]any) FrontMatter {
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneFrontMatterValue(value)
	}
	return FrontMatter{values: cloned}
}

func EmptyFrontMatter() FrontMatter {
	return FrontMatter{values: map[string]any{}}
}

func NewImage(url, time string, id int) *Image {
	return &Image{
		URL:  url,
		Time: time,
		ID:   id,
	}
}

func (a *Article) ParseDateTime() (time.Time, error) {
	if t, err := time.Parse("2006-01-02T15:04:05Z", a.Date); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", a.Date); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("failed to parse date %s", a.Date)
}

func (a *Article) Clone() *Article {
	cloned := *a
	if a.Tags != nil {
		cloned.Tags = append([]string(nil), a.Tags...)
	}
	if a.Images != nil {
		cloned.Images = append([]*Image(nil), a.Images...)
	}
	cloned.FrontMatter = NewFrontMatter(a.FrontMatter.Values())
	return &cloned
}

func (fm FrontMatter) Values() map[string]any {
	cloned := make(map[string]any, len(fm.values))
	for key, value := range fm.values {
		cloned[key] = cloneFrontMatterValue(value)
	}
	return cloned
}

func (fm FrontMatter) MarshalYAML() (data []byte, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			data = nil
			err = fmt.Errorf("failed to marshal front matter: %v", recovered)
		}
	}()
	return yaml.Marshal(fm.values)
}

func (fm FrontMatter) IsEmpty() bool {
	return len(fm.values) == 0
}

func cloneFrontMatterValue(value any) any {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		cloned := make([]any, len(typed))
		for i, item := range typed {
			cloned[i] = cloneFrontMatterValue(item)
		}
		return cloned
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, item := range typed {
			cloned[key] = cloneFrontMatterValue(item)
		}
		return cloned
	default:
		return typed
	}
}
