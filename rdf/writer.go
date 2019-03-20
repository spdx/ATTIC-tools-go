package rdf

import (
	"errors"
	"github.com/deltamobile/goraptor"
	"github.com/spdx/tools-go/spdx"
	"os"
	"strconv"
	"strings"
)

// WriteRdf writes the input to the specified RDF format. Does not parse the RDF file into a
// spdx.Document struct and only converts between RDF formats using goraptor.
func WriteRdf(input *os.File, output *os.File, formatIn, formatOut string) error {
	if formatIn == "rdf" {
		formatIn = "guess"
	}
	if formatOut == "rdf" {
		formatOut = Fmt_rdfxmlAbbrev
	}
	parser := goraptor.NewParser(formatIn)
	defer parser.Free()

	serializer := goraptor.NewSerializer(formatOut)
	defer serializer.Free()

	parser.SetNamespaceHandler(func(pfx, uri string) { serializer.SetNamespace(pfx, uri) })

	statements := parser.Parse(input, baseUri)
	err := serializer.SetFile(output, baseUri)
	if err != nil {
		return err
	}
	return serializer.AddN(statements)
}

// Write writes a SPDX Document to rdf/xml abbreviated format.
func Write(output *os.File, doc *spdx.Document) error {
	f := NewFormatter(output, Fmt_rdfxmlAbbrev)
	_, err := f.Document(doc)
	f.Close()
	return err
}

// WriteFormat writes a SPDX Document to raptor format. format must be one of the format
// constants (Fmt_*)
func WriteFormat(output *os.File, doc *spdx.Document, format string) error {
	if format == "rdf" {
		format = Fmt_rdfxmlAbbrev
	} else if !FormatOk(format) {
		return errors.New("Invalid output format.")
	}
	f := NewFormatter(output, format)
	_, err := f.Document(doc)
	f.Close()
	return err
}

// Used to write SPDX Documents in RDF format
type Formatter struct {
	serializer *goraptor.Serializer
	nodeIds    map[string]int

	// index file nodes by name
	fileIds map[string]goraptor.Term
}

// Create a new Formatter that writes to output
func NewFormatter(output *os.File, format string) *Formatter {
	s := goraptor.NewSerializer(format)
	s.StartStream(output, baseUri)

	s.SetNamespace("rdf", "http://www.w3.org/1999/02/22-rdf-syntax-ns#")
	s.SetNamespace("", "http://spdx.org/rdf/terms#")
	s.SetNamespace("rdfs", "http://www.w3.org/2000/01/rdf-schema#")
	s.SetNamespace("doap", "http://usefulinc.com/ns/doap#")

	return &Formatter{
		serializer: s,
		nodeIds:    make(map[string]int),
		fileIds:    make(map[string]goraptor.Term),
	}
}

// Create a new node id for the given prefix
func (f *Formatter) newId(prefix string) *goraptor.Blank {
	f.nodeIds[prefix]++
	id := goraptor.Blank(prefix + strconv.Itoa(f.nodeIds[prefix]))
	return &id
}

// Sets the type t to node
func (f *Formatter) setType(node, t goraptor.Term) error {
	return f.add(node, prefix("ns:type"), t)
}

// Add `key`=`value` at object `to`.
func (f *Formatter) add(to, key, value goraptor.Term) error {
	return f.serializer.Add(&goraptor.Statement{
		Subject:   to,
		Predicate: key,
		Object:    value,
	})
}

// Using the SPDX baseUri, add a goraptor.Term.
func (f *Formatter) addTerm(to goraptor.Term, key string, value goraptor.Term) error {
	return f.add(to, prefix(key), value)
}

// Using the SPDX baseUri, add pairs of literals.
func (f *Formatter) addPairs(to goraptor.Term, pairs ...pair) error {
	for _, p := range pairs {
		if err := f.addLiteral(to, p.key, p.val); err != nil {
			return err
		}
	}
	return nil
}

// Using the SPDX baseUri, add one literal. Does not write anything and returns
// nil if the value is empty.
func (f *Formatter) addLiteral(to goraptor.Term, key, value string) error {
	if value == "" {
		return nil
	}
	return f.add(to, prefix(key), &goraptor.Literal{Value: value})
}

// Document writes a document.
func (f *Formatter) Document(doc *spdx.Document) (docId goraptor.Term, err error) {
	if doc == nil {
		return nil, errors.New("Cannot print nil document.")
	}

	docId = blank("doc")

	if err = f.setType(docId, typeDocument); err != nil {
		return
	}

	if err = f.addLiteral(docId, "specVersion", doc.SpecVersion.Val); err != nil {
		return
	}

	if doc.DataLicence.Val != "" {
		if err = f.addTerm(docId, "dataLicense", uri(licenceUri+doc.DataLicence.Val)); err != nil {
			return
		}
	}

	if id, err := f.CreationInfo(doc.CreationInfo); err == nil {
		if err = f.addTerm(docId, "creationInfo", id); err != nil {
			return docId, err
		}
	} else {
		return docId, err
	}

	if err = f.ExtrLicInfos(docId, "hasExtractedLicensingInfo", doc.ExtractedLicences); err != nil {
		return
	}

	if err = f.addLiteral(docId, "rdfs:comment", doc.Comment.Val); err != nil {
		return
	}

	if err = f.Reviews(docId, "reviewed", doc.Reviews); err != nil {
		return
	}

	if err = f.Packages(docId, "describesPackage", doc.Packages); err != nil {
		return
	}

	if err = f.Files(docId, "referencesFile", doc.Files); err != nil {
		return
	}

	return docId, nil
}

// CreationInfo writes creation info.
func (f *Formatter) CreationInfo(cr *spdx.CreationInfo) (id goraptor.Term, err error) {
	id = f.newId("cri")

	if err = f.setType(id, typeCreationInfo); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"created", cr.Created.V()},
		pair{"licenseListVersion", cr.LicenceListVersion.V()},
		pair{"rdfs:comment", cr.Comment.V()},
	)

	if err != nil {
		return
	}

	for _, creator := range cr.Creator {
		if err = f.addLiteral(id, "creator", creator.V()); err != nil {
			return
		}
	}

	return id, nil
}

// Reviews writes a slice of reviews.
func (f *Formatter) Reviews(parent goraptor.Term, element string, rs []*spdx.Review) error {
	if len(rs) == 0 {
		return nil
	}
	for _, r := range rs {
		revId, err := f.Review(r)
		if err != nil {
			return err
		}
		if revId == nil {
			continue
		}
		if err = f.addTerm(parent, element, revId); err != nil {
			return err
		}
	}
	return nil
}

// Review writes a review.
func (f *Formatter) Review(r *spdx.Review) (id goraptor.Term, err error) {
	id = f.newId("rev")

	if err = f.setType(id, typeReview); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"reviewer", r.Reviewer.V()},
		pair{"reviewDate", r.Date.V()},
		pair{"rdfs:comment", r.Comment.Val},
	)

	return id, err
}

// Packages writes a slice of packages.
func (f *Formatter) Packages(parent goraptor.Term, element string, pkgs []*spdx.Package) error {
	if len(pkgs) == 0 {
		return nil
	}
	for _, pkg := range pkgs {
		pkgid, err := f.Package(pkg)
		if err != nil {
			return err
		}
		if err = f.addTerm(parent, element, pkgid); err != nil {
			return err
		}
	}
	return nil
}

// Package writes a package.
func (f *Formatter) Package(pkg *spdx.Package) (id goraptor.Term, err error) {
	id = f.newId("pkg")

	if err = f.setType(id, typePackage); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"name", pkg.Name.Val},
		pair{"versionInfo", pkg.Version.Val},
		pair{"packageFileName", pkg.FileName.Val},
		pair{"supplier", pkg.Supplier.V()},
		pair{"originator", pkg.Originator.V()},
		pair{"downloadLocation", pkg.DownloadLocation.Val},
		pair{"doap:homepage", pkg.HomePage.Val},
		pair{"sourceInfo", pkg.SourceInfo.Val},
		pair{"licenseComments", pkg.LicenceComments.Val},
		pair{"copyrightText", pkg.CopyrightText.Val},
		pair{"summary", pkg.Summary.Val},
		pair{"description", pkg.Description.Val},
	)
	if err != nil {
		return
	}

	if pkg.VerificationCode != nil {
		pkgid, err := f.VerificationCode(pkg.VerificationCode)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "packageVerificationCode", pkgid); err != nil {
			return id, err
		}
	}

	if pkg.Checksum != nil {
		cksumId, err := f.Checksum(pkg.Checksum)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "checksum", cksumId); err != nil {
			return id, err
		}
	}

	if pkg.LicenceConcluded != nil {
		licId, err := f.Licence(pkg.LicenceConcluded)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "licenseConcluded", licId); err != nil {
			return id, err
		}
	}

	if pkg.LicenceDeclared != nil {
		licId, err := f.Licence(pkg.LicenceDeclared)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "licenseDeclared", licId); err != nil {
			return id, err
		}
	}

	err = f.Licences(id, "licenseInfoFromFiles", pkg.LicenceInfoFromFiles)
	return
}

// VerificationCode writes a VerificationCode
func (f *Formatter) VerificationCode(vc *spdx.VerificationCode) (id goraptor.Term, err error) {
	id = f.newId("vc")

	if err = f.setType(id, typeVerificationCode); err != nil {
		return
	}

	err = f.addLiteral(id, "packageVerificationCodeValue", vc.Value.Val)
	if err != nil {
		return
	}

	for _, excl := range vc.ExcludedFiles {
		err = f.addLiteral(id, "packageVerificationCodeExcludedFile", excl.Val)
		if err != nil {
			return
		}
	}

	return id, nil
}

// Checksum writes a Checksum
func (f *Formatter) Checksum(cksum *spdx.Checksum) (id goraptor.Term, err error) {
	id = f.newId("cksum")

	if err = f.setType(id, typeChecksum); err != nil {
		return
	}

	err = f.addLiteral(id, "checksumValue", cksum.Value.Val)
	if err != nil {
		return
	}

	algo := strings.ToLower(cksum.Algo.Val)
	if algo == "sha1" {
		err = f.addTerm(id, "algorithm", prefix("checksumAlgorithm_sha1"))
	} else {
		err = f.addLiteral(id, "algorithm", algo)
	}

	return id, err
}

// Licences writes a slice of AnyLicence
func (f *Formatter) Licences(parent goraptor.Term, element string, lics []spdx.AnyLicence) error {
	if len(lics) == 0 {
		return nil
	}
	for _, lic := range lics {
		id, err := f.Licence(lic)
		if err != nil {
			return err
		}
		if id == nil {
			continue
		}
		if err = f.addTerm(parent, element, id); err != nil {
			return err
		}
	}
	return nil
}

// Licence writes AnyLicence
func (f *Formatter) Licence(licence spdx.AnyLicence) (id goraptor.Term, err error) {
	switch lic := licence.(type) {
	case spdx.Licence:
		val := lic.LicenceId()
		if !lic.IsReference() {
			return uri(licenceUri + val), nil
		}
		return blank(val), nil
	case spdx.ConjunctiveLicenceSet:
		id = f.newId("lic")
		if err = f.setType(id, typeConjunctiveSet); err != nil {
			return
		}
		for _, mem := range lic.Members {
			memberId, err := f.Licence(mem)
			if err != nil {
				return id, err
			}
			if err = f.addTerm(id, "member", memberId); err != nil {
				return id, err
			}
		}
		return id, nil
	case spdx.DisjunctiveLicenceSet:
		id = f.newId("lic")
		if err = f.setType(id, typeDisjunctiveSet); err != nil {
			return
		}
		for _, mem := range lic.Members {
			memberId, err := f.Licence(mem)
			if err != nil {
				return id, err
			}
			if err = f.addTerm(id, "member", memberId); err != nil {
				return id, err
			}
		}
		return id, nil
	case *spdx.ExtractedLicence:
		return f.ExtrLicInfo(lic)
	}
	return nil, errors.New("Licence type not processed. Please report this error along with the SPDX file you were processing.")
}

// ExtrLicInfos writes a slice of ExtractedLicence
func (f *Formatter) ExtrLicInfos(parent goraptor.Term, element string, lics []*spdx.ExtractedLicence) error {
	if len(lics) == 0 {
		return nil
	}
	for _, lic := range lics {
		licId, err := f.ExtrLicInfo(lic)
		if err != nil {
			return err
		}
		if err = f.addTerm(parent, element, licId); err != nil {
			return err
		}
	}
	return nil
}

// ExtrLicInfo writes an ExtractedLicence
func (f *Formatter) ExtrLicInfo(lic *spdx.ExtractedLicence) (id goraptor.Term, err error) {
	id = blank(lic.LicenceId())
	if lic.LicenceId() == "" {
		id = f.newId("LicenseRef-spdxGoGenId")
	}

	if err = f.setType(id, typeExtractedLicence); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"licenseId", lic.LicenceId()},
		pair{"extractedText", lic.Text.Val},
		pair{"rdfs:comment", lic.Comment.Val},
	)

	if err != nil {
		return
	}

	for _, name := range lic.Name {
		if err = f.addLiteral(id, "name", name.Val); err != nil {
			return
		}
	}

	for _, xref := range lic.CrossReference {
		if err = f.addLiteral(id, "rdfs:seeAlso", xref.Val); err != nil {
			return
		}
	}

	return
}

// Files writes a slice of files.
func (f *Formatter) Files(parent goraptor.Term, element string, files []*spdx.File) error {
	if len(files) == 0 {
		return nil
	}
	for _, file := range files {
		fId, err := f.File(file)
		if err != nil {
			return err
		}
		if fId == nil {
			continue
		}
		if err = f.addTerm(parent, element, fId); err != nil {
			return err
		}
	}
	return nil
}

// File writes a file.
func (f *Formatter) File(file *spdx.File) (id goraptor.Term, err error) {
	id, ok := f.fileIds[file.Name.Val]
	if ok {
		return
	}

	id = f.newId("file")
	f.fileIds[file.Name.Val] = id

	if err = f.setType(id, typeFile); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"fileName", file.Name.Val},
		pair{"licenseComments", file.LicenceComments.Val},
		pair{"copyrightText", file.CopyrightText.Val},
		pair{"rdfs:comment", file.Comment.Val},
		pair{"noticeText", file.Notice.Val},
	)

	if err != nil {
		return
	}

	if file.Type.Val != "" {
		if err = f.addTerm(id, "fileType", prefix(file.Type.Val)); err != nil {
			return
		}
	}

	if file.Checksum != nil {
		cksumId, err := f.Checksum(file.Checksum)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "checksum", cksumId); err != nil {
			return id, err
		}
	}

	if file.LicenceConcluded != nil {
		licId, err := f.Licence(file.LicenceConcluded)
		if err != nil {
			return id, err
		}
		if err = f.addTerm(id, "licenseConcluded", licId); err != nil {
			return id, err
		}
	}

	if err = f.Files(id, "fileDependency", file.Dependency); err != nil {
		return
	}

	err = f.Licences(id, "licenseInfoInFile", file.LicenceInfoInFile)
	return
}

// Closes the stream and frees the serializer. Always call after writing using
// the Formatter.
func (f *Formatter) Close() {
	f.serializer.EndStream()
	f.serializer.Free()
}
