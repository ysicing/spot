package cmd

import "testing"

func TestCanDeleteImage(t *testing.T) {
	tests := []struct {
		imageType string
		want      bool
	}{
		{imageType: "官方", want: false},
		{imageType: "自定义镜像", want: true},
		{imageType: "共享镜像", want: true},
		{imageType: "未知", want: false},
	}

	for _, tt := range tests {
		if got := canDeleteImage(tt.imageType); got != tt.want {
			t.Errorf("canDeleteImage(%q) = %t, want %t", tt.imageType, got, tt.want)
		}
	}
}
