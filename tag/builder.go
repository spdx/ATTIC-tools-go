package tag

import "../spdx"

import (
	"errors"
	"strings"
)

// Updater
type updater (func(string) error)
type updaterMapping map[string]updater

// Accept list of error as error
type errorList []error

func (e errorList) Error() string {
	str := ""
	for err := range e {
		str += err.Error() + "\n"
	}
	return strings.TrimSpace(str)
}

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

// Apply the relevant updater function if the given pair matches any.
func applyMapping(input Pair, mapping updaterMapping) (ok bool, err error) {
	f, ok := mapping[input.Key]
	if !ok {
		return false, nil
	}
	return true, f(input.Value)
}

// Parse a []Pair list to a *spdx.Document
func parseDocument(input []Pair) (*spdx.Document, error) {
	doc := new(spdx.Document)
	doc.CreationInfo = new(spdx.CreationInfo)
	doc.CreationInfo.Creator = make([]string, 0)

	mapping := map[string]updater{
		"SpecVersion":        upd(&doc.SpecVersion),
		"DataLicense":        upd(&doc.DataLicence),
		"DocumentComment":    upd(&doc.Comment),
		"Creator":            updList(&doc.CreationInfo.Creator),
		"Created":            upd(&doc.CreationInfo.Created),
		"CreatorComment":     upd(&doc.CreationInfo.Comment),
		"LicenseListVersion": upd(&doc.CreationInfo.LicenceListVersion),
	}

	for _, p := range input {
		f, ok := mapping[p.Key]
		if !ok {
			break
		}
		f(p.Value)
		// todo
	}

	return doc, nil
}

// Parse a []Pair to a *spdx.Package
func parsePackage(input []Pair) (pkg *spdx.Package, err error) {
	pkg = new(spdx.Package)
	pkg.Checksum = new(spdx.Checksum)
	pkg.VerificationCode = new(spdx.VerificationCode)
	pkg.Checksum = new(spdx.Checksum)

	mapping := map[string]updater{
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
	for _, p := range input {
		f, ok := mapping[p.Key]
		if !ok {
			break
		}
		f(p.Value)
		// todo
	}

	return
}
