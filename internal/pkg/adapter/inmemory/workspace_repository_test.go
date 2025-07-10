package inmemory

import (
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

func TestWorkspaceRepository(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*WorkspaceRepository)
		shouldError bool
		check       func(*testing.T, *domain.Workspace, error)
	}{
		{
			name: "SaveAndGet",
			setup: func(r *WorkspaceRepository) {
				w := domain.NewWorkspace("file:///root", "/root")
				err := r.Save(w)
				if err != nil {
					t.Fatalf("failed to save workspace: %v", err)
				}
			},
			shouldError: false,
			check: func(t *testing.T, w *domain.Workspace, err error) {
				if w == nil {
					t.Fatal("expected workspace, got nil")
				}
				if w.RootUri != "file:///root" {
					t.Errorf("unexpected RootUri: want %q, got %q", "file:///root", w.RootUri)
				}
			},
		},
		{
			name: "GetWorkspace_NotFound",
			setup: func(r *WorkspaceRepository) {
				// No setup needed
			},
			shouldError: true,
			check: func(t *testing.T, w *domain.Workspace, err error) {
				if w != nil {
					t.Error("expected nil workspace when not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewWorkspaceRepository()
			tt.setup(repo)
			w, err := repo.GetWorkspace()
			if (err != nil) != tt.shouldError {
				t.Fatalf("unexpected error status: want error=%v, got err=%v", tt.shouldError, err)
			}
			tt.check(t, w, err)
		})
	}
}
