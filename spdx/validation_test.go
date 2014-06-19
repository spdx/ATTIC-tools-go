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

// Single line of text
func TestSingleLineErrors(t *testing.T) {
	val := Str("This is a multi-line\n value", nil)
	validator := new(Validator)
	validator.SingleLineErr(val, "err")
	validator.SingleLineWarn(val, "warn")
	if !validator.HasWarnings() {
		t.Error("No warnings.")
	}
	if !validator.HasErrors() {
		t.Error("No errors.")
	}
}

func TestSingleLineOK(t *testing.T) {
	val := Str("This is a single-line value.", nil)
	validator := new(Validator)
	validator.SingleLineErr(val, "err")
	validator.SingleLineWarn(val, "warn")
	if validator.HasWarnings() {
		t.Error("Unexpected warnings.")
	}
	if validator.HasErrors() {
		t.Error("Unexpected errors.")
	}
}

// Mandatory text
func TestMandatoryText(t *testing.T) {
	val := Str("", nil)
	validator := new(Validator)
	validator.MandatoryText(val, false, false, "a")
	if !validator.HasErrors() {
		t.Error("Empty value shouldn't be permitted.")
	}
}

func TestMandatoryTextErrNONE(t *testing.T) {
	val := Str(NONE, nil)
	validator := new(Validator)
	validator.MandatoryText(val, false, false, "a")
	if !validator.HasErrors() {
		t.Error("NONE value shouldn't be permitted.")
	}
}

func TestMandatoryTextErrNOASSERTION(t *testing.T) {
	val := Str(NOASSERTION, nil)
	validator := new(Validator)
	validator.MandatoryText(val, false, false, "a")
	if !validator.HasErrors() {
		t.Error("NOASSERTION value shouldn't be permitted.")
	}
}

func TestMandatoryTextErrNONEallowedNOASSERTION(t *testing.T) {
	val := Str(NOASSERTION, nil)
	validator := new(Validator)
	validator.MandatoryText(val, false, true, "a")
	if !validator.HasErrors() {
		t.Error("NOASSERTION value shouldn't be permitted.")
	}
}

func TestMandatoryTextNOASSERTION(t *testing.T) {
	val := Str(NOASSERTION, nil)
	validator := new(Validator)
	validator.MandatoryText(val, true, false, "a")
	if validator.HasErrors() {
		t.Error("NOASSERTION value shouldn be permitted.")
	}
}

// Test date
func TestValueDateInvalid(t *testing.T) {
	val := NewValueDate("not a valid format.", nil)
	validator := new(Validator)
	validator.Date(&val)
	if !validator.HasErrors() {
		t.Error("No error.")
	}
}

func TestValueDate(t *testing.T) {
	val := NewValueDate("2014-04-11T12:32:44Z", nil)
	validator := new(Validator)
	validator.Date(&val)
	if validator.HasErrors() {
		t.Error("Unexpected errors.")
	}
}

// Validate URL

func TestUrlInvalid(t *testing.T) {
	val := Str("not an url, obviously", nil)
	validator := new(Validator)
	if validator.Url(&val, false, false, "a") {
		t.Error("No error.")
	}
}

func TestUrl(t *testing.T) {
	val := Str("http://spdx.org/", nil)
	validator := new(Validator)
	if !validator.Url(&val, false, false, "a") {
		t.Fail()
	}
}

// Validate DateLicence
func TestDataLicence(t *testing.T) {
	val := Str("CC0-1.0", nil)
	validator := new(Validator)
	if !validator.DataLicence(&val) {
		t.Fail()
	}
}

func TestDataLicenceWarning(t *testing.T) {
	val := Str("cc0-1.0", nil)
	validator := new(Validator)
	validator.DataLicence(&val)
	if !validator.HasWarnings() {
		t.Fail()
	}
}

func TestDataLicenceError(t *testing.T) {
	val := Str("cc", nil)
	validator := new(Validator)
	if validator.DataLicence(&val) {
		t.Fail()
	}
}

// Creator

func TestCreatorIncorrectSyntax(t *testing.T) {
	val := NewValueCreator("Something Wrong", nil)
	validator := new(Validator)
	if validator.Creator(&val, false, false, "Test", nil) {
		t.Fail()
	}
	if !validator.HasErrors() {
		t.Error("Error was not added to the validator.")
	}
}

func TestCreatorInvalidWhat(t *testing.T) {
	val := NewValueCreator("Human: John", nil)
	validator := new(Validator)
	if validator.Creator(&val, false, false, "Test", []string{"Tool", "Organization"}) {
		t.Fail()
	}
	if !validator.HasErrors() {
		t.Error("Error was not added to the validator.")
	}
}

func TestCreatorIncorrectCase(t *testing.T) {
	val := NewValueCreator("TOOL: fdas", nil)
	validator := new(Validator)
	validator.Creator(&val, false, false, "Test", []string{"Tool", "Organization"})
	if validator.HasErrors() {
		t.Error("Should be a warning.")
	}
	if !validator.HasWarnings() {
		t.Error("Doesn't have warnings.")
	}
}

func TestCreatorNoEmail(t *testing.T) {
	val := NewValueCreator("Tool: fdas (john@example.com)", nil)
	validator := new(Validator)
	validator.Creator(&val, false, false, "Test", []string{"Test", "Tool", "Organization"}, 0, 1)
	if validator.HasErrors() {
		t.Error("Should be a warning.")
	}
	if !validator.HasWarnings() {
		t.Error("Doesn't have warnings.")
	}
}

func TestCreatorOK(t *testing.T) {
	val := NewValueCreator("Organization: fdas (contact@example.com)", nil)
	validator := new(Validator)
	if !validator.Creator(&val, false, false, "Test", []string{"Test", "Tool", "Organization"}, 0, 1) {
		t.Error("Should've returned true")
	}
	if validator.HasErrors() {
		t.Error("Shouldn't have errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

// Checksum test
func TestChecksumOK(t *testing.T) {
	val := &Checksum{
		Algo:  Str("SHA1", nil),
		Value: Str("2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", nil),
	}
	validator := new(Validator)
	validator.Major = 1
	if !validator.Checksum(val) {
		t.Error("Should return true.")
	}
	if validator.HasErrors() {
		t.Error("Should not have a errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestChecksumWrongLength(t *testing.T) {
	val := &Checksum{
		Algo:  Str("SHA1", nil),
		Value: Str("2fd4e1c67a2d28fced849ee1bb76e7391b9", nil),
	}
	validator := new(Validator)
	validator.Major = 1

	if validator.Checksum(val) {
		t.Error("Should return false.")
	}
	if !validator.HasErrors() {
		t.Error("Should have an error")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestChecksumNotHex(t *testing.T) {
	val := &Checksum{
		Algo:  Str("SHA1", nil),
		Value: Str("2fd4e1c67a2d28fced849ee1bb76e7391b9_xb12", nil),
	}
	validator := new(Validator)
	validator.Major = 1

	if validator.Checksum(val) {
		t.Error("Should return false.")
	}
	if !validator.HasErrors() {
		t.Error("Should have an error")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestChecksumWarning(t *testing.T) {
	val := &Checksum{
		Algo:  Str("MD5", nil),
		Value: Str("2fd4e1c67a2d28fced849ee1bb76e739", nil),
	}
	validator := new(Validator)
	validator.Major = 1

	if !validator.Checksum(val) {
		t.Error("Should return true.")
	}
	if validator.HasErrors() {
		t.Error("Should not have a errors")
	}
	if !validator.HasWarnings() {
		t.Error("Should have warnings.")
	}
}

// Test Verification Code
func TestVerificationCodeOK(t *testing.T) {
	val := &VerificationCode{
		Value: Str("2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", nil),
	}
	validator := new(Validator)
	if !validator.VerificationCode(val) {
		t.Error("Should return true.")
	}
	if validator.HasErrors() {
		t.Error("Should not have a errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestVerificationCodeWrongLength(t *testing.T) {
	val := &VerificationCode{
		Value: Str("2fd4e1c67a2d28f", nil),
	}
	validator := new(Validator)
	if validator.VerificationCode(val) {
		t.Error("Should return false.")
	}
	if !validator.HasErrors() {
		t.Error("Should have errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestVerificationCodeNotHex(t *testing.T) {
	val := &VerificationCode{
		Value: Str("2fd4e1c67a2d28fced849ee1bb76x7391y93eb12", nil),
	}
	validator := new(Validator)
	if validator.VerificationCode(val) {
		t.Error("Should return false.")
	}
	if !validator.HasErrors() {
		t.Error("Should have errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}

func TestVerificationCodeEmptyExcludedFiles(t *testing.T) {
	val := &VerificationCode{
		Value:         Str("2fd4e1c67a2d28fced849ee1bb76c7391393eb12", nil),
		ExcludedFiles: []ValueStr{Str("this_is_fine.txt", nil), Str("", nil)},
	}
	validator := new(Validator)
	if validator.VerificationCode(val) {
		t.Error("Should return false.")
	}
	if !validator.HasErrors() {
		t.Error("Should have errors.")
	}
	if validator.HasWarnings() {
		t.Error("Shouldn't have warnings.")
	}
}
