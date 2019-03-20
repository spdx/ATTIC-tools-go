package spdx

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Error types.
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
func (err *ValidationError) Error() string {
	var prefix string
	if err.Type == ValidError {
		prefix = "ERROR: "
	} else {
		prefix = "WARNING: "
	}
	return prefix + err.msg
}

// Creates a new validation error.
func NewVError(msg string, m *Meta) *ValidationError { return &ValidationError{msg, ValidError, m} }

// Creates a new validation warning.
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

// A Validator is used to validate SPDX Documents and subsets of documents. A
// Validator must be created using the `NewValidator()` function.
//
// Unless the whole document is validated (using `Validator.Document()`), the
// SPDX Version should be set by either:
// - Validating the SpecVersion property of a document by calling
// `Validator.SpecVersion()`
// - or manually setting the values for `Validator.Major` and `Validator.Minor`.
//
// As a convention for validator methods (such as `Validator.Document`,
// `Validator.Creator`), the return value should be `false` if there were errors
// added to the validator; warnings do not count. The return value should be
// `true` if there are no errors added (warnings are allowed). If warnings are
// added, the return value should still be `true`. If a validator method behaves
// differently, it will be clearly documented.
type Validator struct {
	Major    int // Version major
	Minor    int // Verison Minor
	LicMajor int // Licence List Major
	LicMinor int // Licence List Minor

	// Licence references used/defined and where
	licUsed    map[string]*Meta
	licDefined map[string]*Meta

	// Objects already validated
	validated map[interface{}]bool

	// File references
	files map[string]*File

	// Validator errors
	errs []*ValidationError
}

// Creates a new Validator struct.
func NewValidator() *Validator {
	return &Validator{
		licUsed:    make(map[string]*Meta),
		licDefined: make(map[string]*Meta),
		validated:  make(map[interface{}]bool),
	}
}

// Errors returns all the errors and warnings that this validator has.
func (v *Validator) Errors() []*ValidationError { return v.errs }

// Add a new error to this validator.
func (v *Validator) addErr(msg string, m *Meta, args ...interface{}) {
	v.add(NewVError(fmt.Sprintf(msg, args...), m))
}

// Add a new warning to this validator.
func (v *Validator) addWarn(msg string, m *Meta, args ...interface{}) {
	v.add(NewVWarning(fmt.Sprintf(msg, args...), m))
}

// Ok returns whether there are no errors and no warnings.
func (v *Validator) Ok() bool { return len(v.errs) == 0 }

// HasWarnings returns true if there are no warnings in this validator, false otherwise.
func (v *Validator) HasWarnings() bool { return v.hasErrType(ValidWarning) }

// HasErrors returns true if there are no errors in this validator, false otherwise.
func (v *Validator) HasErrors() bool { return v.hasErrType(ValidError) }

// Internal usage. Returns true if there are errors of type t (ValidationError.Type==t) in this validator, false otherwise.
func (v *Validator) hasErrType(t int) bool {
	for _, e := range v.errs {
		if e.Type == t {
			return true
		}
	}
	return false
}

// Adds a list of errors (or warnings) to this validator. Internally used by *Validator.addErr and *Validator.addWarn
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

// SingleLineErr adds an error to this validator if `val.V()` has more than one lines of text.
func (v *Validator) SingleLineErr(val Value, property string) bool {
	if strings.Index(val.V(), "\n") >= 0 {
		v.addErr("%s must be a single line.", val.M(), property)
		return false
	}
	return true
}

// SingleLineWarn adds a warning to this validator if `val.V()` has more than one lines of text.
// Returns `false` if there was a warning added, `true` otherwise.
func (v *Validator) SingleLineWarn(val Value, property string) bool {
	if strings.Index(val.V(), "\n") >= 0 {
		v.addWarn("%s should be a single line.", val.M(), property)
		return false
	}
	return true
}

// MandatoryText adds an error if `val.V()` is empty.
//
// NOASSERTION and NONE values are considered invalid if `noassert` and `none`, respectively are set to false.
// These values are treated as valid (do not generate errors) if the arguments are set to true.
//
// The `property` string is used in the error message.
func (v *Validator) MandatoryText(val Value, noassert, none bool, property string) bool {
	str := val.V()

	if str == "" {
		v.addErr("%s cannot be empty.", val.M(), property)
		return false
	}

	if (!noassert && str == NOASSERTION) || (!none && str == NONE) {
		v.addErr("%s cannot be %s.", val.M(), property, str)
		return false
	}

	return true
}

// Date validates `*ValueDate` values. If `val.Time() == nil` this generates an error and returns `false`.
// It returns `true` otherwise.
func (v *Validator) Date(val *ValueDate) bool {
	if val.Time() == nil {
		v.addErr("Invalid date format.", val.Meta)
		return false
	}
	return true
}

// Url validates URLs. URLs must have a scheme to be valid.
func (v *Validator) Url(val *ValueStr, noassert, none bool, property string) bool {
	if (noassert && val.V() == NOASSERTION) || (none && val.V() == NONE) {
		return true
	}
	if val.V() == "" {
		v.addErr("%s cannot be empty.", val.Meta, property)
		return false
	}
	u, err := url.Parse(val.V())
	if err != nil || u.Scheme == "" {
		v.addErr("%s: Invalid URL.", val.Meta, property)
		return false
	}

	return true
}

// Document validates a *Document. After validating doc, it checks whether all licence
// references are in place (all "LicenceRef-" type licences used inside the
// document and its nested elements are defined in doc.ExtractedLicences).
//
// This method adds the following errors, if found:
// - SPDX Version format is not valid
// - SPDX Version in the document is not currently supported by this tool
// - No valid document creator
// - ExtractedLicence (a licence with ID starting with "LicenceRef") used
//   but not defined within the parsed SPDX file
// - all errors added by the nested elements
//
// This method adds the following warnings, if found:
// - SPDX Version format is lowercase or does not start with "SPDX-"
// - ExtractedLicence defined in this SPDX file but not used in any of the
//   nested elements of this document.
// - all warnings added by nested elements.
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

	// validate packages
	for _, pkg := range doc.Packages {
		v.Package(pkg)
	}

	// In SPDX 1.x, there must be one package per document
	if v.Major == 1 && len(doc.Packages) > 1 {
		v.addErr("A document cannot have more than one package in SPDX-1.x.", doc.Packages[1].Meta)
	} else if v.Major == 1 && len(doc.Packages) == 0 {
		v.addErr("A document must have one Package in SPDX-1.x.", nil)
	}

	for _, file := range doc.Files {
		v.File(file)
	}

	for _, lic := range doc.ExtractedLicences {
		v.ExtractedLicence(lic)
		v.defineLicenceRef(lic.LicenceId(), lic.Id.M())
	}

	for _, rev := range doc.Reviews {
		v.Review(rev)
	}

	v.LicReferences()

	return v.HasErrors()
}

// LicReferences checks if all the licences referenced in the SPDX Document (and indexed by
// the Validator) are used and defined.
//
// Licence References used but not defined generate errors.
// Licence References defined but not used generate warnings.
func (v *Validator) LicReferences() bool {
	r := true
	for k, m := range v.licUsed {
		_, ok := v.licDefined[k]
		if !ok {
			v.addErr("Licence reference \"%s\" used but not defined.", m, k)
			r = false
		} else {
			delete(v.licDefined, k)
		}
	}
	for k, m := range v.licDefined {
		_, ok := v.licUsed[k]
		if !ok {
			v.addWarn("Licence reference \"%s\" defined but not used.", m, k)
		}
	}
	return r
}

// SpecVersion validates SpecVersion. Updates v.Major and v.Minor.
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
	v.addErr("Invalid SpecVersion format. The rest of the validation might be incorrect or incomplete.", val.Meta)
	return false
}

// VersionSupported checks if the SPDX version this validator has is currently supported by this
// library and adds an error if it is not.
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

// DataLicence validates Data Licence. The only valid value is "CC0-1.0".
// A warning is generated for non-uppercase "CC".
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

// DocumentCreator validates DocumentCreator. It returns whether the checked value is valid or not.
func (v *Validator) DocumentCreator(val *ValueCreator) bool {
	return v.Creator(val, false, false, "Document Creator", []string{"Tool", "Organization", "Person"}, 0)
}

// Creator validates *ValueCreator types.
//
// The ValueCreator Syntax is: "What: Who (email)" where the "(email)" is optional.
//
// If noassert (or none) parameter is set to true, NOASSERTION (or NONE) value will be considered valid.
// The property parameter is only used in error/warning messages.
//
// whats parameter is a slice of the values that are considered valid for val.What(),
// which is the part before the first ":" in the ValueCreator syntax. A case-insensitive match
// is performed, but a warning is added if the case in val.What() is different then the one in whats.
//
// noemails is a slice of indexes in the whats slice. For those indexes, an email address is not permitted.
// A warning is added if there is an e-mail address provided.
func (v *Validator) Creator(val *ValueCreator, noassert, none bool, property string, whats []string, noemails ...int) bool {
	if (noassert && val.V() == NOASSERTION) || (none && val.V() == NONE) {
		return true
	}
	if !v.MandatoryText(val, noassert, none, property) {
		return false
	}

	what, name, email := val.What(), val.Name(), val.Email()

	if what == "" || name == "" {
		v.addErr("%s does not have the correct syntax: \"what: name (email)\"", val.Meta, property)
		return false
	}

	caseSensitive, match := correctCaseMatch(what, whats)
	if match < 0 {
		v.addErr("%s of type \"%s\" is not valid. Valid options: %s", val.Meta, property, what, strings.Join(whats, ", "))
		return false
	}

	if !caseSensitive {
		v.addWarn("Incorrect or no capitalization in \"%s\".", val.Meta, what)
	}

	for _, id := range noemails {
		if match == id && email != "" {
			v.addWarn("%s should not have e-mail addresses.", val.Meta, whats[id])
			break
		}
	}

	return true
}

// Review validates a Review
func (v *Validator) Review(rev *Review) bool {
	if rev.Reviewer.val == "" && rev.Date.val == "" {
		return true
	}
	r := rev.Reviewer.val == "" || v.Creator(&rev.Reviewer, false, false, "Reviewer", []string{"Person", "Organization", "Tool"}, 2)
	return v.Date(&rev.Date) && r
}

// Package validates a Package.
//
// Adds the following errors, if found:
// - Package name is empty or on multiple lines.
// - Package version is on multiple lines.
// - Package File Name is on multiple lines.
// - Package Supplier or Package Originator are not in a valid "creator" format:
//      What: Name (Email)
//   Valid options for "What" are: "Person" and "Organization".
// - Package download location is not a valid URL
// - Invalid Package Verification Code
// - Invalid Package Checksum
// - Package home page is not a valid URL
// - No licence concluded defined
// - No licence declared defined
// - Empty licence info from files
// - Wrong values for any of the licences
// - all errors from the nested files
func (v *Validator) Package(pkg *Package) bool {
	if cache, ok := v.validated[pkg]; ok {
		return cache
	}
	r := v.MandatoryText(pkg.Name, false, false, "Package Name")
	r = v.SingleLineErr(pkg.Name, "Package Name") && r

	r = v.SingleLineErr(pkg.Version, "Package Version") && r
	r = v.SingleLineErr(pkg.FileName, "Package File Name") && r

	r = (pkg.Supplier.V() == "" || v.Creator(&pkg.Supplier, true, false, "Package Supplier", []string{"Person", "Organization"})) && r
	r = (pkg.Originator.V() == "" || v.Creator(&pkg.Originator, true, false, "Package Originator", []string{"Person", "Organization"})) && r

	r = v.Url(&pkg.DownloadLocation, true, true, "Package Download Location") && r

	r = v.VerificationCode(pkg.VerificationCode) && r
	r = (pkg.Checksum == nil || (pkg.Checksum.Value.V() == "" && pkg.Checksum.Algo.V() == "") || v.Checksum(pkg.Checksum)) && r

	r = (pkg.HomePage.V() == "" || v.Url(&pkg.HomePage, true, true, "Package Home Page")) && r
	r = v.MandatoryText(&pkg.CopyrightText, true, true, "Package Copyright Text") && r

	if pkg.LicenceConcluded == nil {
		v.addErr("Package Licence Concluded cannot be empty.", pkg.Meta)
		r = false
	} else {
		r = v.AnyLicenceOptionals(pkg.LicenceConcluded, true, true, true, "Package Licence Concluded") && r
	}

	if pkg.LicenceDeclared == nil {
		v.addErr("Package Licence Declared cannot be empty.", pkg.Meta)
		r = false
	} else {
		r = v.AnyLicenceOptionals(pkg.LicenceDeclared, true, true, true, "Package Licence Declared") && r
	}

	for _, lic := range pkg.LicenceInfoFromFiles {
		if lic == nil {
			v.addErr("Package Licence Info from Files cannot be empty.", pkg.Meta)
			r = false
		} else {
			r = v.AnyLicenceOptionals(lic, false, true, true, "Licence Info From File") && r
		}
	}

	for _, file := range pkg.Files {
		r = v.File(file) && r
	}
	v.validated[pkg] = r
	return r
}

// File validates File.
//
// Adds the following errors, if found:
// - Empty file name
// - Same file defined twice (indexed by name)
// - File name spans on multiple lines
// - Invalid file type for the SPDX Version used
// - Invalid checksum (and errors added by file checksum validation)
// - Empty file contributor
// - File contributor spans on multiple lines
// - Empty copyright text
// - Errors from nested ArtifactOf elements
// - Errors from nested licence elements
// - Empty licence concluded or licence info in file
// - Errors from validating the file dependencies
//
// Adds the following warnings, if found:
// - Valid file type, but in incorrect case (eg. "binary", not "BINARY")
// - Warnings added by file checksum validation.
// - Warnings from nested licence elements.
// - Warnings from nested ArtifactOf elements.
// - Warnings from validating file dependencies
func (v *Validator) File(f *File) bool {
	if cache, ok := v.validated[f]; ok {
		return cache
	}
	r := v.MandatoryText(&f.Name, false, false, "File Name")

	// file indexing
	if r {
		if v.files == nil {
			v.files = make(map[string]*File)
		}
		_f, ok := v.files[f.Name.Val]
		if ok {
			if f != _f {
				// file name already defined
				if m := _f.Meta; m != nil {
					v.addErr("File already defined at line %d.", f.Meta, _f.Meta.LineStart)
				} else {
					v.addErr("File already defiend.", f.Meta)
				}
				r = false
			} else {
				// already validated, just return true and skip
				return true
			}
		} else {
			v.files[f.Name.Val] = f
		}
	}

	r = v.SingleLineErr(&f.Name, "File Name") && r

	if f.Type.Val != "" {
		var fileTypes []string
		if v.Major == 1 {
			fileTypes = []string{FT_BINARY, FT_SOURCE, FT_ARCHIVE, FT_OTHER}
		} else if v.Major == 2 {
			fileTypes = []string{FT_BINARY, FT_SOURCE, FT_ARCHIVE, FT_OTHER, FT_AUDIO, FT_VIDEO, FT_APPLICATION, FT_TEXT, FT_IMAGE}
		}

		ci, index := correctCaseMatch(f.Type.Val, fileTypes)
		if index < 0 {
			v.addErr("Incorrect File Type %s. Permitted values for SPDX-%d.%d are: %s.", f.Type.Meta, f.Type.Val, v.Major, v.Minor, strings.Join(fileTypes, ", "))
			r = false
		} else if ci == false {
			v.addWarn("Incorrect File Type case %s. Correct value is '%s'.", f.Type.Meta, f.Type.Val, fileTypes[index])
		}
	}
	r = f.Checksum != nil && v.Checksum(f.Checksum) && r
	if f.LicenceConcluded == nil {
		v.addErr("File Licence Concluded cannot be empty.", f.Meta)
		r = false
	} else {
		r = v.AnyLicenceOptionals(f.LicenceConcluded, true, true, true, "File Licence Concluded") && r
	}
	for _, lic := range f.LicenceInfoInFile {
		if lic == nil {
			v.addErr("Licence Info In File cannot be empty.", f.Meta)
			r = false
		} else {
			r = v.AnyLicenceOptionals(lic, false, true, true, "Licence Info in File") && r
		}
	}

	r = v.MandatoryText(&f.CopyrightText, true, true, "File Copyright Text") && r

	for _, file := range f.Dependency {
		r = v.File(file) && r
	}

	for _, contrib := range f.Contributor {
		r = v.MandatoryText(contrib, false, false, "File Contributor") &&
			v.SingleLineErr(contrib, "File Contributor") && r
	}

	// ArtifactOf
	for _, artif := range f.ArtifactOf {
		r = v.ArtifactOf(artif) && r
	}

	v.validated[f] = r
	return r
}

// ArtifactOf validates ArtifactOf.
//
// Adds an error if:
// - The artifact is empty (nil pointer or empty values)
// - ProjectUri is not a valid URI
// - HomePage is neither UNKNOWN or a valid URL
func (v *Validator) ArtifactOf(a *ArtifactOf) bool {
	if a == nil {
		v.addErr("No Artifact defined.", nil)
		return false
	}
	if cache, ok := v.validated[a]; ok {
		return cache
	}
	notEmpty := a.Name.Val != "" || a.ProjectUri.Val != "" || (a.HomePage.Val != "" && a.HomePage.Val != "UNKNOWN")
	if !notEmpty {
		v.addErr("Artifact is empty.", a.Meta)
		return false
	}
	r := v.Url(&a.ProjectUri, false, false, "Artifact Project URI")
	r = (a.HomePage.Val == "UNKNOWN" || v.Url(&a.HomePage, false, false, "Artifact Home Page")) && r
	v.validated[a] = r
	return r
}

// Package Verification Code validation
//
// Adds an error if:
// - the value is not exactly 40 hexadecimal digits
// - any of the ExcludedFiles slice elements is empty
func (v *Validator) VerificationCode(vc *VerificationCode) bool {
	if vc == nil {
		v.addErr("Package Verification Code is mandatory.", nil)
		return false
	}
	if cache, ok := v.validated[vc]; ok {
		return cache
	}
	r := true
	if len(vc.Value.V()) != 40 || !isHex(vc.Value.V()) {
		v.addErr("Package Verification Code value must be exactly 40 lowercase hexadecimal digits.", vc.Meta)
		r = false
	}

	for _, e := range vc.ExcludedFiles {
		r = v.MandatoryText(e, false, false, "Package Verification Code Excluded File") && r
	}

	v.validated[vc] = r
	return r
}

// In spec verison SPDX-1.x the recommended algorithm is SHA1. If other algorithm is used, a warning is generated.
//
// Adds an error if:
// - Algorithm or Value are empty
// - For common algorithms, if the vlaue has the incorrect length
func (v *Validator) Checksum(cksum *Checksum) bool {
	if cache, ok := v.validated[cksum]; ok {
		return cache
	}
	if !v.MandatoryText(cksum.Algo, false, false, "Checksum Algorithm") || !v.MandatoryText(cksum.Value, false, false, "Checksum Value") {
		return false
	}

	// some algorithms and hex output length
	algos := map[string]int{
		"MD5":     32,
		"SHA1":    40,
		"SHA256":  64,
		"SHA-256": 64,
		"SHA512":  128,
		"SHA-512": 128,
		"SHA384":  96,
		"SHA-384": 96,
	}

	if v.Major == 1 && cksum.Algo.V() != "SHA1" {
		v.addWarn("The checksum algorithm recommeded for SPDX-1.x is SHA1 but now using %s.", cksum.Meta, cksum.Algo.V())
	}

	if l, ok := algos[cksum.Algo.V()]; ok && (len(cksum.Value.V()) != l || !isHex(cksum.Value.V())) {
		v.validated[cksum] = false
		v.addErr("Checksum value for algorithm %s must be hexadecimal of length %d.", cksum.Meta, cksum.Algo.V(), l)
		return false
	}
	v.validated[cksum] = true
	return true
}

// Returns whether `val` is a hexadecimal value.
func isHex(val string) bool {
	b, _ := regexp.MatchString("^[a-f0-9]*$", val)
	return b
}

// Adds `id` as a defined Licence for this validator. Creates a warning if the validator already has this Licence ID.
func (v *Validator) defineLicenceRef(id string, m *Meta) {
	if v.licDefined == nil {
		v.licDefined = make(map[string]*Meta)
		v.licDefined[id] = m
		return
	}
	at, ok := v.licDefined[id]
	if ok {
		if at != nil {
			v.addWarn("Licence %s already defined at lines %d to %d.", m, id, at.LineStart, at.LineEnd)
		} else {
			v.addWarn("Licence %s already defined.", m, id)
		}
	}
	v.licDefined[id] = m
}

// Adds 'id' as a used Licence for this validator.
func (v *Validator) useLicence(id string, m *Meta) {
	if v.licUsed == nil {
		v.licUsed = make(map[string]*Meta)
	}
	v.licUsed[id] = m
}

// AnyLicenceOptionals validates an AnyLicence object, treating NONE and NOASSERTION.
func (v *Validator) AnyLicenceOptionals(lic AnyLicence, allowSets, none, noassert bool, property string) bool {
	t, ok := lic.(Licence)
	if ok && ((none && t.V() == NONE) || (noassert && t.V() == NOASSERTION)) {
		return true
	}
	return v.AnyLicence(lic, allowSets, property)
}

// Licences.
//
// Adds errors if:
// - Licence found and the ID is neither in the SPDX Licence List or a valid "LicenceRef-" ID.
// - Licence Set found but not sets are allowed.
// - Unknown licence type is found (something else than Licence, ExtractedLicence,
//   DisjunctiveLicenceSet or ConjunctiveLicenceSet).
// - any validation errors from validating ExtractedLicence, if the case
func (v *Validator) AnyLicence(lic AnyLicence, allowSets bool, property string) bool {
	switch t := lic.(type) {
	case Licence:
		if isLicIdRef(t.LicenceId()) {
			v.LicenceRefId(t.LicenceId(), t.M(), property)
			v.useLicence(t.LicenceId(), t.M())
			return true
		}
		if !CheckLicence(t.V()) {
			v.addErr("%s: Licence Reference not in SPDX Licence List and not a custom licence reference.", t.M(), t.V())
			return false
		}
		return true
	case ConjunctiveLicenceSet:
		if !allowSets {
			v.addErr("%s: Sets are not allowed but found a Conjunctive Licence Set.", t.M(), property)
			return false
		}
		r := true
		for _, l := range t.Members {
			r = v.AnyLicence(l, true, property) && r
		}
		return r
	case DisjunctiveLicenceSet:
		if !allowSets {
			v.addErr("%s: Sets are not allowed but found a Disjunctive Licence Set.", t.M(), property)
			return false
		}
		r := true
		for _, l := range t.Members {
			r = v.AnyLicence(l, true, property) && r
		}
		return r
	case *ExtractedLicence:
		v.useLicence(t.LicenceId(), t.M())
		return v.ExtractedLicence(t)
	default:
		var m *Meta
		if lic != nil {
			m = lic.M()
		}
		v.addErr("%s: Unknown Licence type.", m, property)
		return false
	}
}

// Raise warning if invalid characters are used in LicenseRef ID.
// Returns `false` if a warnings is created, `true` otherwise.
func (v *Validator) LicenceRefId(id string, meta *Meta, property string) bool {
	validChars := "a-z A-Z 0-9 + - ."
	var ok bool
	if v.Major > 1 || (v.Major == 1 && v.Minor >= 2) {
		ok, _ = regexp.MatchString("^LicenseRef-[a-zA-Z0-9+\\.-]+$", id)
	} else {
		// only numbers allowed in spdx < 1.2
		ok, _ = regexp.MatchString("^LicenseRef-[0-9]+$", id)
		validChars = "0-9"
	}
	if ok {
		return true
	}
	v.addWarn("%s: Licence ID Reference has unsupported characters. Valid characters for SPDX-%d.%d are: %s", meta, property, v.Major, v.Minor, validChars)
	return false
}

// ExtractedLicence validates ExtractedLicence object
//
// Adds errors if:
// - ID is not a valid Licence Reference format ("LicenseRef-...")
// - No names are defined
// - No Cross-References (reference URIs) are defined.
// - Any of the names is empty or spans on multiple lines.
// - Any of the Cross-References is not a valid URI.
func (v *Validator) ExtractedLicence(lic *ExtractedLicence) bool {
	if cache, ok := v.validated[lic]; ok {
		return cache
	}
	r := true
	if !isLicIdRef(lic.Id.V()) {
		r = false
		v.addErr("Not a valid licence reference format.", lic.Id.M())
	} else {
		v.LicenceRefId(lic.Id.V(), lic.Id.M(), "Extracted Licence ID")
	}

	if len(lic.Name) == 0 {
		r = false
		v.addErr("Licences not in the SPDX Licence List must have at least one name defined.", lic.M())
	}

	if len(lic.CrossReference) == 0 {
		r = false
		v.addErr("Licences not in the SPDX Licence List must have at least one reference URI.", lic.M())
	}

	for _, name := range lic.Name {
		r = v.MandatoryText(name, false, false, "Extracted Licence Name") && r
		r = v.SingleLineErr(name, "Extracted Licence Name") && r
	}

	for _, url := range lic.CrossReference {
		r = v.Url(&url, false, false, "Extracted Licence Cross Reference") && r
	}
	if v.validated == nil {
		v.validated = make(map[interface{}]bool)
	}
	v.validated[lic] = r
	return r
}
