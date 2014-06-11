package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"io"
	"strings"
)

func isMultiline(property string) bool {
	_, ok := multilineProperties[property]
	return ok
}

func isMultilineValue(val string) bool {
	return strings.Index(val, "\n") >= 0
}

var multilineProperties = multilineInit()

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

func cksumStr(cksum *spdx.Checksum) string {
	if cksum == nil || (cksum.Algo == "" && cksum.Value == "") {
		return ""
	}
	return cksum.Algo + ": " + cksum.Value
}

func verifCodeStr(verif *spdx.VerificationCode) string {
	if verif == nil || (verif.Value == "" && len(verif.ExcludedFiles) == 0) {
		return ""
	}
	if len(verif.ExcludedFiles) == 0 {
		return verif.Value
	}
	return verif.Value + " (Excludes: " + strings.Join(verif.ExcludedFiles, ", ") + ")"
}

func writeProperties(f io.Writer, props []Pair) error {
	for _, p := range props {
		if err := writeProperty(f, p.Key, p.Value); err != nil {
			return err
		}
	}
	return nil
}

func writeProperty(f io.Writer, tag, value string) error {
	if value == "" {
		return nil
	}

	if isMultiline(tag) || isMultilineValue(value) {
		value = "<text>" + value + "</text>"
	}

	_, err := io.WriteString(f, tag+": "+value+"\n")
	if err != nil {
		return err
	}
	return nil
}

func writePropertySlice(f io.Writer, tag string, values []string) error {
	for _, val := range values {
		if err := writeProperty(f, tag, val); err != nil {
			return err
		}
	}
	return nil
}

func writePropertyLicenceSlice(f io.Writer, tag string, values []spdx.AnyLicenceInfo) error {
	for _, lic := range values {
		if err := writeProperty(f, tag, lic.LicenceId()); err != nil {
			return err
		}
	}
	return nil
}

func fileInList(file *spdx.File, list []*spdx.File) bool {
	for _, f := range list {
		if f == file {
			return true
		}
	}
	return false
}

func writeDocument(f io.Writer, doc *spdx.Document) error {
	if doc == nil {
		return nil
	}

	err := writeProperties(f, []Pair{
		{"SpecVersion", doc.SpecVersion},
		{"DataLicense", doc.DataLicence},
		{"DocumentComment", doc.Comment},
	})

	if err != nil {
		return err
	}

	if err = writeCreationInfo(f, doc.CreationInfo); err != nil {
		return err
	}

	if err = writePackages(f, doc.Packages); err != nil {
		return err
	}

	files := doc.Files
	for _, pkg := range doc.Packages {
		// add all files that are not there yet
		for _, file := range pkg.Files {
			if !fileInList(file, files) {
				files = append(files, file)
			}
		}
	}

	if err = writeFiles(f, doc.Files); err != nil {
		return err
	}

	if err = writeReviews(f, doc.Reviews); err != nil {
		return err
	}

	return writeExtractedLicenceInfo(f, doc.ExtractedLicenceInfo)
}

func writeCreationInfo(f io.Writer, ci *spdx.CreationInfo) error {
	if ci == nil {
		return nil
	}

	if err := writePropertySlice(f, "Creator", ci.Creator); err != nil {
		return err
	}

	return writeProperties(f, []Pair{
		{"Created", ci.Created},
		{"CreatorComment", ci.Comment},
		{"LicenseListVersion", ci.LicenceListVersion},
	})
}

func writePackages(f io.Writer, pkgs []*spdx.Package) error {
	for _, pkg := range pkgs {
		if err := writePkg(f, pkg); err != nil {
			return err
		}
	}
	return nil
}

func writePkg(f io.Writer, pkg *spdx.Package) error {
	if pkg == nil {
		return nil
	}

	err := writeProperties(f, []Pair{
		{"PackageName", pkg.Name},
		{"PackageVersion", pkg.Version},
		{"PackageFileName", pkg.FileName},
		{"PackageSupplier", pkg.Supplier},
		{"PackageOriginator", pkg.Originator},
		{"PackageDownloadLocation", pkg.DownloadLocation},
		{"PackageVerificationCode", verifCodeStr(pkg.VerificationCode)},
		{"PackageChecksum", cksumStr(pkg.Checksum)},
		{"PackageHomePage", pkg.HomePage},
		{"PackageSourceInfo", pkg.SourceInfo},
	})

	if err != nil {
		return err
	}
	if pkg.LicenceConcluded != nil {
		if err = writeProperty(f, "PackageLicenseConcluded", pkg.LicenceConcluded.LicenceId()); err != nil {
			return err
		}
	}
	if pkg.LicenceDeclared != nil {
		if err = writeProperty(f, "PackageLicenseDeclared", pkg.LicenceDeclared.LicenceId()); err != nil {
			return err
		}
	}
	if err = writePropertyLicenceSlice(f, "PackageLicenseInfoFromFiles", pkg.LicenceInfoFromFiles); err != nil {
		return err
	}

	return writeProperties(f, []Pair{
		{"PackageLicenseComments", pkg.LicenceComments},
		{"PackageCopyrightText", pkg.CopyrightText},
		{"PackageSummary", pkg.Summary},
		{"PackageDescription", pkg.Description},
	})
}

func writeFiles(f io.Writer, files []*spdx.File) error {
	for _, file := range files {
		if err := writeFile(f, file); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(f io.Writer, file *spdx.File) error {
	if file == nil {
		return nil
	}
	err := writeProperties(f, []Pair{
		{"FileName", file.Name},
		{"FileType", file.Type},
		{"FileChecksum", cksumStr(file.Checksum)},
	})

	if file.LicenceConcluded != nil {
		if err = writeProperty(f, "LicenseConcluded", file.LicenceConcluded.LicenceId()); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	if err = writePropertyLicenceSlice(f, "LicenseInfoInFile", file.LicenceInfoInFile); err != nil {
		return err
	}

	err = writeProperties(f, []Pair{
		{"LicenseComments", file.LicenceComments},
		{"FileCopyrightText", file.CopyrightText},
		{"FileComment", file.Comment},
		{"FileNotice", file.Notice},
	})
	if err != nil {
		return err
	}

	if err = writePropertySlice(f, "FileContributor", file.Contributor); err != nil {
		return err
	}

	for _, fname := range file.Dependency {
		if err = writeProperty(f, "FileDependency", fname.Name); err != nil {
			return err
		}
	}

	return nil
}

func writeReviews(f io.Writer, reviews []*spdx.Review) error {
	for _, review := range reviews {
		if err := writeReview(f, review); err != nil {
			return err
		}
	}
	return nil
}

func writeReview(f io.Writer, review *spdx.Review) error {
	if review == nil {
		return nil
	}

	return writeProperties(f, []Pair{
		{"Reviewer", review.Reviewer},
		{"ReviewDate", review.Date},
		{"ReviewComment", review.Comment},
	})
}

func writeExtractedLicenceInfo(f io.Writer, lics []*spdx.ExtractedLicensingInfo) error {
	for _, lic := range lics {
		if err := writeExtrLicInfo(f, lic); err != nil {
			return err
		}
	}
	return nil
}

func writeExtrLicInfo(f io.Writer, lic *spdx.ExtractedLicensingInfo) error {
	if lic == nil {
		return nil
	}
	err := writeProperties(f, []Pair{
		{"LicenseID", lic.Id},
		{"ExtractedText", lic.Text},
	})
	if err != nil {
		return err
	}
	if err = writePropertySlice(f, "LicenseName", lic.Name); err != nil {
		return err
	}
	if err = writePropertySlice(f, "LicenseCrossReference", lic.CrossReference); err != nil {
		return err
	}
	return writeProperty(f, "LicenseComment", lic.Comment)
}
