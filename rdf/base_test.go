package rdf

import (
	"testing"
)

func TestFormatOk(t *testing.T) {
	if !FormatOk(Fmt_turtle) {
		t.Error("Not accepted a good format.")
	}
	if FormatOk("this is not a valid format") {
		t.Error("Accepted a good format")
	}
}

func TestPrefix(t *testing.T) {
	tests := map[string]string{
		"ns:something": "http://www.w3.org/1999/02/22-rdf-syntax-ns#something",
		"ns:":          "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"":             baseUri,
	}
	for short, long := range tests {

		res := prefix(short)
		if termStr(res) != long {
			t.Errorf("Found: %#v (expected %#v)", termStr(res), long)
		}
	}
}

func TestShortPrefix(t *testing.T) {
	tests := map[string]string{
		"ns:something": "http://www.w3.org/1999/02/22-rdf-syntax-ns#something",
		"ns:":          "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
		"":             baseUri,
	}
	for short, long := range tests {
		res := shortPrefix(uri(long))
		if res != short {
			t.Errorf("Found: %#v (expected %#v)", res, short)
		}
	}
}
