package domain

import "testing"

func TestDiagnosticConstants(t *testing.T) {
	if SeverityError != 1 || SeverityWarning != 2 || SeverityInfo != 3 || SeverityHint != 4 {
		t.Error("diagnostic severity constants are incorrect")
	}
}

func TestDiagnosticStruct(t *testing.T) {
	d := Diagnostic{
		Range: Range{
			Start: Position{Line: 1, Character: 2},
			End:   Position{Line: 3, Character: 4},
		},
		Severity: SeverityWarning,
		Source:   "test",
		Message:  "msg",
	}
	if d.Range.Start.Line != 1 || d.Range.End.Character != 4 || d.Severity != 2 || d.Source != "test" || d.Message != "msg" {
		t.Errorf("unexpected diagnostic: %+v", d)
	}
}
