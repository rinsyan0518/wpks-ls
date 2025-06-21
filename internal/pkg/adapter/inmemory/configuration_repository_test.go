package inmemory

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

func TestConfigurationRepository(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*ConfigurationRepository)
		shouldError bool
		check       func(*testing.T, *domain.Configuration, error)
	}{
		{
			name: "SaveAndGet",
			setup: func(r *ConfigurationRepository) {
				conf := domain.NewConfiguration("file:///root", "/root", false)
				err := r.Save(conf)
				if err != nil {
					t.Fatalf("failed to save configuration: %v", err)
				}
			},
			shouldError: false,
			check: func(t *testing.T, conf *domain.Configuration, err error) {
				if conf == nil {
					t.Fatal("expected configuration, got nil")
				}
				if conf.RootUri != "file:///root" {
					t.Errorf("unexpected RootUri: want %q, got %q", "file:///root", conf.RootUri)
				}
			},
		},
		{
			name: "GetConfiguration_NotFound",
			setup: func(r *ConfigurationRepository) {
				// No setup needed
			},
			shouldError: true,
			check: func(t *testing.T, conf *domain.Configuration, err error) {
				if conf != nil {
					t.Error("expected nil configuration when not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewConfigurationRepository()
			tt.setup(repo)
			conf, err := repo.GetConfiguration()
			if (err != nil) != tt.shouldError {
				t.Fatalf("unexpected error status: want error=%v, got err=%v", tt.shouldError, err)
			}
			tt.check(t, conf, err)
		})
	}
}
