package domain

import "testing"

func TestNewConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		rootUri  string
		rootPath string
	}{
		{"true case", "file:///root", "/root"},
		{"false case", "file:///another/root", "/another/root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfiguration(tt.rootUri, tt.rootPath)
			if c.RootUri != tt.rootUri || c.RootPath != tt.rootPath {
				t.Errorf("unexpected configuration: want %+v, got %+v", tt, c)
			}
		})
	}
}

func TestConfiguration_StripRootUri(t *testing.T) {
	c := NewConfiguration("file:///root", "/root")
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{"normal case", "file:///root/foo/bar.rb", "foo/bar.rb"},
		{"edge case with trailing slash", "file:///root/", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.StripRootUri(tt.uri)
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestConfiguration_BuildFileUri(t *testing.T) {
	c := NewConfiguration("file:///root", "/root")
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"normal case", "foo/bar.rb", "file:///root/foo/bar.rb"},
		{"empty path", "", "file:///root/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.BuildFileUri(tt.filePath)
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}
