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

// Test constants
const (
	testRootURI  = "file:///root"
	testRootPath = "/root"

	// Expected violation message from packwerk output
	expectedViolationMessage = "Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'. Are we missing an abstraction? Is the code making the reference, and the referenced constant, in the right packages?"

	// Test URIs
	testURI1        = "file:///root/lib/sample.rb"
	testURI2        = "file:///root/lib/another.rb"
	expectedFileURI = "file:///root/packs/users/app/controllers/users_controller.rb"
)

// fakePackwerkRunner is a mock implementation for testing
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

// Test helper functions

// setupTestRepository creates and configures a test repository
func setupTestRepository(t *testing.T) *inmemory.WorkspaceRepository {
	t.Helper()
	repo := inmemory.NewWorkspaceRepository()
	err := repo.Save(domain.NewWorkspace(testRootURI, testRootPath))
	if err != nil {
		t.Fatalf("failed to save workspace: %v", err)
	}
	return repo
}

// loadTestFixture reads and returns the content of a test fixture file
func loadTestFixture(t *testing.T, filename string) string {
	t.Helper()
	fixturePath := filepath.Join("./testdata", filename)
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	return string(data)
}

// createDiagnoser creates a DiagnoseFile instance with test workspace
func createDiagnoser(t *testing.T, fixtureFile string) *DiagnoseFile {
	t.Helper()
	repo := setupTestRepository(t)
	output := loadTestFixture(t, fixtureFile)
	return NewDiagnoseFile(repo, &fakePackwerkRunner{output: output})
}

// assertTotalDiagnosticCount checks if the total number of diagnostics matches expected count
func assertTotalDiagnosticCount(t *testing.T, diagnosticsByFile map[string][]domain.Diagnostic, expected int) {
	t.Helper()
	total := 0
	for _, diagnostics := range diagnosticsByFile {
		total += len(diagnostics)
	}
	if total != expected {
		t.Errorf("expected %d total diagnostics, got %d", expected, total)
	}
}

// assertDiagnosticsForURI checks diagnostics for a specific URI
func assertDiagnosticsForURI(t *testing.T, diagnosticsByFile map[string][]domain.Diagnostic, uri string, expectedMessages []string) {
	t.Helper()

	diagnostics, exists := diagnosticsByFile[uri]
	if !exists {
		t.Fatalf("expected diagnostics for URI %s, but got none", uri)
	}

	if len(diagnostics) != len(expectedMessages) {
		t.Fatalf("for URI %s, expected %d diagnostics, got %d", uri, len(expectedMessages), len(diagnostics))
	}

	for i, expectedMsg := range expectedMessages {
		gotMsg := diagnostics[i].Message
		// Normalize line endings for comparison
		if strings.TrimRight(gotMsg, "\n") != strings.TrimRight(expectedMsg, "\n") {
			t.Errorf("diagnostic %d for URI %s:\nwant: %q\ngot:  %q", i, uri, expectedMsg, gotMsg)
		}
	}
}

// Test cases

func TestDiagnoseFile_Diagnose(t *testing.T) {
	tests := []struct {
		name         string
		uris         []string
		fixtureFile  string
		wantCount    int
		wantMessages map[string][]string
	}{
		{
			name:         "empty URIs returns no diagnostics",
			uris:         []string{},
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: map[string][]string{},
		},
		{
			name:        "single URI with violations",
			uris:        []string{testURI1},
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   2,
			wantMessages: map[string][]string{
				testURI1: {expectedViolationMessage, expectedViolationMessage},
			},
		},
		{
			name:        "multiple URIs with violations",
			uris:        []string{testURI1, testURI2},
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   4, // 2 violations per URI
			wantMessages: map[string][]string{
				testURI1: {expectedViolationMessage, expectedViolationMessage},
				testURI2: {expectedViolationMessage, expectedViolationMessage},
			},
		},
		{
			name:         "multiple URIs with empty output",
			uris:         []string{testURI1, testURI2},
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnoser := createDiagnoser(t, tt.fixtureFile)

			diagnosticsByFile, err := diagnoser.Diagnose(context.Background(), tt.uris...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertTotalDiagnosticCount(t, diagnosticsByFile, tt.wantCount)

			for uri, expectedMessages := range tt.wantMessages {
				assertDiagnosticsForURI(t, diagnosticsByFile, uri, expectedMessages)
			}
		})
	}
}

func TestDiagnoseFile_DiagnoseAll(t *testing.T) {
	tests := []struct {
		name         string
		fixtureFile  string
		wantCount    int
		wantMessages map[string][]string
	}{
		{
			name:        "multiple violations",
			fixtureFile: "packwerk_output_multiple.txt",
			wantCount:   2,
			wantMessages: map[string][]string{
				expectedFileURI: {expectedViolationMessage, expectedViolationMessage},
			},
		},
		{
			name:         "empty output returns no diagnostics",
			fixtureFile:  "packwerk_output_empty.txt",
			wantCount:    0,
			wantMessages: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnoser := createDiagnoser(t, tt.fixtureFile)

			diagnosticsByFile, err := diagnoser.DiagnoseAll(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertTotalDiagnosticCount(t, diagnosticsByFile, tt.wantCount)

			for uri, expectedMessages := range tt.wantMessages {
				assertDiagnosticsForURI(t, diagnosticsByFile, uri, expectedMessages)
			}
		})
	}
}
