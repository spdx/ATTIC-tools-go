package tag

import "testing"

func TestIsMultilineValue(t *testing.T) {
	tests := map[string]bool{
		"some string":  false,
		"some\nstring": true,
		"\n":           true,
	}
	for test, expected := range tests {
		result := isMultilineValue(test)
		if result != expected {
			t.Errorf("Incorrect. Test: %#v. Result is %#v but expected %#v", test, result, expected)
		}
	}
}

func TestIsMultiline(t *testing.T) {
	tests := map[string]bool{
		"FileName":       false,
		"LicenseComment": true,
		"PackageSummary": true,
		"SPDXVersion":    false,
	}
	for test, expected := range tests {
		result := isMultiline(test)
		if result != expected {
			t.Errorf("Incorrect. Test: %#v. Result is %#v but expected %#v", test, result, expected)
		}
	}
}
