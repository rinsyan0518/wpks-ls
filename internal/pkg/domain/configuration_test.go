package domain

import "testing"

func TestNewConfiguration(t *testing.T) {
	c := NewConfiguration("file:///root", "/root")
	if c.RootUri != "file:///root" || c.RootPath != "/root" {
		t.Errorf("unexpected configuration: %+v", c)
	}
}

func TestConfiguration_StripRootUri(t *testing.T) {
	c := NewConfiguration("file:///root", "/root")
	uri := "file:///root/foo/bar.rb"
	stripped := c.StripRootUri(uri)
	if stripped != "foo/bar.rb" {
		t.Errorf("expected 'foo/bar.rb', got '%s'", stripped)
	}

	// Edge: no trailing slash in RootUri
	c2 := NewConfiguration("file:///root", "/root")
	uri2 := "file:///root/"
	stripped2 := c2.StripRootUri(uri2)
	if stripped2 != "" {
		t.Errorf("expected '', got '%s'", stripped2)
	}
}
