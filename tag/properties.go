package tag

import "strings"

var properties map[string]interface{}
var propertiesLower map[string]string

func initProperties() {
	if properties != nil {
		return
	}
	props := []string{
		"SpecVersion",
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

// Checks whether the property is valid, case sensitive
func IsValidProperty(prop string) bool {
	initProperties()
	_, ok := properties[prop]
	return ok
}

// Returns whether the given property is valid in a case-insensitive manner
// along with the Correct Case for that property.
func IsValidPropertyInsensitive(prop string) (bool, string) {
	initProperties()
	correct, ok := propertiesLower[strings.ToLower(prop)]
	return ok, correct
}
