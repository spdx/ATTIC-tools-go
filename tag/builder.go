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

// Gets all the key/value combinations in src and puts them in dest (overwrites if values already exist)
func mapMerge(dest *updaterMapping, src updaterMapping) {
	mp := *dest
	for k, v := range src {
		mp[k] = v
	}
}

// Document mapping.
func documentMap(doc *spdx.Document) updaterMapping {
	doc.CreationInfo = new(spdx.CreationInfo)
	doc.CreationInfo.Creator = make([]string, 0)
	doc.Packages = make([]*spdx.Package, 0)
	doc.Files = make([]*spdx.File, 0)
	doc.ExtractedLicenceInfo = make([]*spdx.ExtractedLicensingInfo, 0)

	var mapping updaterMapping

	mapping = map[string]updater{
		// SpdxDocument
		"SpecVersion":        upd(&doc.SpecVersion),
		"DataLicense":        upd(&doc.DataLicence),
		"DocumentComment":    upd(&doc.Comment),
		"Creator":            updList(&doc.CreationInfo.Creator),
		"Created":            upd(&doc.CreationInfo.Created),
		"CreatorComment":     upd(&doc.CreationInfo.Comment),
		"LicenseListVersion": upd(&doc.CreationInfo.LicenceListVersion),

		// Package
		"PackageName": func(val string) error {
			pkg := &spdx.Package{
				Name:             val,
				Checksum:         new(spdx.Checksum),
				VerificationCode: new(spdx.VerificationCode),
			}
			doc.Packages = append(doc.Packages, pkg)

			// Add package values that are now available
			mapMerge(&mapping, updaterMapping{
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
			})

			return nil
		},
		// File
		"FileName": func(val string) error {
			file := &spdx.File{
				Checksum:   new(spdx.Checksum),
				Dependency: make([]*spdx.File, 0),
				ArtifactOf: make([]*spdx.ArtifactOf, 0),
			}

			mapMerge(&mapping, updaterMapping{
				"FileType":          upd(&file.Type),
				"FileChecksum":      checksum(file.Checksum),
				"LicenseConcluded":  anyLicence(&file.LicenceConcluded),
				"LicenseInfoInFile": anyLicenceList(&file.LicenceInfoInFile),
				"LicenseComments":   upd(&file.LicenceComments),
				"FileCopyrightText": upd(&file.CopyrightText),
				"FileComment":       upd(&file.Comment),
				"FileNotice":        upd(&file.Notice),
				"FileContributor":   updList(&file.Contributor),
				"FileDependency":    updFileNameList(&file.Dependency),
				"ArtifactOfProjectName": func(val string) error {
					artif := new(spdx.ArtifactOf)
					mapMerge(&mapping, updaterMapping{
						"ArtifactOfProjectHomepage": upd(&artif.HomePage),
						"ArtifactOfProjectUri":      upd(&artif.ProjectUri),
					})
					file.ArtifactOf = append(file.ArtifactOf, artif)
					return nil
				},
			})

			return nil
		},

		// ExtractedLicensingInfo
		"LicenseID": func(val string) error {
			lic := &spdx.ExtractedLicensingInfo{
				Name:           make([]string, 0),
				CrossReference: make([]string, 0),
			}
			mapMerge(&mapping, updaterMapping{
				"ExtractedText":         upd(&lic.Text),
				"LicenseName":           updList(&lic.Name),
				"LicenseCrossReference": updList(&lic.CrossReference),
				"LicenseComment":        upd(&lic.Comment),
			})
			doc.ExtractedLicenceInfo = append(doc.ExtractedLicenceInfo, lic)
			return nil
		},
	}

	return mapping
}

// Apply the relevant updater function if the given pair matches any.
//
// ok means whether the property was in the map or not
// err is the error returned by applying the mapping function or, if ok == false, an error with the relevant "mapping not found" message
//
// It returns two arguments to allow for easily creating parsing modes such as "ignore not known mapping"
func applyMapping(input Pair, mapping updaterMapping) (ok bool, err error) {
	f, ok := mapping[input.Key]
	if !ok {
		return false, errors.New("Invalid property or property needs another property to be defined before it: " + input.Key)
	}
	return true, f(input.Value)
}

// Parse a []Pair list to a *spdx.Document
func parseDocument(input []Pair) (*spdx.Document, error) {
	doc := new(spdx.Document)
	mapping := documentMap(doc)

	for _, p := range input {
		_, err := applyMapping(p, mapping)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}
