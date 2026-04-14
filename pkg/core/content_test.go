package core

import (
	"testing"
	"time"
)

func TestNewImage(t *testing.T) {
	tests := []struct {
		name string
		url  string
		time string
		id   int
		want *Image
	}{
		{
			name: "正常なImage作成",
			url:  "https://example.com/image.png",
			time: "2021-01-01_120000",
			id:   0,
			want: &Image{
				URL:  "https://example.com/image.png",
				Time: "2021-01-01_120000",
				ID:   0,
			},
		},
		{
			name: "IDが正の値",
			url:  "https://example.com/another.png",
			time: "2021-12-31_235959",
			id:   42,
			want: &Image{
				URL:  "https://example.com/another.png",
				Time: "2021-12-31_235959",
				ID:   42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewImage(tt.url, tt.time, tt.id)
			assertEqualCmp(t, *tt.want, *got)
		})
	}
}

func TestArticle_ParseDateTime(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "ISO8601 format",
			date:    "2021-01-01T00:00:00Z",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "date only format",
			date:    "2021-01-01",
			want:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			date:    "invalid-date",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := &Article{Date: tt.date}
			got, err := article.ParseDateTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assertEqualCmp(t, tt.want, got)
			}
		})
	}
}
