package spdx

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	ValidWarning = iota
	ValidError   = iota
)

// Validation error. Holds the property name and metadata (line numbers) of errors.
type ValidationError struct {
	msg  string
	Type int
	*Meta
}

// Make *ValidationError implement the error interface.
func (err *ValidationError) Error() string { return err.msg }

func NewVError(msg string, m *Meta) *ValidationError   { return &ValidationError{msg, ValidError, m} }
func NewVWarning(msg string, m *Meta) *ValidationError { return &ValidationError{msg, ValidWarning, m} }

// Check if val matches any of the items in correct. Return whether they have the same
// case or only a case-insensitive match was found.
func correctCaseMatch(val string, correct []string) (caseSensitive bool, index int) {
	for i, s := range correct {
		if s == val {
			return true, i
		}
	}
	for i, s := range correct {
		if strings.ToLower(s) == strings.ToLower(val) {
			return false, i
		}
	}
	return false, -1
}

type Validator struct {
	Major    int // Version major
	Minor    int // Verison Minor
	LicMajor int // Licence List Major
	LicMinor int // Licence List Minor
	errs     []*ValidationError
}

func (v *Validator) Errors() []*ValidationError  { return v.errs }
func (v *Validator) addErr(msg string, m *Meta)  { v.add(NewVError(msg, m)) }
func (v *Validator) addWarn(msg string, m *Meta) { v.add(NewVWarning(msg, m)) }
func (v *Validator) Ok() bool                    { return len(v.errs) == 0 }
func (v *Validator) HasWarnings() bool           { return v.hasErrType(ValidWarning) }
func (v *Validator) HasErrors() bool             { return v.hasErrType(ValidError) }

func (v *Validator) hasErrType(t int) bool {
	for _, e := range v.errs {
		if e.Type == t {
			return true
		}
	}
	return false
}

func (v *Validator) add(err ...*ValidationError) {
	if len(err) == 0 {
		return
	}
	if v.errs == nil {
		v.errs = err
		return
	}
	v.errs = append(v.errs, err...)
}

// Validate a *Document
func (v *Validator) Document(doc *Document) bool {
	if v.SpecVersion(&doc.SpecVersion) {
		v.VersionSupported(doc.SpecVersion.Meta)
	}
	v.DataLicence(&doc.DataLicence)

	// validate creation info
	if doc.CreationInfo != nil {
		creators := 0
		var meta *Meta
		for _, cr := range doc.CreationInfo.Creator {
			meta = cr.Meta
			if v.DocumentCreator(&cr) {
				creators++
			}
		}
		if creators == 0 {
			v.addErr("At least one valid creator is required.", meta)
		}

		// Creation date
		v.Date(&doc.CreationInfo.Created)

		// LicenceListVersion
		if llv := doc.CreationInfo.LicenceListVersion; llv.V() != "" {
			if _, err := fmt.Sscanf(llv.V(), "%d.%d", &v.LicMajor, &v.LicMinor); err != nil {
				v.addErr("Invalid format for LicenceListVersion.", llv.Meta)
			}
		}
	} else {
		v.addErr("No creation info found. Creation date and at least one creator are mandatory.", nil)
	}

	return len(v.errs) == 0
}

// Validate SpecVersion. Updates v.Major and v.Minor.
//
// Valid option: SPDX-M.m (M major version, m minor version)
// Warning on: (any case "SPDX"): spdx-M.m, SPDXM.m, M.m
// Error on anything else.
func (v *Validator) SpecVersion(val *ValueStr) bool {
	if _, err := fmt.Sscanf(val.Val, "SPDX-%d.%d", &v.Major, &v.Minor); err == nil {
		return true
	}

	reg := regexp.MustCompile("(?i)spdx-?")
	ver := reg.ReplaceAllLiteralString(val.Val, "")

	if _, err := fmt.Sscanf(ver, "%d.%d", &v.Major, &v.Minor); err == nil {
		v.addWarn(fmt.Sprintf("SpecVersion was parsed to SPDX-%d.%d but it is in an invalid format.", v.Major, v.Minor), val.Meta)
		return true
	}
	v.addErr("Invalid SpecVersion format.", val.Meta)
	return false
}

// Check if the version this validator has is currently supported by this library.
// Please keep SpecVersions updated in spdx/base.go.
func (v *Validator) VersionSupported(m *Meta) bool {
	ver := [2]int{v.Major, v.Minor}
	for _, sv := range SpecVersions {
		if sv == ver {
			return true
		}
	}
	v.addErr(fmt.Sprintf("SPDX Specification version SPDX-%d.%d is not supported by this version of spdx-go.", v.Major, v.Minor), m)
	return false
}

// Validate Data Licence
func (v *Validator) DataLicence(val *ValueStr) bool {
	if val.Val == "CC0-1.0" {
		return true
	}
	if strings.ToUpper(val.Val) == "CC0-1.0" {
		v.addWarn("Data License should be exactly 'CC0-1.0' (uppercase CC).", val.Meta)
		return true
	}
	v.addErr("Invalid Data License. Must be 'CC0-1.0'.", val.Meta)
	return false
}

// Validate DocumentCreator. It returns whether the checked value is valid or not.
func (v *Validator) DocumentCreator(val *ValueCreator) bool {
	what, name, email := val.What(), val.Name(), val.Email()

	if what == "" || name == "" {
		v.addErr("Creator does not have the correct syntax: \"what: name (email)\"", val.Meta)
		return false
	}

	caseSensitive, match := correctCaseMatch(what, []string{"Organization", "Tool", "Person"})
	if match < 0 {
		v.addErr(fmt.Sprintf("Creator of type \"%s\" is not valid. Valid options: \"Organization\", \"Tool\" and \"Person\"", what), val.Meta)
		return false
	}

	if !caseSensitive {
		v.addWarn(fmt.Sprintf("Incorrect or no capitalization in \"%s\".", what), val.Meta)
	}

	if match == 1 && email != "" {
		v.addWarn("Tools should not have e-mail addresses.", val.Meta)
	}

	return true
}

// Validate dates.
func (v *Validator) Date(val *ValueDate) bool {
	if val.Time() == nil {
		v.addErr("Invalid date format.", val.Meta)
		return false
	}
	return true
}
