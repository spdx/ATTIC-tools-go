package spdx

import "testing"

// Testing Validator.SpecVersion

func TestSpecVersion(t *testing.T) {
	val := Str("SPDX-1.2", nil)
	validator := new(Validator)
	validator.SpecVersion(&val)

	if !validator.Ok() || validator.Major != 1 || validator.Minor != 2 {
		t.Fail()
	}
}

func TestSpecVersionWarning(t *testing.T) {
	val := Str("spdx-1.2", nil)
	validator := new(Validator)
	validator.SpecVersion(&val)

	if !validator.HasWarnings() || validator.Major != 1 || validator.Minor != 2 {
		t.Error("Failed to parse \"spdx-1.2\".")
	}

	val = Str("1.2", nil)
	validator = new(Validator)
	validator.SpecVersion(&val)

	if !validator.HasWarnings() || validator.Major != 1 || validator.Minor != 2 {
		t.Error("Failed to parse \"1.2\".")
	}

	val = Str("spdx1.2", nil)
	validator = new(Validator)
	validator.SpecVersion(&val)

	if !validator.HasWarnings() || validator.Major != 1 || validator.Minor != 2 {
		t.Error("Failed to parse \"spdx1.2\".")
	}
}

func TestSpecVersionError(t *testing.T) {
	val := Str("spdx-1", nil)
	validator := new(Validator)
	validator.SpecVersion(&val)

	if !validator.HasErrors() {
		t.Error("Didn't fail at \"spdx-1\".")
	}
}
