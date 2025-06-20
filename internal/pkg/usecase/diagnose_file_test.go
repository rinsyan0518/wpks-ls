package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type fakePackwerkRunner struct {
	output string
}

func (f *fakePackwerkRunner) RunCheck(rootPath, path string) (*domain.CheckResult, error) {
	return domain.NewCheckResult(f.output), nil
}

func (f *fakePackwerkRunner) RunCheckAll(rootPath string) (*domain.CheckResult, error) {
	return domain.NewCheckResult(f.output), nil
}

func TestDiagnoseFile_Diagnose(t *testing.T) {
	fullMessage := strings.Join([]string{
		"Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'.",
		"Are we missing an abstraction?",
		"Is the code making the reference, and the referenced constant, in the right packages?",
	}, "\n")

	tests := []struct {
		name         string
		fixtureFile  string
		wantCount    int
		wantMessages []string
	}{
		{
			name:        "multiple violations",
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   2,
			wantMessages: []string{
				fullMessage,
				fullMessage,
			},
		},
		{
			name:         "empty output",
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: nil,
		},
	}

	repo := inmemory.NewConfigurationRepository()
	repo.Save(domain.NewConfiguration("file:///root", "/root"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "test", tt.fixtureFile)
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}
			diagnoser := NewDiagnoseFile(repo, &fakePackwerkRunner{output: string(data)})
			diagnostics, err := diagnoser.Diagnose("file:///root/lib/sample.rb")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(diagnostics) != tt.wantCount {
				t.Fatalf("expected %d diagnostics, got %d", tt.wantCount, len(diagnostics))
			}
			for i, wantMsg := range tt.wantMessages {
				gotMsg := diagnostics[i].Message
				if strings.TrimRight(gotMsg, "\n") != strings.TrimRight(wantMsg, "\n") {
					t.Errorf("diagnostic %d: want message\n--- want ---\n%q\n--- got ---\n%q", i, wantMsg, gotMsg)
				}
			}
		})
	}
}
