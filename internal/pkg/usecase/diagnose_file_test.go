package usecase

import (
	"os"
	"path/filepath"
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

func TestDiagnoseFile_Diagnose(t *testing.T) {
	repo := inmemory.NewConfigurationRepository()
	repo.Save(domain.NewConfiguration("file:///root", "/root"))

	fixturePath := filepath.Join("..", "..", "..", "test", "packwerk_output_fixture.txt")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	diagnoser := NewDiagnoseFile(repo, &fakePackwerkRunner{output: string(data)})
	diagnostics, err := diagnoser.Diagnose("file:///root/lib/sample.rb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}
	d := diagnostics[0]
	if d.Range.Start.Line != 2 || d.Range.Start.Character != 2 || d.Range.End.Line != 3 || d.Range.End.Character != 0 || d.Message != "Sample violation from diagnose file" {
		t.Errorf("unexpected diagnostic: %+v", d)
	}
}

func TestDiagnoseFile_Diagnose_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		fixtureFile  string
		wantCount    int
		wantMessages []string
	}{
		{
			name:        "multiple violations",
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   3,
			wantMessages: []string{
				"First violation",
				"Second violation",
				"Third violation",
			},
		},
		{
			name:         "empty output",
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: nil,
		},
		{
			name:         "malformed output",
			fixtureFile:  "packwerk_output_malformed.txt",
			wantCount:    1,
			wantMessages: []string{"But this one is valid"},
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
			for i, msg := range tt.wantMessages {
				if diagnostics[i].Message != msg {
					t.Errorf("diagnostic %d: want message %q, got %q", i, msg, diagnostics[i].Message)
				}
			}
		})
	}
}
