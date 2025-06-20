package inmemory

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

func TestConfigurationRepository_SaveAndGet(t *testing.T) {
	repo := NewConfigurationRepository()
	conf := domain.NewConfiguration("file:///root", "/root", false)
	err := repo.Save(conf)
	if err != nil {
		t.Fatalf("unexpected error on save: %v", err)
	}
	got, err := repo.GetConfiguration()
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if got != conf {
		t.Errorf("expected same configuration pointer")
	}
}

func TestConfigurationRepository_GetConfiguration_NotFound(t *testing.T) {
	repo := NewConfigurationRepository()
	_, err := repo.GetConfiguration()
	if err == nil {
		t.Error("expected error when configuration not found")
	}
}
