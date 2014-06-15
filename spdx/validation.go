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

type Validator struct {
	Major int // Version major
	Minor int // Verison Minor
	errs  []*ValidationError
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
func (v *Validator) Document(doc *Document) {
	v.SpecVersion(&doc.SpecVersion)
	v.DataLicence(&doc.DataLicence)
}

// Validate SpecVersion. Updates v.Major and v.Minor.
//
// Valid option: SPDX-M.m (M major version, m minor version)
// Warning on: (any case "SPDX"): spdx-M.m, SPDXM.m, M.m
// Error on anything else.
func (v *Validator) SpecVersion(val *ValueStr) {
	if _, err := fmt.Sscanf(val.Val, "SPDX-%d.%d", &v.Major, &v.Minor); err == nil {
		v.VersionSupported(val.Meta)
		return
	}

	reg := regexp.MustCompile("(?i)spdx-?")
	ver := reg.ReplaceAllLiteralString(val.Val, "")

	if _, err := fmt.Sscanf(ver, "%d.%d", &v.Major, &v.Minor); err == nil {
		v.addWarn(fmt.Sprintf("SpecVersion was parsed to SPDX-%d.%d but it is in an invalid format.", v.Major, v.Minor), val.Meta)
		v.VersionSupported(val.Meta)

		return
	}
	v.addErr("Invalid SpecVersion format.", val.Meta)
}

// Check if the version this validator has is currently supported by this library.
// Please keep SpecVersions updated in spdx/base.go.
func (v *Validator) VersionSupported(m *Meta) {
	ver := [2]int{v.Major, v.Minor}
	ok := false
	for _, sv := range SpecVersions {
		if sv == ver {
			ok = true
			break
		}
	}
	if !ok {
		v.addErr(fmt.Sprintf("SPDX Specification version SPDX-%d.%d is not supported by this version of spdx-go.", v.Major, v.Minor), m)
	}
}

// Validate Data Licence
func (v *Validator) DataLicence(val *ValueStr) {
	if val.Val == "CC0-1.0" {
		return
	}
	if strings.ToUpper(val.Val) == "CC0-1.0" {
		v.addWarn("Data License should be exactly 'CC0-1.0' (uppercase CC).", val.Meta)
		return
	}
	v.addErr("Invalid Data License. Must be 'CC0-1.0'.", val.Meta)
}
