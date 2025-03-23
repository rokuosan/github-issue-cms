package ghapi

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v67/github"
	"github.com/h2non/gock"
)

func strPtr(s string) *string {
	return &s
}

func Test_IssuesByRepo(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/golang/go/issues").
		Reply(200).
		JSON([]*github.Issue{
			{Title: strPtr("title1")},
			{Title: strPtr("title2")},
		})

	t.Run("Success", func(t *testing.T) {
		client := NewClient(github.NewClient(nil))
		res, err := client.Issues(context.Background(), IssuesInput{
			Owner: "golang",
			Name:  "go",
		})
		if err != nil {
			t.Error(err)
		}
		want := []Issue{
			{Title: strPtr("title1")},
			{Title: strPtr("title2")},
		}
		if diff := cmp.Diff(want, res); diff != "" {
			t.Errorf("IssuesByRepo() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		client := NewClient(github.NewClient(nil))
		_, err := client.Issues(ctx, IssuesInput{
			Owner: "golang",
			Name:  "go",
		})
		if err == nil {
			t.Error("IssuesByRepo() got nil, want context.Canceled")
		}
		if err != context.Canceled {
			t.Errorf("IssuesByRepo() got %v, want %v", err, context.Canceled)
		}
	})
}
