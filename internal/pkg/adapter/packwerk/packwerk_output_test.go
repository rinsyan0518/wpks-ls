package packwerk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackwerkOutput_Parse(t *testing.T) {
	type expectedViolation struct {
		File      string
		Line      uint32
		Character uint32
		Type      string
		Message   string
	}

	tests := []struct {
		name               string
		fixtureFile        string
		expectedViolations []expectedViolation
	}{
		{
			name:        "single violation",
			fixtureFile: "check_result_fixture.txt",
			expectedViolations: []expectedViolation{
				{
					File:      "packs/users/app/controllers/users_controller.rb",
					Line:      20,
					Character: 4,
					Type:      "Dependency violation",
					Message:   "Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'. Are we missing an abstraction? Is the code making the reference, and the referenced constant, in the right packages?",
				},
			},
		},
		{
			name:        "multiple violations",
			fixtureFile: "packwerk_output_multiple.txt",
			expectedViolations: []expectedViolation{
				{
					File:      "packs/users/app/controllers/users_controller.rb",
					Line:      20,
					Character: 4,
					Type:      "Dependency violation",
					Message:   "Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'. Are we missing an abstraction? Is the code making the reference, and the referenced constant, in the right packages?",
				},
				{
					File:      "packs/users/app/controllers/users_controller.rb",
					Line:      26,
					Character: 4,
					Type:      "Dependency violation",
					Message:   "Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'. Are we missing an abstraction? Is the code making the reference, and the referenced constant, in the right packages?",
				},
			},
		},
		{
			name:               "empty file",
			fixtureFile:        "packwerk_output_empty.txt",
			expectedViolations: []expectedViolation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "..", "test", tt.fixtureFile)
			data, err := os.ReadFile(fixturePath)
			if err != nil {
				t.Fatalf("failed to read fixture: %v", err)
			}

			result := NewPackwerkOutput(string(data))
			violations := result.Parse()

			if len(violations) != len(tt.expectedViolations) {
				t.Fatalf("expected %d violations, got %d", len(tt.expectedViolations), len(violations))
			}

			for i, v := range violations {
				ev := tt.expectedViolations[i]
				if v.File != ev.File {
					t.Errorf("violation %d: unexpected file: got %q, want %q", i, v.File, ev.File)
				}
				if v.Line != ev.Line {
					t.Errorf("violation %d: unexpected line: got %d, want %d", i, v.Line, ev.Line)
				}
				if v.Character != ev.Character {
					t.Errorf("violation %d: unexpected character: got %d, want %d", i, v.Character, ev.Character)
				}
				if v.Type != ev.Type {
					t.Errorf("violation %d: unexpected type: got %q, want %q", i, v.Type, ev.Type)
				}
				if v.Message != ev.Message {
					t.Errorf("violation %d: unexpected message:\n--- got ---\n%q\n--- want ---\n%q", i, v.Message, ev.Message)
				}
			}
		})
	}
}
