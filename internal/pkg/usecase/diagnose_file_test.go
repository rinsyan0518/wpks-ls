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
	}, " ")

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
	repo.Save(domain.NewConfiguration("file:///root", "/root", false))

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

func TestDiagnoseFile_DiagnoseAll(t *testing.T) {
	fullMessage := strings.Join([]string{
		"Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'.",
		"Are we missing an abstraction?",
		"Is the code making the reference, and the referenced constant, in the right packages?",
	}, " ")

	tests := []struct {
		name                  string
		fixtureFile           string
		checkAllOnInitialized bool
		wantCount             int
		wantMessages          map[string][]string
	}{
		{
			name:                  "multiple violations with checkAllOnInitialized true",
			fixtureFile:           "packwerk_output_multiple.txt",
			checkAllOnInitialized: true,
			wantCount:             2,
			wantMessages: map[string][]string{
				"file:///root/packs/users/app/controllers/users_controller.rb": {
					fullMessage,
					fullMessage,
				},
			},
		},
		{
			name:                  "multiple violations with checkAllOnInitialized false",
			fixtureFile:           "packwerk_output_multiple.txt",
			checkAllOnInitialized: false,
			wantCount:             0,
			wantMessages:          nil,
		},
		{
			name:                  "empty output",
			fixtureFile:           "packwerk_output_empty.txt",
			checkAllOnInitialized: true,
			wantCount:             0,
			wantMessages:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := inmemory.NewConfigurationRepository()
			repo.Save(domain.NewConfiguration("file:///root", "/root", tt.checkAllOnInitialized))

			fixturePath := filepath.Join("..", "..", "..", "test", tt.fixtureFile)
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}
			diagnoser := NewDiagnoseFile(repo, &fakePackwerkRunner{output: string(data)})
			diagnosticsByFile, err := diagnoser.DiagnoseAll()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			totalDiagnostics := 0
			for _, diagnostics := range diagnosticsByFile {
				totalDiagnostics += len(diagnostics)
			}
			if totalDiagnostics != tt.wantCount {
				t.Fatalf("expected %d diagnostics, got %d", tt.wantCount, totalDiagnostics)
			}

			for uri, wantMsgs := range tt.wantMessages {
				gotDiagnostics, ok := diagnosticsByFile[uri]
				if !ok {
					t.Fatalf("expected diagnostics for URI %s, but got none", uri)
				}
				if len(gotDiagnostics) != len(wantMsgs) {
					t.Fatalf("for URI %s, expected %d diagnostics, got %d", uri, len(wantMsgs), len(gotDiagnostics))
				}
				for i, wantMsg := range wantMsgs {
					if gotDiagnostics[i].Message != wantMsg {
						t.Errorf("diagnostic %d for URI %s: want message %q, got %q", i, uri, wantMsg, gotDiagnostics[i].Message)
					}
				}
			}
		})
	}
}
