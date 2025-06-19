package domain

import "testing"

func TestViolationStruct(t *testing.T) {
	v := Violation{
		File:      "foo.rb",
		Line:      42,
		Character: 7,
		Message:   "msg",
		Type:      "Dependency violation",
	}
	if v.File != "foo.rb" || v.Line != 42 || v.Character != 7 || v.Message != "msg" || v.Type != "Dependency violation" {
		t.Errorf("unexpected violation: %+v", v)
	}
}
