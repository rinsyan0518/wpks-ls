package domain

import (
	"os"
	"path/filepath"
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

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if violations[0].File != "packs/users/app/controllers/users_controller.rb" ||
		violations[0].Line != 20 ||
		violations[0].Character != 4 ||
		violations[0].Message != "Dependency violation: ::Book belongs to 'packs/books', but 'packs/users' does not specify a dependency on 'packs/books'." ||
		violations[0].Type != "Dependency violation" {
		t.Errorf("unexpected violation: %+v", violations[0])
	}
}
