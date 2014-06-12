package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"errors"
	"io"
	"strings"
	"unicode"
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
	if cksum == nil || (cksum.Algo.Val == "" && cksum.Value.Val == "") {
		return ""
	}
	return cksum.Algo.Val + ": " + cksum.Value.Val
}

func verifCodeStr(verif *spdx.VerificationCode) string {
	if verif == nil || (verif.Value.Val == "" && len(verif.ExcludedFiles) == 0) {
		return ""
	}
	if len(verif.ExcludedFiles) == 0 {
		return verif.Value.Val
	}
	return verif.Value.Val + " (Excludes: " + spdx.Join(verif.ExcludedFiles, ", ") + ")"
}

func countLeft(str string, sep byte) (count int) {
	for i := 0; i < len(str); i++ {
		if str[i] != sep {
			return
		}
		count++
	}
	return
}

func fileInList(file *spdx.File, list []*spdx.File) bool {
	for _, f := range list {
		if f == file {
			return true
		}
	}
	return false
}

// Formatter is the pretty-printer for Tag format. It is aware of what has been printed previously
// in order to leave nice newlines.
type Formatter struct {
	lastWritten string
	out         io.Writer
}

func NewFormatter(f io.Writer) *Formatter {
	return &Formatter{"", f}
}

// Print newlines where appropriate.
// Currently when
// - a property is followed by a comment
// - a property that creates a new domain comes after another property
// (those are: FileName, PackageName, LicenseID)
func (f *Formatter) spaces(now string) {
	if f.lastWritten == "" || f.lastWritten == "__comment" {
		return
	}

	breaks := []string{"FileName", "LicenseID", "PackageName", "Reviewer", "ArtifactOfProjectName"}

	for _, w := range breaks {
		if w == now {
			f.out.Write([]byte{'\n'})
			return
		}
	}
}

// Read all tokens from a lexer and pretty-print them
func (f *Formatter) Lexer(lex lexer) error {
	for lex.Lex() {
		err := f.Token(lex.Token())
		if err != nil {
			return err
		}
	}
	return lex.Err()
}

// Write a Token
func (f *Formatter) Token(tok *Token) error {
	if tok == nil || (tok.Type == TokenPair && tok.Pair.Value == "") {
		return nil
	}
	if tok.Type == TokenComment {
		f.spaces("__comment")
		hashes := countLeft(tok.Pair.Value, '#')
		if hashes != len(tok.Pair.Value) && !unicode.IsSpace(rune(tok.Pair.Value[hashes])) {
			tok.Pair.Value = tok.Pair.Value[:hashes] + " " + tok.Pair.Value[hashes+1:]
		}

		f.lastWritten = "__comment"
		_, err := io.WriteString(f.out, "#"+tok.Pair.Value+"\n")
		return err
	}

	if tok.Type != TokenPair {
		return errors.New("Unsupported token type.")
	}

	return f.Property(tok.Pair.Key, tok.Pair.Value)
}

// Write a property (tag: value)
func (f *Formatter) Property(tag, value string) error {
	if value == "" {
		return nil
	}

	f.spaces(tag)
	if isMultiline(tag) || isMultilineValue(value) {
		value = "<text>" + value + "</text>"
	}

	f.lastWritten = tag
	_, err := io.WriteString(f.out, tag+": "+value+"\n")
	return err
}

// Write a list of properties
func (f *Formatter) Properties(props []Pair) error {
	for _, p := range props {
		if err := f.Property(p.Key, p.Value); err != nil {
			return err
		}
	}
	return nil
}

// Write a property with multiple values
func (f *Formatter) PropertySlice(tag string, values []spdx.ValueStr) error {
	for _, val := range values {
		if err := f.Property(tag, val.Val); err != nil {
			return err
		}
	}
	return nil
}

// Write a list of licences
func (f *Formatter) PropertyLicenceSlice(tag string, values []spdx.AnyLicenceInfo) error {
	for _, lic := range values {
		if err := f.Property(tag, lic.LicenceId()); err != nil {
			return err
		}
	}
	return nil
}

// Write a spdx.Document, incuding all its contents
func (f *Formatter) Document(doc *spdx.Document) error {
	if doc == nil {
		return nil
	}

	err := f.Properties([]Pair{
		{"SpecVersion", doc.SpecVersion.Val},
		{"DataLicense", doc.DataLicence.Val},
		{"DocumentComment", doc.Comment.Val},
	})

	if err != nil {
		return err
	}

	if err = f.CreationInfo(doc.CreationInfo); err != nil {
		return err
	}

	if err = f.Packages(doc.Packages); err != nil {
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

	if err = f.Files(doc.Files); err != nil {
		return err
	}

	if err = f.Reviews(doc.Reviews); err != nil {
		return err
	}

	return f.ExtractedLicenceInfo(doc.ExtractedLicenceInfo)
}

// Write the creation info part of a document
func (f *Formatter) CreationInfo(ci *spdx.CreationInfo) error {
	if ci == nil {
		return nil
	}

	if err := f.PropertySlice("Creator", ci.Creator); err != nil {
		return err
	}

	return f.Properties([]Pair{
		{"Created", ci.Created.Val},
		{"CreatorComment", ci.Comment.Val},
		{"LicenseListVersion", ci.LicenceListVersion.Val},
	})
}

func (f *Formatter) Packages(pkgs []*spdx.Package) error {
	for _, pkg := range pkgs {
		if err := f.Package(pkg); err != nil {
			return err
		}
	}
	return nil
}

// Write a package
func (f *Formatter) Package(pkg *spdx.Package) error {
	if pkg == nil {
		return nil
	}

	err := f.Properties([]Pair{
		{"PackageName", pkg.Name.Val},
		{"PackageVersion", pkg.Version.Val},
		{"PackageFileName", pkg.FileName.Val},
		{"PackageSupplier", pkg.Supplier.Val},
		{"PackageOriginator", pkg.Originator.Val},
		{"PackageDownloadLocation", pkg.DownloadLocation.Val},
		{"PackageVerificationCode", verifCodeStr(pkg.VerificationCode)},
		{"packageChecksum", cksumStr(pkg.Checksum)},
		{"PackageHomePage", pkg.HomePage.Val},
		{"PackageSourceInfo", pkg.SourceInfo.Val},
	})

	if err != nil {
		return err
	}
	if pkg.LicenceConcluded != nil {
		if err = f.Property("PackageLicenseConcluded", pkg.LicenceConcluded.LicenceId()); err != nil {
			return err
		}
	}
	if pkg.LicenceDeclared != nil {
		if err = f.Property("PackageLicenseDeclared", pkg.LicenceDeclared.LicenceId()); err != nil {
			return err
		}
	}
	if err = f.PropertyLicenceSlice("PackageLicenseInfoFromFiles", pkg.LicenceInfoFromFiles); err != nil {
		return err
	}

	return f.Properties([]Pair{
		{"PackageLicenseComments", pkg.LicenceComments.Val},
		{"PackageCopyrightText", pkg.CopyrightText.Val},
		{"PackageSummary", pkg.Summary.Val},
		{"PackageDescription", pkg.Description.Val},
	})
}

// Write a list of Files
func (f *Formatter) Files(files []*spdx.File) error {
	for _, file := range files {
		if err := f.File(file); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) File(file *spdx.File) error {
	if file == nil {
		return nil
	}
	err := f.Properties([]Pair{
		{"FileName", file.Name.Val},
		{"FileType", file.Type.Val},
		{"FileChecksum", cksumStr(file.Checksum)},
	})

	if file.LicenceConcluded != nil {
		if err = f.Property("LicenseConcluded", file.LicenceConcluded.LicenceId()); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	if err = f.PropertyLicenceSlice("LicenseInfoInFile", file.LicenceInfoInFile); err != nil {
		return err
	}

	err = f.Properties([]Pair{
		{"LicenseComments", file.LicenceComments.Val},
		{"FileCopyrightText", file.CopyrightText.Val},
		{"FileComment", file.Comment.Val},
		{"FileNotice", file.Notice.Val},
	})
	if err != nil {
		return err
	}

	if err = f.PropertySlice("FileContributor", file.Contributor); err != nil {
		return err
	}

	for _, fname := range file.Dependency {
		if err = f.Property("FileDependency", fname.Name.Val); err != nil {
			return err
		}
	}

	return nil
}

func (f *Formatter) Reviews(reviews []*spdx.Review) error {
	for _, review := range reviews {
		if err := f.Review(review); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) Review(review *spdx.Review) error {
	if review == nil {
		return nil
	}

	return f.Properties([]Pair{
		{"Reviewer", review.Reviewer.Val},
		{"ReviewDate", review.Date.Val},
		{"ReviewComment", review.Comment.Val},
	})
}

func (f *Formatter) ExtractedLicenceInfo(lics []*spdx.ExtractedLicensingInfo) error {
	for _, lic := range lics {
		if err := f.ExtrLicInfo(lic); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) ExtrLicInfo(lic *spdx.ExtractedLicensingInfo) error {
	if lic == nil {
		return nil
	}
	err := f.Properties([]Pair{
		{"LicenseID", lic.Id.Val},
		{"ExtractedText", lic.Text.Val},
	})
	if err != nil {
		return err
	}
	if err = f.PropertySlice("LicenseName", lic.Name); err != nil {
		return err
	}
	if err = f.PropertySlice("LicenseCrossReference", lic.CrossReference); err != nil {
		return err
	}
	return f.Property("LicenseComment", lic.Comment.Val)
}
