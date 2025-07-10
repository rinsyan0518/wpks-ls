package usecase

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
)

func TestCreateWorkspace_Create(t *testing.T) {
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
			repo := inmemory.NewWorkspaceRepository()
			uc := NewCreateWorkspace(repo)
			err := uc.Create(tt.rootUri, tt.rootPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			conf, err := repo.GetWorkspace()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if conf.RootUri != tt.rootUri || conf.RootPath != tt.rootPath {
				t.Errorf("unexpected workspace: want %+v, got %+v", tt, conf)
			}
		})
	}
}
