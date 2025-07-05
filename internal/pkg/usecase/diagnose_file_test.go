package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/inmemory"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/adapter/packwerk"
	"github.com/rinsyan0518/wpks-ls/internal/pkg/domain"
)

type fakePackwerkRunner struct {
	output string
}

func (f *fakePackwerkRunner) RunCheck(ctx context.Context, rootPath string, paths ...string) ([]domain.Violation, error) {
	violations := packwerk.NewPackwerkOutput(f.output).Parse()

	// If there are no violations or no paths, return empty
	if len(violations) == 0 || len(paths) == 0 {
		return []domain.Violation{}, nil
	}

	// Simulate violations for each path by duplicating the violations
	var allViolations []domain.Violation
	for _, path := range paths {
		for _, v := range violations {
			// Create a copy of the violation with the specific path
			violation := v
			violation.File = path
			allViolations = append(allViolations, violation)
		}
	}

	return allViolations, nil
}

func (f *fakePackwerkRunner) RunCheckAll(ctx context.Context, rootPath string) ([]domain.Violation, error) {
	return packwerk.NewPackwerkOutput(f.output).Parse(), nil
}

func TestDiagnoseFile_Diagnose(t *testing.T) {
	fullMessage := strings.Join([]string{
		"Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'.",
		"Are we missing an abstraction?",
		"Is the code making the reference, and the referenced constant, in the right packages?",
	}, " ")

	tests := []struct {
		name         string
		uris         []string
		fixtureFile  string
		wantCount    int
		wantMessages map[string][]string
	}{
		{
			name:         "empty URIs",
			uris:         []string{},
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: map[string][]string{},
		},
		{
			name:        "single URI with violations",
			uris:        []string{"file:///root/lib/sample.rb"},
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   2,
			wantMessages: map[string][]string{
				"file:///root/lib/sample.rb": {
					fullMessage,
					fullMessage,
				},
			},
		},
		{
			name:        "multiple URIs with violations",
			uris:        []string{"file:///root/lib/sample.rb", "file:///root/lib/another.rb"},
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   4, // 2 violations per URI
			wantMessages: map[string][]string{
				"file:///root/lib/sample.rb": {
					fullMessage,
					fullMessage,
				},
				"file:///root/lib/another.rb": {
					fullMessage,
					fullMessage,
				},
			},
		},
		{
			name:         "multiple URIs with empty output",
			uris:         []string{"file:///root/lib/sample.rb", "file:///root/lib/another.rb"},
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: map[string][]string{},
		},
	}

	repo := inmemory.NewConfigurationRepository()
	err := repo.Save(domain.NewConfiguration("file:///root", "/root", false))
	if err != nil {
		t.Fatalf("failed to save configuration: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "test", tt.fixtureFile)
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}
			diagnoser := NewDiagnoseFile(repo, &fakePackwerkRunner{output: string(data)})
			diagnosticsByFile, err := diagnoser.Diagnose(context.Background(), tt.uris...)
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
					gotMsg := gotDiagnostics[i].Message
					if strings.TrimRight(gotMsg, "\n") != strings.TrimRight(wantMsg, "\n") {
						t.Errorf("diagnostic %d for URI %s: want message\n--- want ---\n%q\n--- got ---\n%q", i, uri, wantMsg, gotMsg)
					}
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
			err := repo.Save(domain.NewConfiguration("file:///root", "/root", tt.checkAllOnInitialized))
			if err != nil {
				t.Fatalf("failed to save configuration: %v", err)
			}

			fixturePath := filepath.Join("..", "..", "..", "test", tt.fixtureFile)
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}
			diagnoser := NewDiagnoseFile(repo, &fakePackwerkRunner{output: string(data)})
			diagnosticsByFile, err := diagnoser.DiagnoseAll(context.Background())
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
