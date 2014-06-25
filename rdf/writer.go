package rdf

import (
	"github.com/deltamobile/goraptor"
	"github.com/vladvelici/spdx-go/spdx"
	"io"
	"os"
	"strconv"
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
	nodeIds    map[string]int
}

// Create a new Formatter that writes to output
func NewFormatter(output *os.File, format string) *Formatter {
	s := goraptor.NewSerializer(format)
	s.StartStream(output, baseUri)
	return &Formatter{
		serializer: s,
		nodeIds:    make(map[string]int),
	}
}

// Create a new node id for the given prefix
func (f *Formatter) newId(prefix string) *goraptor.Blank {
	f.nodeIds[prefix]++
	id := goraptor.Blank(prefix + strconv.Itoa(f.nodeIds[prefix]))
	return &id
}

// Sets the type t to node
func (f *Formatter) setType(node goraptor.Term, t string) error {
	return f.add(node, prefix("ns:type"), prefix(t))
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

// Using the SPDX baseUri, add one literal. Does not write anything and returns nil if the value is empty.
func (f *Formatter) addLiteral(to goraptor.Term, key, value string) error {
	if value == "" {
		return nil
	}
	return f.add(to, prefix(key), &goraptor.Literal{Value: value})
}

// Write a document.
func (f *Formatter) Document(doc *spdx.Document) (docId goraptor.Term, err error) {
	_docId := goraptor.Blank("doc")
	docId = &_docId

	if err = f.setType(docId, "SpdxDocument"); err != nil {
		return
	}

	if err = f.addLiteral(docId, "specVersion", doc.SpecVersion.Val); err != nil {
		return
	}

	if err = f.addTerm(docId, "dataLicense", uri(licenceUri+doc.DataLicence.Val)); err != nil {
		return
	}

	if id, err := f.CreationInfo(doc.CreationInfo); err == nil {
		if err = f.addTerm(docId, "creationInfo", id); err != nil {
			return docId, err
		}
	} else {
		return docId, err
	}

	if err = f.addLiteral(docId, "rdfs:comment", doc.Comment.Val); err != nil {
		return
	}

	/*
		if err = f.Reviews(docId, "reviewed", doc.Reviews); err != nil {
			return
		}
	*/
	if err = f.Packages(docId, "describesPackage", doc.Packages); err != nil {
		return
	}

	/*
	   if err = f.Files(docId, "referencesFile", doc.Files); err != nil {
	       return
	   }

	   if err = f.ExtrLicInfos(docId, "hasExtractedLicensingInfo", doc.ExtractedLicenceInfo); err != nil {
	       return
	   }
	*/

	return docId, nil
}

// Write creation info.
func (f *Formatter) CreationInfo(cr *spdx.CreationInfo) (id goraptor.Term, err error) {
	id = f.newId("cri")

	if err = f.setType(id, "CreationInfo"); err != nil {
		return
	}

	err = f.addPairs(id,
		pair{"created", cr.Created.V()},
		pair{"rdfs:comment", cr.Comment.V()},
		pair{"licenseListVersion", cr.LicenceListVersion.V()},
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

// Write a slice of reviews.
func (f *Formatter) Reviews(rs []*spdx.Review) error {
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

// Write a package.
func (f *Formatter) Package(pkg *spdx.Package) (id goraptor.Term, err error) {
	id = f.newId("pkg")

	if err = f.setType(id, "Package"); err != nil {
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
	return id, nil
}

func (f *Formatter) VerificationCode(vc *spdx.VerificationCode) (id goraptor.Term, err error) {
	id = f.newId("vc")

	if err = f.setType(id, "PackageVerificationCode"); err != nil {
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

// Write a slice of ExtractedLicensingInfo
func (f *Formatter) ExtrLicInfos(lics []*spdx.ExtractedLicensingInfo) error {
	for _, lic := range lics {
		if err := f.ExtrLicInfo(lic); err != nil {
			return err
		}
	}
	return nil
}

// Write an ExtractedLicensingInfo
func (f *Formatter) ExtrLicInfo(lic *spdx.ExtractedLicensingInfo) error {
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
	f.serializer.Free()
}
