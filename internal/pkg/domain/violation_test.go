package domain

import "testing"

func TestViolationStruct(t *testing.T) {
	v := Violation{
		File:    "foo.rb",
		Line:    42,
		Column:  7,
		Message: "msg",
	}
	if v.File != "foo.rb" || v.Line != 42 || v.Column != 7 || v.Message != "msg" {
		t.Errorf("unexpected violation: %+v", v)
	}
}
