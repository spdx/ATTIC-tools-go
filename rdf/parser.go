package rdf

import (
	"fmt"
	"github.com/deltamobile/goraptor"
	"github.com/vladvelici/spdx-go/spdx"
	"io"
)

var (
	uri_nstype = uri("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")

	typeDocument               = prefix("SpdxDocument")
	typeCreationInfo           = prefix("CreationInfo")
	typePackage                = prefix("Package")
	typeFile                   = prefix("File")
	VerificationCode           = prefix("PackageVerificationCode")
	typeChecksum               = prefix("Checksum")
	typeReview                 = prefix("Review")
	typeExtractedLicensingInfo = prefix("ExtractedLicensingInfo")
	typeAnyLicenceInfo         = prefix("AnyLicenseInfo")
	typeConjunctiveSet         = prefix("ConjunctiveLicenseSet")
	typeDisjunctiveSet         = prefix("DisjunctiveLicenseSet")
)

const (
	msgIncompatibleTypes    = "%s is already set to be type %s and cannot be changed to type %s."
	msgIncompatibleTypesGo  = "Node %s has type %T but expected %s."
	msgPropertyNotSupported = "Property %s is not supported for %s."
	msgAlreadyDefined       = "Property already defined."
	msgUnknownType          = "Found type %s which is unknown."
)

// Simple, one function call interface to parse a document
func Parse(input io.Reader, format string) (*spdx.Document, error) {
	parser := NewParser(input, format)
	defer parser.Free()
	return parser.Parse()
}

// goraptor.Term to string
func termStr(term goraptor.Term) string {
	switch t := term.(type) {
	case *goraptor.Uri:
		return string(*t)
	case *goraptor.Blank:
		return string(*t)
	case *goraptor.Literal:
		return t.Value
	default:
		return ""
	}
}

// Update a ValString pointer
func upd(ptr *spdx.ValueStr) updater {
	set := false
	return func(term goraptor.Term) error {
		if set {
			return fmt.Errorf(msgAlreadyDefined)
		}

		ptr.Val = termStr(term)
		set = true
		return nil
	}
}

// Update a []ValString pointer
func updList(arr *[]spdx.ValueStr) updater {
	return func(term goraptor.Term) error {
		*arr = append(*arr, spdx.Str(termStr(term), nil))
		return nil
	}
}

// Update a ValueCreator pointer
func updCreator(ptr *spdx.ValueCreator) updater {
	set := false
	return func(term goraptor.Term) error {
		if set {
			return fmt.Errorf(msgAlreadyDefined)
		}
		ptr.SetValue(termStr(term))
		set = true
		return nil
	}
}

// Update a ValueDate pointer
func updDate(ptr *spdx.ValueDate) updater {
	set := false
	return func(term goraptor.Term) error {
		if set {
			return fmt.Errorf(msgAlreadyDefined)
		}
		ptr.SetValue(termStr(term))
		set = true
		return nil
	}
}

// Update a []ValueCreator pointer
func updListCreator(arr *[]spdx.ValueCreator) updater {
	return func(term goraptor.Term) error {
		*arr = append(*arr, spdx.NewValueCreator(termStr(term), nil))
		return nil
	}
}

type builder struct {
	t        goraptor.Term // type of element this builder represents
	ptr      interface{}   // the spdx element that this builder builds
	updaters map[string]updater
}

func (b *builder) apply(pred, obj goraptor.Term) error {
	property := termStr(pred)
	f, ok := b.updaters[termStr(pred)]
	if !ok {
		return fmt.Errorf(msgPropertyNotSupported, property, b.t)
	}
	return f(obj)
}

type updater func(goraptor.Term) error

type Parser struct {
	rdfparser *goraptor.Parser
	input     io.Reader
	index     map[string]*builder
	buffer    map[string][]*goraptor.Statement
	doc       *spdx.Document
}

// This creates a goraptor.Parser object that needs to be freed after use.
// Call Parser.Free() after using the Parser.
func NewParser(input io.Reader, format string) *Parser {
	if format == "rdf" {
		format = "guess"
	}

	return &Parser{
		rdfparser: goraptor.NewParser(format),
		input:     input,
		index:     make(map[string]*builder),
		buffer:    make(map[string][]*goraptor.Statement),
	}
}

// Parse the whole input stream and return the resulting spdx.Document or the first error that occurred.
func (p *Parser) Parse() (*spdx.Document, error) {
	ch := p.rdfparser.Parse(p.input, baseUri)
	defer func() {
		// consume the channel if there's anything left.
		for _ = range ch {
		}
	}()
	var err error
	for statement := range ch {
		if err = p.processTruple(statement); err != nil {
			break
		}
	}
	for {
		_, ok := <-ch
		if !ok {
			break
		}
	}
	return p.doc, err
}

// Free the goraptor parser.
func (p *Parser) Free() {
	p.rdfparser.Free()
	p.doc = nil
}

func (p *Parser) setType(node string, t goraptor.Term) (interface{}, error) {
	bldr, ok := p.index[node]
	if ok {
		if bldr.t.Equals(t) {
			return nil, fmt.Errorf(msgIncompatibleTypes, node, bldr.t, t)
		}
		return bldr.ptr, nil
	}

	// new builder by type
	switch {
	case t.Equals(typeDocument):
		p.doc = new(spdx.Document)
		bldr = p.documentMap(p.doc)
	default:
		return nil, fmt.Errorf(msgUnknownType, t)
	}

	p.index[node] = bldr

	// run buffer
	buf := p.buffer[node]
	for _, stm := range buf {
		if err := bldr.apply(stm.Predicate, stm.Object); err != nil {
			return nil, err
		}
	}
	delete(p.buffer, node)

	return bldr.ptr, nil
}

func (p *Parser) processTruple(stm *goraptor.Statement) error {
	node := termStr(stm.Subject)
	if stm.Predicate.Equals(uri_nstype) {
		_, err := p.setType(node, stm.Object)
		return err
	}

	// apply function if it's a builder
	bldr, ok := p.index[node]
	if ok {
		return bldr.apply(stm.Predicate, stm.Object)
	}

	// buffer statement
	if _, ok := p.buffer[node]; !ok {
		p.buffer[node] = make([]*goraptor.Statement, 0)
	}
	p.buffer[node] = append(p.buffer[node], stm)

	return nil
}

// Parser.req* functions are supposded to get the node from either the index or the buffer,
// check if it's the required type and return a pointer to the relevant spdx.* object.

func (p *Parser) reqSomething(node string, t goraptor.Term) (interface{}, error) {
	bldr, ok := p.index[node]
	if ok {
		if !bldr.t.Equals(t) {
			return nil, fmt.Errorf(msgIncompatibleTypes, node, bldr.t, t)
		}
		return bldr.ptr, nil
	}
	return p.setType(node, t)
}

func (p *Parser) reqDocument(node string) (*spdx.Document, error) {
	obj, err := p.reqSomething(node, typeDocument)
	return obj.(*spdx.Document), err
}
func (p *Parser) reqCreationInfo(node string) (*spdx.CreationInfo, error) {
	obj, err := p.reqSomething(node, typeCreationInfo)
	return obj.(*spdx.CreationInfo), err
}
func (p *Parser) reqPackage(node string) (*spdx.Package, error) {
	obj, err := p.reqSomething(node, typePackage)
	return obj.(*spdx.Package), err
}
func (p *Parser) reqFile(node string) (*spdx.File, error) {
	obj, err := p.reqSomething(node, typeFile)
	return obj.(*spdx.File), err
}
func (p *Parser) reqVerificationCode(node string) (*spdx.VerificationCode, error) {
	obj, err := p.reqSomething(node, VerificationCode)
	return obj.(*spdx.VerificationCode), err
}
func (p *Parser) reqChecksum(node string) (*spdx.Checksum, error) {
	obj, err := p.reqSomething(node, typeChecksum)
	return obj.(*spdx.Checksum), err
}
func (p *Parser) reqReview(node string) (*spdx.Review, error) {
	obj, err := p.reqSomething(node, typeReview)
	return obj.(*spdx.Review), err
}
func (p *Parser) reqExtractedLicensingInfo(node string) (*spdx.ExtractedLicensingInfo, error) {
	obj, err := p.reqSomething(node, typeExtractedLicensingInfo)
	return obj.(*spdx.ExtractedLicensingInfo), err
}
func (p *Parser) reqAnyLicenceInfo(node string) (*spdx.AnyLicenceInfo, error) {
	obj, err := p.reqSomething(node, typeAnyLicenceInfo)
	return obj.(*spdx.AnyLicenceInfo), err
}

func (p *Parser) documentMap(doc *spdx.Document) *builder {
	bldr := &builder{t: typeDocument, ptr: doc}
	bldr.updaters = map[string]updater{
		"specVersion":  upd(&doc.SpecVersion),
		"dataLicense":  upd(&doc.DataLicence),
		"rdfs:comment": upd(&doc.Comment),
		"creationInfo": func(obj goraptor.Term) error {
			cri, err := p.reqCreationInfo(termStr(obj))
			if err != nil {
				return err
			}
			doc.CreationInfo = cri
			return nil
		},
	}

	return bldr
}

func (p *Parser) creationInfoMap(cri *spdx.CreationInfo) *builder {
	bldr := &builder{t: typeCreationInfo, ptr: cri}
	bldr.updaters = map[string]updater{
		"creator":            updListCreator(&cri.Creator),
		"rdfs:comment":       upd(&cri.Comment),
		"created":            updDate(&cri.Created),
		"licenseListVersion": upd(&cri.LicenceListVersion),
	}
	return bldr
}
