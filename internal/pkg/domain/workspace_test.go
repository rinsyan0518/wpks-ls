package domain

import "testing"

func TestNewWorkspace(t *testing.T) {
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
			c := NewWorkspace(tt.rootUri, tt.rootPath)
			if c.RootUri != tt.rootUri || c.RootPath != tt.rootPath {
				t.Errorf("unexpected workspace: want %+v, got %+v", tt, c)
			}
		})
	}
}

func TestWorkspace_StripRootUri(t *testing.T) {
	w := NewWorkspace("file:///root", "/root")
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
			got := w.StripRootUri(tt.uri)
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestWorkspace_BuildFileUri(t *testing.T) {
	w := NewWorkspace("file:///root", "/root")
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
			got := w.BuildFileUri(tt.filePath)
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}
