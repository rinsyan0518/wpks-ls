package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckResult_Parse(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "..", "test", "check_result_fixture.txt")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	result := NewCheckResult(string(data))
	violations := result.Parse()

	expectedFile := "packs/users/app/controllers/users_controller.rb"
	expectedLine := uint32(20)
	expectedChar := uint32(4)
	expectedType := "Dependency violation"
	expectedMessage := strings.Join([]string{
		"Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'.",
		"Are we missing an abstraction?",
		"Is the code making the reference, and the referenced constant, in the right packages?",
	}, " ")

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	v := violations[0]

	if v.File != expectedFile {
		t.Errorf("unexpected file: got %q, want %q", v.File, expectedFile)
	}
	if v.Line != expectedLine {
		t.Errorf("unexpected line: got %d, want %d", v.Line, expectedLine)
	}
	if v.Character != expectedChar {
		t.Errorf("unexpected character: got %d, want %d", v.Character, expectedChar)
	}
	if v.Type != expectedType {
		t.Errorf("unexpected type: got %q, want %q", v.Type, expectedType)
	}
	if v.Message != expectedMessage {
		t.Errorf("unexpected message:\n--- got ---\n%q\n--- want ---\n%q", v.Message, expectedMessage)
	}
}
