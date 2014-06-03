package tag

import "../spdx"

import (
	"errors"
	"strings"
)

// Updater
type updater (func(string) error)
type updaterMapping map[string]updater

// Update a string pointer
func upd(ptr *string) updater {
	return func(val string) error {
		*ptr = val
		return nil
	}
}

// Update a string slice pointer
func updList(arr *[]string) updater {
	return func(val string) error {
		*arr = append(*arr, val)
		return nil
	}
}

// Update a VerificationCode pointer
func verifCode(vc *spdx.VerificationCode) updater {
	return func(val string) error {
		vc.Value = val
		open := strings.Index(val, "(")
		if open > 0 {
			vc.Value = val[:open]
			start := open + len("excludes:")

			if len(val) <= start {
				return errors.New("Invalid VerificationCode excludes format.")
			}

			vc.ExcludedFiles = strings.Split(val[start:len(val)-1], ",")
			for i, _ := range vc.ExcludedFiles {
				vc.ExcludedFiles[i] = strings.TrimSpace(vc.ExcludedFiles[i])
			}
		}
		return nil
	}
}

// Update a Checksum pointer
func checksum(cksum *spdx.Checksum) updater {
	return func(val string) error {
		split := strings.Split(val, ":")
		if len(split) != 2 {
			return errors.New("Invalid Package Checksum format.")
		}
		cksum.Algo, cksum.Value = strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
		return nil
	}
}

// Update a AnyLicenceInfo pointer
func anyLicence(lic *spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

// Update a []anyLicenceInfo pointer
func anyLicenceList(licList *[]spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

// Creates a file that only has the FileName and appends it to the initially given pointer
func updFileNameList(fl *[]*spdx.File) updater {
	return func(val string) error {
		file := &spdx.File{Name: val}
		*fl = append(*fl, file)
		return nil
	}
}

// Apply the relevant updater function if the given pair matches any.
func applyMapping(input Pair, mapping updaterMapping) (ok bool, err error) {
	f, ok := mapping[input.Key]
	if !ok {
		return false, nil
	}
	return true, f(input.Value)
}

// Document mapping.
func documentMap(doc *spdx.Document) updaterMapping {
	if doc.CreationInfo == nil {
		doc.CreationInfo = new(spdx.CreationInfo)
		doc.CreationInfo.Creator = make([]string, 0)
	}

	if doc.Packages == nil {
		doc.Packages = make([]*spdx.Package, 0)
	}

	if doc.Files == nil {
		doc.Files = make([]*spdx.File, 0)
	}

	if doc.ExtractedLicenceInfo == nil {
		doc.extractedLicenceInfo = make([]*ExtractedLicensingInfo, 0)
	}

	return map[string]updater{
		"SpecVersion":        upd(&doc.SpecVersion),
		"DataLicense":        upd(&doc.DataLicence),
		"DocumentComment":    upd(&doc.Comment),
		"Creator":            updList(&doc.CreationInfo.Creator),
		"Created":            upd(&doc.CreationInfo.Created),
		"CreatorComment":     upd(&doc.CreationInfo.Comment),
		"LicenseListVersion": upd(&doc.CreationInfo.LicenceListVersion),
	}
}

// Package mapping.
func packageMap(pkg *spdx.Package) updaterMapping {
	if pkg.Checksum == nil {
		pkg.Checksum = new(spdx.Checksum)
	}

	if pkg.VerificationCode == nil {
		pkg.VerificationCode = new(spdx.VerificationCode)
	}

	return map[string]updater{
		"PackageName":                 upd(&pkg.Name),
		"PackageVersion":              upd(&pkg.Version),
		"PackageFileName":             upd(&pkg.FileName),
		"PackageSupplier":             upd(&pkg.Supplier),
		"PackageOriginator":           upd(&pkg.Originator),
		"PackageDownloadLocation":     upd(&pkg.DownloadLocation),
		"PackageVerificationCode":     verifCode(pkg.VerificationCode),
		"PackageChecksum":             checksum(pkg.Checksum),
		"PackageHomePage":             upd(&pkg.HomePage),
		"PackageSourceInfo":           upd(&pkg.SourceInfo),
		"PackageLicenseConcluded":     anyLicence(&pkg.LicenceConcluded),
		"PackageLicenseInfoFromFiles": anyLicenceList(&pkg.LicenceInfoFromFiles),
		"PackageLicenseDeclared":      anyLicence(&pkg.LicenceDeclared),
		"PackageLicenseComments":      upd(&pkg.LicenceComments),
		"PackageCopyrightText":        upd(&pkg.CopyrightText),
		"PackageSummary":              upd(&pkg.Summary),
		"PackageDescription":          upd(&pkg.Description),
	}
}

// File mapping.
func fileMap(file *spdx.File) updaterMapping {
	if file.Checksum == nil {
		file.Checksum = new(spdx.Checksum)
	}

	if file.Dependency == nil {
		file.Dependency = make([]*spdx.File, 0)
	}

	if file.ArtifactOf == nil {
		file.ArtifactOf = new([]*spdx.ArtifactOf, 0)
	}
	/* else if p.Key == "ArtifactOfProjectName" {
	if file == nil {
		return nil, errors.New("ArtifactOfProjectName without describing a file.")
	}
	file.ArtifactOf = append(file.ArtifactOf, artif)
	mapping = artifactMap(artif)
	*/return map[string]updater{
		"FileName":          upd(&file.Name),
		"FileType":          upd(&file.Type),
		"FileChecksum":      checksum(file.Checksum),
		"LicenseConcluded":  anyLicence(&file.LicenceConcluded),
		"LicenseInfoInFile": anyLicenceList(&file.LicenceInfoFromFiles),
		"LicenseComments":   upd(&file.LicenseComments),
		"FileCopyrightText": upd(&file.CopyrightText),
		"FileComment":       upd(&file.Comment),
		"FileNotice":        upd(&file.Notice),
		"FileContributor":   updList(&file.Contributor),
		"FileDependency":    updFileNameList(&file.Dependency),
	}
}

// ArtifactOf mapping.
func artifactMap(artif *spdx.ArtifactOf) updaterMapping {
	return map[string]updater{
		"ArtifactOfProjectName":     upd(&artif.Name),
		"ArtifactOfProjectHomepage": upd(&artif.HomePage),
		"ArtifactOfProjectUri":      upd(&artif.ProjectUri),
	}
}

// ExtractedLicensingInfo mapping.
func extractedLicenceMap(lic *spdx.ExtractedLicensingInfo) updaterMapping {
	return map[string]updater{
		"LicenseID":             upd(&lic.Id),
		"ExtractedText":         upd(&lic.Text),
		"LicenseName":           updList(&lic.Name),
		"LicenseCrossReference": updList(&lic.CrossReference),
		"LicenseComment":        upd(&lic.Comment),
	}
}

// Parse a []Pair list to a *spdx.Document
func parseDocument(input []Pair) (*spdx.Document, error) {
	doc := new(spdx.Document)
	mapping := documentMap(doc)

	for _, p := range input {
		if p.Key == "PackageName" {
			pkg := new(spdx.Package)

			doc.Packages = append(doc.Packages, pkg)
			mapping = packageMap(pkg)
		} else if p.Key == "FileName" {
			file := new(spdx.File)

			doc.Files = append(doc.Files, file)
			mapping = fileMap(file)
		} else if p.Key == "LicenseID" {
			file, pkg = nil, nil
			lic := new(spdx.ExtractedLicensingInfo)
			doc.ExtractedLicenceInfo = append(doc.ExtractedLicenceInfo, lic)
			mapping = extractedLicenceMap(lic)
		}

		ok, err := applyMapping(p, mapping)
	}

	return doc, nil
}
