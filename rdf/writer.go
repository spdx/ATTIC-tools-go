package rdf

import (
	"github.com/vladvelici/goraptor"
	"io"
	"os"
)

// Writes the input to the specified RDF format
func WriteRdf(input io.Reader, output io.Writer, format string) error {
	parser := goraptor.NewParser("guess")
	defer parser.Free()

	serializer := goraptor.NewSerializer(format)
	defer serializer.Free()

	parser.SetNamespaceHandler(func(pfx, uri string) { serializer.SetNamespace(pfx, uri) })

	statements := parser.Parse(input, "")

	str, err := serializer.Serialize(statements, "")
	if err != nil {
		return err
	}

	_, err = io.WriteString(output, str)
	return err
}

// Used to write documents in RDF format
type Formatter struct {
	serializer *goraptor.Serializer
}

// Create a new Formatter that writes to output
func NewFormatter(output *os.File, format string) *Formatter {
	s := goraptor.NewSerializer(format)
	s.StartStream(output, "")
	return &Fromatter{
		serializer: s,
	}
}

// Write a document.
func (f *Formatter) Document(doc *spdx.Document) error {
	return nil
}

// Write creation info.
func (f *Formatter) CreationInfo(cr *spdx.CreationInfo) error {
	return nil
}

// Write a slice of reviews.
func (f *Formatter) Reviews(rs *[]spdx.Review) error {
	for _, r := range rs {
		if err := f.Review(r); err != nil {
			return err
		}
	}
	return nil
}

// Write a review.
func (f *Formatter) Review(r *spdx.Review) error {
	return nil
}

// Write a slice of packages.
func (f *Formatter) Packages(pkgs []*spdx.Package) error {
	for _, pkg := range pkgs {
		if err := f.Package(pkg); err != nil {
			return err
		}
	}
	return nil
}

// Write a package.
func (f *Formatter) Package(pkg *spdx.Package) error {
	return nil
}

// Write a slice of ExtractedLicenceInfo
func (f *Formatter) ExtrLicInfos(lics []*spdx.ExtractedLicenceInfo) error {
	for _, lic := range lics {
		if err := f.ExtrLicInfo(lic); err != nil {
			return err
		}
	}
	return nil
}

// Write an ExtractedLicenceInfo
func (f *Formatter) ExtrLicInfo(lic *spdx.ExtractedLicenceInfo) error {
	return nil
}

// Write a slice of files.
func (f *Formatter) Files(files []*spdx.File) error {
	for _, file := range files {
		if err := f.File(file); err != nil {
			return err
		}
	}
	return nil
}

// Write a file.
func (f *Formatter) File(file *spdx.File) error {
	return nil
}

// Closes the stream and frees the serializer. Always call after writing using the Formatter.
func (f *Formatter) Close() {
	f.serializer.EndStream()
	f.serializer.Close()
}
