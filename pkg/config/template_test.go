package config

import (
	"testing"
	"time"
)

func Test_compile(t *testing.T) {
	type args struct {
		datetime time.Time
		template string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test compile",
			args: args{
				datetime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				template: "%Y-%m-%d",
			},
			want: "2021-01-01",
		},
		{
			name: "Compile with filepath string",
			args: args{
				datetime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				template: "content/posts/%Y/%m/%d",
			},
			want: "content/posts/2021/01/01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compile(tt.args.datetime, tt.args.template); got != tt.want {
				t.Errorf("compile() = %v, want %v", got, tt.want)
			}
		})
	}
}
