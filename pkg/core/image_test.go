package core

import "testing"

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
