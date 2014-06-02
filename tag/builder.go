package tag

import "../spdx"
import (
	"errors"
	"strings"
)

type updater (func(string) error)
type updaterMapping map[string]updater

func upd(ptr *string) updater {
	return func(val string) error {
		*ptr = val
		return nil
	}
}

func updList(arr *[]string) updater {
	return func(val string) error {
		*arr = append(*arr, val)
		return nil
	}
}

func applyMapping(input Pair, mapping updaterMapping) (bool, error) {
	f, ok := mapping[input.Key]
	if !ok {
		return false, nil
	}
	return true, f(input.Value)
}

func parseDocument(input []Pair) (*spdx.Document, error) {
	doc := new(spdx.Document)
	doc.CreationInfo = new(spdx.CreationInfo)
	doc.CreationInfo.Creator = make([]string, 0)

	mapping := map[string]updater{
		"SpecVersion":     upd(&doc.SpecVersion),
		"DataLicense":     upd(&doc.DataLicence),
		"DocumentComment": upd(&doc.Comment),

		// Creation info:
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

func checksum(cksum *spdx.Checksum) updater {
	return func(val string) error {
		split := strings.Split(val, ":")
		if len(split) != 2 {
			return errors.New("Invalid Package checksum format.")
		}
		cksum.Algo, cksum.Value = strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
		return nil
	}
}

func anyLicence(lic *spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

func anyLicenceList(licList *[]spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

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
