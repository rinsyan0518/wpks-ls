package usecase

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
)

func TestConfigure_Configure(t *testing.T) {
	tests := []struct {
		name                  string
		rootUri               string
		rootPath              string
		checkAllOnInitialized bool
	}{
		{
			name:                  "with checkAllOnInitialized true",
			rootUri:               "file:///root",
			rootPath:              "/root",
			checkAllOnInitialized: true,
		},
		{
			name:                  "with checkAllOnInitialized false",
			rootUri:               "file:///another/root",
			rootPath:              "/another/root",
			checkAllOnInitialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewConfigurationRepository()
			uc := NewConfigure(repo)
			err := uc.Configure(tt.rootUri, tt.rootPath, tt.checkAllOnInitialized)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			conf, err := repo.GetConfiguration()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if conf.RootUri != tt.rootUri || conf.RootPath != tt.rootPath || conf.CheckAllOnInitialized != tt.checkAllOnInitialized {
				t.Errorf("unexpected configuration: want %+v, got %+v", tt, conf)
			}
		})
	}
}
