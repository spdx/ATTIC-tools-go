package rdf

import "testing"

// Test goraptor term to string
func TestTermStr(t *testing.T) {
	val := "some value"
	if val != termStr(literal(val)) {
		t.Fail()
	}
	if val != termStr(uri(val)) {
		t.Fail()
	}
	if val != termStr(blank(val)) {
		t.Fail()
	}
}
