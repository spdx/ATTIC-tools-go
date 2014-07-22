package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"errors"
	"io"
	"unicode"
)

const commentLastWritten = "__comment"

// spdx.Checksum representation as Tag string
func cksumStr(cksum *spdx.Checksum) string {
	if cksum == nil || (cksum.Algo.Val == "" && cksum.Value.Val == "") {
		return ""
	}
	return cksum.Algo.Val + ": " + cksum.Value.Val
}

// spdx.VerificationCode representation as a valid Tag string
func verifCodeStr(verif *spdx.VerificationCode) string {
	if verif == nil || (verif.Value.Val == "" && len(verif.ExcludedFiles) == 0) {
		return ""
	}
	if len(verif.ExcludedFiles) == 0 {
		return verif.Value.Val
	}
	return verif.Value.Val + " (Excludes: " + spdx.Join(verif.ExcludedFiles, ", ") + ")"
}

// Count the number of *sep* at the beginning of *str*.
func countLeft(str string, sep byte) (count int) {
	for i := range str {
		if str[i] != sep {
			return
		}
		count++
	}
	return
}

// Checks if the given *spdx.File is in the given []*spdx.File.
func fileInList(file *spdx.File, list []*spdx.File) bool {
	for _, f := range list {
		if f == file {
			return true
		}
	}
	return false
}

// Formatter is the pretty-printer for Tag format. It is aware of what has been
// printed previously in order to leave nice newlines.
type Formatter struct {
	lastWritten string
	out         io.Writer
}

// Create a new *Formatter that writes to f.
func NewFormatter(f io.Writer) *Formatter {
	return &Formatter{"", f}
}

// Print newlines where appropriate.
//
// Currently when:
// - a property is followed by a comment
// - printing one of these properties: FileName, LicenseID, PackageName,
//   Reviewer, ArtifactOfProjectName
func (f *Formatter) spaces(now string) {
	if f.lastWritten == "" || f.lastWritten == commentLastWritten {
		return
	}

	breaks := []string{"FileName", "LicenseID", "PackageName", "Reviewer", "ArtifactOfProjectName"}

	for _, w := range breaks {
		if w == now {
			f.out.Write([]byte{'\n'})
			return
		}
	}

	if now == commentLastWritten {
		f.out.Write([]byte{'\n'})
	}
}

// Read all tokens from a lexer and pretty-print them.
func (f *Formatter) Lexer(lex lexer) error {
	for lex.Lex() {
		err := f.Token(lex.Token())
		if err != nil {
			return err
		}
	}
	return lex.Err()
}

// Write a Token.
func (f *Formatter) Token(tok *Token) error {
	if tok == nil || (tok.Type == TokenPair && tok.Pair.Value == "") {
		return nil
	}
	switch tok.Type {
	case TokenComment:
		return f.Comment(tok.Pair.Value)
	case TokenPair:
		return f.Property(tok.Pair.Key, tok.Pair.Value)
	default:
		return errors.New("Unsupported token type.")
	}
}

// Write a comment (# comment).
func (f *Formatter) Comment(comment string) error {
	f.spaces(commentLastWritten)

	hashes := countLeft(comment, '#')
	if hashes == len(comment) || (hashes < len(comment) && !unicode.IsSpace(rune(comment[hashes]))) {
		comment = comment[:hashes] + " " + comment[hashes:]
	}

	f.lastWritten = commentLastWritten
	_, err := io.WriteString(f.out, "#"+comment+"\n")
	return err
}

// Write a property (tag: value).
func (f *Formatter) Property(tag, value string) error {
	if value == "" {
		return nil
	}

	f.spaces(tag)
	if value != spdx.NOASSERTION && value != spdx.NONE && (isMultiline(tag) || isMultilineValue(value)) {
		value = "<text>" + value + "</text>"
	}

	f.lastWritten = tag
	_, err := io.WriteString(f.out, tag+": "+value+"\n")
	return err
}

// Write a list of properties.
func (f *Formatter) Properties(props []Pair) error {
	for _, p := range props {
		if err := f.Property(p.Key, p.Value); err != nil {
			return err
		}
	}
	return nil
}

// Write a property with multiple values. The same tag is printed for each
// non-empty value found in `values`.
func (f *Formatter) PropertySlice(tag string, values []spdx.ValueStr) error {
	for _, val := range values {
		if err := f.Property(tag, val.Val); err != nil {
			return err
		}
	}
	return nil
}

// Write a ValueCreator slice. The same tag is printed for each non-empty value
// found in `values`.
func (f *Formatter) CreatorSlice(tag string, values []spdx.ValueCreator) error {
	for _, val := range values {
		if err := f.Property(tag, val.V()); err != nil {
			return err
		}
	}
	return nil
}

// Write a list of licences. The same tag is printed for each non-empty value
// found in `values`.
func (f *Formatter) PropertyLicenceSlice(tag string, values []spdx.AnyLicence) error {
	for _, lic := range values {
		if err := f.Property(tag, lic.LicenceId()); err != nil {
			return err
		}
	}
	return nil
}

// Write `doc` incuding all its nested elements.
func (f *Formatter) Document(doc *spdx.Document) error {
	if doc == nil {
		return nil
	}

	err := f.Properties([]Pair{
		{"SPDXVersion", doc.SpecVersion.Val},
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

	return f.ExtractedLicences(doc.ExtractedLicences)
}

// Write `ci` the creation info part of a document.
func (f *Formatter) CreationInfo(ci *spdx.CreationInfo) error {
	if ci == nil {
		return nil
	}

	if err := f.CreatorSlice("Creator", ci.Creator); err != nil {
		return err
	}

	return f.Properties([]Pair{
		{"Created", ci.Created.V()},
		{"CreatorComment", ci.Comment.Val},
		{"LicenseListVersion", ci.LicenceListVersion.Val},
	})
}

// Write all the Packages in pkgs.
func (f *Formatter) Packages(pkgs []*spdx.Package) error {
	for _, pkg := range pkgs {
		if err := f.Package(pkg); err != nil {
			return err
		}
	}
	return nil
}

// Write `pkg` and all its nested elements.
func (f *Formatter) Package(pkg *spdx.Package) error {
	if pkg == nil {
		return nil
	}

	err := f.Properties([]Pair{
		{"PackageName", pkg.Name.Val},
		{"PackageVersion", pkg.Version.Val},
		{"PackageFileName", pkg.FileName.Val},
		{"PackageSupplier", pkg.Supplier.V()},
		{"PackageOriginator", pkg.Originator.V()},
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

// Write all elements in `files`.
func (f *Formatter) Files(files []*spdx.File) error {
	for _, file := range files {
		if err := f.File(file); err != nil {
			return err
		}
	}
	return nil
}

// Write `file` and all its nested elements.
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

// Write all the reviews in `reviews`.
func (f *Formatter) Reviews(reviews []*spdx.Review) error {
	for _, review := range reviews {
		if err := f.Review(review); err != nil {
			return err
		}
	}
	return nil
}

// Write the *spdx.Review `rev`.
func (f *Formatter) Review(review *spdx.Review) error {
	if review == nil {
		return nil
	}

	return f.Properties([]Pair{
		{"Reviewer", review.Reviewer.V()},
		{"ReviewDate", review.Date.V()},
		{"ReviewComment", review.Comment.V()},
	})
}

// Write all licences in `lics`.
func (f *Formatter) ExtractedLicences(lics []*spdx.ExtractedLicence) error {
	for _, lic := range lics {
		if err := f.ExtractedLicence(lic); err != nil {
			return err
		}
	}
	return nil
}

// Write the ExtractedLicence `lic`.
func (f *Formatter) ExtractedLicence(lic *spdx.ExtractedLicence) error {
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
