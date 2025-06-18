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

	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	if violations[0].File != "lib/foo.rb" || violations[0].Line != 10 || violations[0].Column != 5 || violations[0].Message != "Some violation message here" {
		t.Errorf("unexpected first violation: %+v", violations[0])
	}
	if violations[1].File != "lib/bar.rb" || violations[1].Line != 20 || violations[1].Column != 3 || violations[1].Message != "Another violation message" {
		t.Errorf("unexpected second violation: %+v", violations[1])
	}
}
