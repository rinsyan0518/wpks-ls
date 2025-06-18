package usecase

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
)

func TestConfigure_Configure(t *testing.T) {
	repo := inmemory.NewConfigurationRepository()
	uc := NewConfigure(repo)
	err := uc.Configure("file:///root", "/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	conf, err := repo.GetConfiguration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.RootUri != "file:///root" || conf.RootPath != "/root" {
		t.Errorf("unexpected configuration: %+v", conf)
	}
}
