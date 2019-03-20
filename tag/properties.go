package tag

import "strings"

// Map indexed by the properties an SPDX document has, in their correct case.
var properties map[string]interface{}

// Map indexed by the properties an SPDX Document has lowercased.
// The values of this map is the valid property, in the correct case.
var propertiesLower map[string]string

// Initialise properties and propertiesLower maps.
func initProperties() {
	if properties != nil {
		return
	}
	props := []string{
		"SPDXVersion",
		"DataLicense",
		"DocumentComment",
		"Creator",
		"Created",
		"CreatorComment",
		"LicenseListVersion",
		"PackageName",
		"PackageVersion",
		"PackageFileName",
		"PackageSupplier",
		"PackageOriginator",
		"PackageDownloadLocation",
		"PackageVerificationCode",
		"PackageChecksum",
		"PackageHomePage",
		"PackageSourceInfo",
		"PackageLicenseConcluded",
		"PackageLicenseInfoFromFiles",
		"PackageLicenseDeclared",
		"PackageLicenseComments",
		"PackageCopyrightText",
		"PackageSummary",
		"PackageDescription",
		"FileName",
		"FileType",
		"FileChecksum",
		"LicenseConcluded",
		"LicenseInfoInFile",
		"LicenseComments",
		"FileCopyrightText",
		"FileComment",
		"FileNotice",
		"FileContributor",
		"FileDependency",
		"ArtifactOfProjectName",
		"ArtifactOfProjectHomePage",
		"ArtifactOfProjectURI",
		"LicenseID",
		"ExtractedText",
		"LicenseName",
		"LicenseCrossReference",
		"LicenseComment",
		"Reviewer",
		"ReviewDate",
		"ReviewComment",
	}

	properties = make(map[string]interface{})
	propertiesLower = make(map[string]string)
	for _, p := range props {
		properties[p] = nil
		propertiesLower[strings.ToLower(p)] = p
	}
}

// IsValidProperty checks whether the property is valid, case sensitive
func IsValidProperty(prop string) bool {
	initProperties()
	_, ok := properties[prop]
	return ok
}

// IsValidPropertyInsensitive returns whether the given property is valid in a case-insensitive manner
// along with the Correct Case for that property.
func IsValidPropertyInsensitive(prop string) (bool, string) {
	initProperties()
	correct, ok := propertiesLower[strings.ToLower(prop)]
	return ok, correct
}

// Returns whether the property given is defined as a multiline property by the
// SPDX Specification.
func isMultiline(property string) bool {
	_, ok := multilineProperties[property]
	return ok
}

// Returns whether the value given has more lines (string contains '\n').
func isMultilineValue(val string) bool {
	return strings.Index(val, "\n") >= 0
}

// Map storing the properties that are multiline in the SPDX Specification.
var multilineProperties = multilineInit()

// Initialise multilineProperties package variable
func multilineInit() map[string]interface{} {
	tags := []string{
		"DocumentComment",
		"CreatorComment",
		"LicenseComment",
		"LicenseComments",
		"ReviewComment",

		"FileComment",
		"FileNotice",
		"FileCopyrightText",

		"PackageLicenseComments",
		"PackageCopyrightText",
		"PackageSummary",
		"PackageDescription",

		"ExtractedText",
		"PackageSourceInfo",
	}

	mps := make(map[string]interface{})

	for _, tag := range tags {
		mps[tag] = nil
	}

	return mps
}
