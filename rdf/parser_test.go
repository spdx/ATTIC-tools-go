package rdf

import "testing"

import (
	"errors"
	"github.com/vladvelici/goraptor"
	"github.com/vladvelici/spdx-go/spdx"
)

// Test goraptor term to string
func TestTermStr(t *testing.T) {
	val := "some value"
	if val != termStr(literal(val)) {
		t.Fail()
	}
	if val != termStr(uri(val)) {
		t.Fail()
	}
	if val != termStr(blank(val)) {
		t.Fail()
	}
}

func TestUpd(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.Str("", nil)
	f := upd(&a)
	err := f(blank("world"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.Str("world", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(blank("hello"), nil)
	if err == nil {
		t.Fail()
	}
	if a == spdx.Str("hello", nil) {
		t.Fail()
	}
}

func TestUpdList(t *testing.T) {
	arr := []spdx.ValueStr{spdx.Str("1", nil), spdx.Str("2", nil), spdx.Str("3", nil)}
	f := updList(&arr)
	meta := spdx.NewMeta(5, 7)
	err := f(literal("4"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if len(arr) != 4 || arr[3].Val != "4" || arr[3].M() != meta {
		t.Fail()
	}
}

func TestUpdCutPrefix(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.Str("", nil)
	f := updCutPrefix("prefix_", &a)
	err := f(blank("prefix_world"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.Str("world", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(blank("prefix_hello"), nil)
	if err == nil {
		t.Fail()
	}
	if a == spdx.Str("hello", nil) {
		t.Fail()
	}
}

func TestUpdCreator(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueCreator("", nil)
	f := updCreator(&a)
	err := f(literal("Tool: spdx-go"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.NewValueCreator("Tool: spdx-go", meta)
	if a != expected {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(literal("Person: Mr. Tester"), nil)
	if err == nil {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestUpdListCreator(t *testing.T) {
	arr := []spdx.ValueCreator{spdx.NewValueCreator("1", nil), spdx.NewValueCreator("2", nil), spdx.NewValueCreator("3", nil)}
	f := updListCreator(&arr)
	meta := spdx.NewMeta(5, 7)
	err := f(literal("4"), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if len(arr) != 4 || arr[3].V() != "4" || arr[3].M() != meta {
		t.Fail()
	}
}

func TestUpdDate(t *testing.T) {
	date := "2010-02-03T00:00:00Z"
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueDate("", nil)
	f := updDate(&a)
	err := f(literal(date), meta)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	expected := spdx.NewValueDate(date, meta)
	if a.V() == expected.V() && a.Time() == expected.Time() {
		t.Errorf("Incorrect update. Found %#v but expected %#v.", a, expected)
	}
	err = f(literal("Person: Mr. Tester"), nil)
	if err == nil {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestBuilder(t *testing.T) {
	a := "hello"
	var meta *spdx.Meta
	builder := &builder{t: blank("test"), ptr: &a}
	builder.updaters = map[string]updater{
		"change_value": func(val goraptor.Term, m *spdx.Meta) error {
			a = termStr(val)
			meta = m
			return nil
		},
		"return_error": func(val goraptor.Term, m *spdx.Meta) error {
			return errors.New(termStr(val))
		},
	}
	metaExpected := spdx.NewMeta(3, 4)
	err := builder.apply(uri("change_value"), blank("world"), metaExpected)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if a != "world" {
		t.Errorf("Incorrect update value %#v", metaExpected)
	}
	if meta != metaExpected {
		t.Error("Meta not updated.")
	}
	errorText := "sample error message"
	err = builder.apply(uri("return_error"), blank(errorText), nil)
	if err == nil || err.Error() != errorText {
		t.Errorf("Incorrect error returned %#v", err)
	}
	err = builder.apply(uri("unknown property"), blank(""), nil)
	if err == nil {
		t.Errorf("No error returned for unknown property.")
	}

	if !builder.has("return_error") {
		t.Error(".has does not return true for a known property.")
	}

	if builder.has("this is not known") {
		t.Error(".has does return true for an unknown property.")
	}
}

func TestEqualTypes(t *testing.T) {
	t1 := blank("hello")
	t2 := literal("hola")
	t3 := literal("abc")

	if equalTypes(t1, t2, t3) {
		t.Error("types ok when shouldn't")
	}

	if !equalTypes(t1, blank("hello")) {
		t.Error("types fail")
	}

	if !equalTypes(t1, t2, t3, blank("hello")) {
		t.Error("Iteration test fail.")
	}
}

func TestCompatibleTyeps(t *testing.T) {
	licTypes := []goraptor.Term{
		typeLicence,
		typeDisjunctiveSet,
		typeConjunctiveSet,
		typeExtractedLicence,
		typeAnyLicence,
	}

	abc := blank("abc")
	for _, lic := range licTypes {
		if !compatibleTypes(lic, typeAnyLicence) {
			t.Errorf("Type %s should be compatible with %s", lic, typeAnyLicence)
		}
		if !compatibleTypes(lic, lic) {
			t.Errorf("Type %s should be compatible with itself.", lic)
		}
		if compatibleTypes(lic, abc) {
			t.Errorf("Type %s should not be compatible with %s", lic, abc)
		}
	}
}

type mer interface {
	M() *spdx.Meta
}

func TestSetType(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	// node does not exist, all types.
	// typeAnyLicence will be tested regarding its variants.
	types := map[string]goraptor.Term{
		"Document":           typeDocument,
		"CreationInfo":       typeCreationInfo,
		"Package":            typePackage,
		"File":               typeFile,
		"VerificationCode":   typeVerificationCode,
		"Checksum":           typeChecksum,
		"ArtifactOf":         typeArtifactOf,
		"Review":             typeReview,
		"ExtractedLicence":   typeExtractedLicence,
		"ConjunctiveSet":     typeConjunctiveSet,
		"DisjunctiveSet":     typeDisjunctiveSet,
		"Licence":            typeLicence,
		"AbstractLicenceSet": typeAbstractLicenceSet,
	}

	for k, tp := range types {
		meta := spdx.NewMeta(len(k), len(k)+len(k))
		bldr, err := parser.setType(blank(k), tp, meta)
		if bldr == nil || err != nil {
			t.Errorf("Nil builder: %s. Or error: %s.", bldr, err)
		}
		found, err := parser.reqType(blank(k), tp)
		if err != nil {
			t.Errorf("Error while reqType: %s", err)
		}
		if found != bldr {
			t.Errorf("Different builders. Expected %+v but found %+v.", bldr, found)
		}

		if bldrWithMeta, ok := bldr.(mer); ok {
			if bldrWithMeta.M() != meta {
				t.Errorf("Wrong meta. Expected %#v but found %#v.", meta, bldrWithMeta.M())
			}
		} else {
			t.Errorf("Builder %#v not indexed in parser.", k)
		}
	}

	// incompatible types
	if _, err := parser.setType(blank("Checksum"), typeVerificationCode, nil); err == nil {
		t.Error("Incompatible types didn't return an error.")
	}

	// return existing builder
	if bld, err := parser.setType(blank("Checksum"), typeChecksum, nil); err != nil || bld == nil {
		t.Errorf("Existing builder returned an error (%s) or a nil builder: %+v", err, bld)
	}

	// unknown type
	if _, err := parser.setType(blank("some_unused_name"), blank("this_type_is_unknown"), nil); err == nil {
		t.Error("Unknown type didn't return an error.")
	}
}

// On Uri node, ArtifactOf.ProjectUri must be updated to node's value.
func TestSetTypeArtifactOfUri(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	artifNode := uri("http://spdx.org")
	meta := spdx.NewMeta(3, 4)
	bldr, err := parser.setType(artifNode, typeArtifactOf, meta)
	if err != nil {
		t.Errorf("Unexpected error at ArtifactOf URI: %s", err)
	}
	if bldr != nil {
		artif, ok := bldr.(*spdx.ArtifactOf)
		if !ok {
			t.Errorf("Wrong ArtifactOf type. Found %+v", bldr)
		}
		if artif.Meta != meta {
			t.Errorf("Incorrect meta found for ArtifactOf (with URI): %+v", artif.Meta)
		}
		if artif.ProjectUri.Val != termStr(artifNode) {
			t.Errorf("Incorrect value for ArtifactOfProjectUri: %#v (expected %#v)", artif.ProjectUri.Val, termStr(artifNode))
		}
		if artif.ProjectUri.Meta != meta {
			t.Errorf("Incorrect meta for ArtifactOfProjectUri: %+v (expected %+v)", artif.ProjectUri.Meta, meta)
		}
	}
}

// Special cases when setting the type to AnyLicence.
func TestSetTypeAnyLicence(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	terms := []goraptor.Term{blank("AnyLicenceToSet"), blank("LicenseRef-test"), uri("AnyLicenceInList")}
	typeSlice := []goraptor.Term{typeAbstractLicenceSet, typeLicence, typeLicence}
	for i, term := range terms {
		meta := spdx.NewMeta(i, i+1)
		bldr, err := parser.setType(term, typeAnyLicence, meta)
		if bldr == nil {
			t.Error("Nil builder for typeAnyLicence.")
		}
		if err != nil {
			t.Errorf("Error while setType with typeAnyLicence: %s", err)
		}

		found, err := parser.reqType(term, typeSlice[i])
		if err != nil {
			t.Errorf("Error while reqType with typeAbstractLicenceSet, node type blank: %s", err)
		}
		if found != bldr {
			t.Errorf("Different builders. Expected %+v but found %+v.", bldr, found)
		}
		if bldrWithMeta, ok := bldr.(mer); ok {
			if bldrWithMeta.M() != meta {
				t.Errorf("Wrong meta. Expected %#v but found %#v.", meta, bldrWithMeta.M())
			}
		} else {
			t.Errorf("Builder %#v not indexed in parser.", termStr(term))
		}
	}
}

func TestSetTypeNsType(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}
	// create a fake builder
	fakeVal := "initial"
	fakeBuilder := &builder{
		ptr: &fakeVal,
		t:   typeDocument, // this type is ignored in this use case
	}
	fakeBuilder.updaters = map[string]updater{
		"ns:type": func(term goraptor.Term, meta *spdx.Meta) error {
			if str := termStr(term); str != "error" {
				*(fakeBuilder.ptr.(*string)) = termStr(term)
				return nil
			} else {
				return errors.New("Error")
			}
		},
	}
	parser.index["fakebldr"] = fakeBuilder

	// test applying ns:type
	_, err := parser.setType(blank("fakebldr"), typePackage, nil)
	if err != nil {
		t.Errorf("Unexpected error while applying ns:type: %s", err)
	}
	if fakeVal != termStr(typePackage) {
		t.Errorf("Value didn't change. Found %#v but expected %#v", fakeVal, termStr(typePackage))
	}
	if bldr, err := parser.setType(blank("falekbldr"), uri("error"), nil); err == nil || bldr != nil {
		t.Errorf("Nil error (%s) or non-nil builder (%+v)", err, bldr)
	}
}

func TestSetTypeApplyBuffer(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}
	k := "document"
	stm := &goraptor.Statement{
		Subject:   blank(k),
		Predicate: prefix("specVersion"),
		Object:    literal("SPDX-1.2"),
	}
	meta := spdx.NewMetaL(1)
	parser.buffer[k] = []bufferEntry{{stm, meta}}
	bldr, err := parser.setType(blank(k), typeDocument, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		t.FailNow()
	}

	if doc, ok := bldr.(*spdx.Document); ok {
		if doc.SpecVersion.Val != "SPDX-1.2" {
			t.Errorf("Buffer was not run correctly. Found value %#v but expected \"SPDX-1.2\".", doc.SpecVersion.Val)
		}
		if doc.SpecVersion.Meta != meta {
			t.Errorf("Incorrect meta for buffered value. Found %+v but expected %+v.", doc.SpecVersion.Meta, meta)
		}
	} else {
		t.Errorf("Type was incorrecty set.")
	}
}

func TestDocumentMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	doc := new(spdx.Document)
	builder := parser.documentMap(doc)

	statements := map[string]string{
		"specVersion":               "SPDX-test-1.2",
		"dataLicense":               "CC-0.1",
		"rdfs:comment":              "test comment",
		"creationInfo":              "ci_node",
		"describesPackage":          "pkg_node",
		"referencesFile":            "file_node",
		"reviewed":                  "review_node",
		"hasExtractedLicensingInfo": "extractedLicence_node",
	}

	fakes := map[string]goraptor.Term{
		"ci_node":               typeCreationInfo,
		"pkg_node":              typePackage,
		"file_node":             typeFile,
		"review_node":           typeReview,
		"extractedLicence_node": typeExtractedLicence,
	}

	for k, t := range fakes {
		parser.setType(blank(k), t, nil)
	}

	for k, v := range statements {
		err := builder.apply(blank(k), blank(v), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", k, err)
		}
	}

	testValues := map[string]string{
		"specVersion":  doc.SpecVersion.Val,
		"dataLicense":  doc.DataLicence.Val,
		"rdfs:comment": doc.Comment.Val,
	}

	for k, res := range testValues {
		expected := statements[k]
		if res != expected {
			t.Errorf("Wrong %s. Found %#v (expected #%v)", k, res, expected)
		}
	}

	if doc.CreationInfo == nil {
		t.Error("No CreationInfo found.")
	}
	if len(doc.Packages) == 0 {
		t.Error("No packages.")
	}
	if len(doc.Files) == 0 {
		t.Error("No files.")
	}
	if len(doc.Reviews) == 0 {
		t.Error("No Reviews")
	}
	if len(doc.ExtractedLicences) == 0 {
		t.Error("No Extracted Licences")
	}
}

func TestCreationInfoMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	ci := new(spdx.CreationInfo)
	builder := parser.creationInfoMap(ci)

	statements := map[string]string{
		"creator":            "Person: Testaculous",
		"rdfs:comment":       "test comment",
		"created":            "2014-01-01T09:40:57Z",
		"licenseListVersion": "11",
	}

	for k, v := range statements {
		err := builder.apply(blank(k), blank(v), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", k, err)
		}
	}

	testValues := map[string]string{
		"creator":            ci.Creator[0].V(),
		"rdfs:comment":       ci.Comment.V(),
		"created":            ci.Created.V(),
		"licenseListVersion": ci.LicenceListVersion.V(),
	}

	for k, res := range testValues {
		expected := statements[k]
		if res != expected {
			t.Errorf("Wrong %s. Found %#v (expected #%v)", k, res, expected)
		}
	}
}

func TestPackageMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	pkg := new(spdx.Package)
	builder := parser.packageMap(pkg)

	statements := map[string]string{
		"name":                    "pkg-test",
		"versionInfo":             "pkg-test-v1",
		"packageFileName":         "pkg-test-file.zip",
		"supplier":                "Person: Testaculous",
		"originator":              "spdx.org",
		"downloadLocation":        "http://spdx.org/",
		"packageVerificationCode": "verif_node",
		"checksum":                "cksum_node",
		"doap:homepage":           "http://spdx.org/",
		"sourceInfo":              "some src info",
		"licenseConcluded":        "lic_concluded_node",
		"licenseInfoFromFiles":    "lic_info_from_file_node",
		"licenseDeclared":         "lic_declared_node",
		"licenseComments":         "some licence comments",
		"copyrightText":           "some copyright text",
		"summary":                 "this package is awesome",
		"description":             "this is the most awesome package ever",
		"hasFile":                 "file_node",
	}

	fakes := map[string]goraptor.Term{
		"verif_node":              typeVerificationCode,
		"cksum_node":              typeChecksum,
		"lic_concluded_node":      typeLicence,
		"lic_info_from_file_node": typeLicence,
		"lic_declared_node":       typeLicence,
		"file_node":               typeFile,
	}

	for k, t := range fakes {
		parser.setType(blank(k), t, nil)
	}

	for k, v := range statements {
		err := builder.apply(blank(k), blank(v), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", k, err)
		}
	}

	testValues := map[string]string{
		"name":             pkg.Name.Val,
		"versionInfo":      pkg.Version.Val,
		"packageFileName":  pkg.FileName.Val,
		"supplier":         pkg.Supplier.V(),
		"originator":       pkg.Originator.V(),
		"downloadLocation": pkg.DownloadLocation.Val,
		"doap:homepage":    pkg.HomePage.Val,
		"sourceInfo":       pkg.SourceInfo.Val,
		"licenseComments":  pkg.LicenceComments.Val,
		"copyrightText":    pkg.CopyrightText.Val,
		"summary":          pkg.Summary.Val,
		"description":      pkg.Description.Val,
	}

	for k, res := range testValues {
		expected := statements[k]
		if res != expected {
			t.Errorf("Wrong %s. Found %#v (expected #%v)", k, res, expected)
		}
	}

	if pkg.VerificationCode == nil {
		t.Error("No verification code.")
	}
	if pkg.Checksum == nil {
		t.Error("No checksum.")
	}
	if pkg.LicenceDeclared == nil {
		t.Error("No licence declared.")
	}
	if pkg.LicenceConcluded == nil {
		t.Error("No licence concluded.")
	}
	if len(pkg.LicenceInfoFromFiles) == 0 {
		t.Error("No licence info from files.")
	}
	if len(pkg.Files) == 0 {
		t.Error("No files.")
	}
}

func TestFileMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	file := new(spdx.File)
	builder := parser.fileMap(file)

	statements := map[string]string{
		"fileName":          "parser_test.go",
		"rdfs:comment":      "some test comment",
		"fileType":          "SOURCE",
		"checksum":          "cksum_node",
		"copyrightText":     "some copyright text",
		"noticeText":        "some notice text",
		"licenseConcluded":  "lic_concluded_node",
		"licenseInfoInFile": "lic_info_in_file_node",
		"licenseComments":   "some licence comments",
		"fileContributor":   "Person: Testaculous",
		"fileDependency":    "file_node",
		"artifactOf":        "artif_node",
	}

	fakes := map[string]goraptor.Term{
		"cksum_node":            typeChecksum,
		"lic_concluded_node":    typeDisjunctiveSet,
		"lic_info_in_file_node": typeConjunctiveSet,
		"file_node":             typeFile,
		"artif_node":            typeArtifactOf,
	}

	for k, t := range fakes {
		parser.setType(blank(k), t, nil)
	}

	for k, v := range statements {
		err := builder.apply(blank(k), blank(v), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", k, err)
		}
	}

	testValues := map[string]string{
		"fileName":        file.Name.Val,
		"rdfs:comment":    file.Comment.Val,
		"fileType":        file.Type.Val,
		"copyrightText":   file.CopyrightText.Val,
		"noticeText":      file.Notice.Val,
		"licenseComments": file.LicenceComments.Val,
		"fileContributor": file.Contributor[0].V(),
	}

	for k, res := range testValues {
		expected := statements[k]
		if res != expected {
			t.Errorf("Wrong %s. Found %#v (expected #%v)", k, res, expected)
		}
	}

	if file.Checksum == nil {
		t.Error("No checksum.")
	}
	if file.LicenceConcluded == nil {
		t.Error("No licence concluded.")
	}
	if len(file.LicenceInfoInFile) == 0 {
		t.Error("No licence info in file.")
	}
	if len(file.Dependency) == 0 {
		t.Error("No dependecies.")
	}
}

func TestChecksumMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	cksum := new(spdx.Checksum)
	builder := parser.checksumMap(cksum)

	statements := map[string]string{
		"algorithm":     "http://spdx.org/rdf/terms#checksumAlgorithm_sha1",
		"checksumValue": "somedummyvalue",
	}

	for k, v := range statements {
		err := builder.apply(blank(k), blank(v), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", k, err)
		}
	}

	if cksum.Algo.Val != "SHA1" {
		t.Errorf("Wrong algorithm. Expected \"SHA1\" but found %#v", cksum.Algo.Val)
	}
	if cksum.Value.Val != statements["checksumValue"] {
		t.Errorf("Wrong value. Expected %#v but found %#v", statements["checksumValue"], cksum.Algo.Val)
	}

}

func TestLicenceSetMap(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	set := new(spdx.LicenceSet)
	builder := parser.licenceSetMap(set)

	parser.index["testnode"] = builder

	statements := []pair{
		{"member", "lic1"},
		{"member", "lic2"},
		{"ns:type", termStr(typeConjunctiveSet)},
	}

	fakes := map[string]goraptor.Term{
		"lic1": typeLicence,
		"lic2": typeLicence,
	}

	for k, t := range fakes {
		parser.setType(blank(k), t, nil)
	}

	for _, pair := range statements {
		err := builder.apply(blank(pair.key), uri(pair.val), nil)
		if err != nil {
			t.Errorf("Applying %s. Unexpected error: %s", pair.key, err)
		}
	}

	if len(set.Members) != 2 {
		t.Errorf("Wrong licence set members. Found: %#v", set.Members)
	}

	newSet, err := parser.reqAnyLicence(blank("testnode"))
	if err != nil {
		t.Errorf("Couldn't get AnyLicence: %s", err)
	}
	if _, ok := newSet.(spdx.ConjunctiveLicenceSet); !ok {
		t.Errorf("Set is of the wrong type or type didn't change. %#v", newSet)
	}
}

func TestProcessTruple(t *testing.T) {
	parser := &Parser{
		index:  make(map[string]*builder),
		buffer: make(map[string][]bufferEntry),
	}

	statements := []*goraptor.Statement{
		{
			Subject:   blank("document"),
			Predicate: prefix("ns:type"),
			Object:    typeDocument,
		},
		{
			Subject:   blank("document"),
			Predicate: uri("specVersion"),
			Object:    literal("SPDX-1.2"),
		},
		{
			Subject:   blank("pkg"),
			Predicate: uri("packageFileName"),
			Object:    literal("pkgfile.zip"),
		},
		{
			Subject:   blank("pkg"),
			Predicate: prefix("ns:type"),
			Object:    typePackage,
		},
	}

	for i, stm := range statements {
		err := parser.processTruple(stm, spdx.NewMetaL(i+1))
		if err != nil {
			t.Errorf("Unexpected error while processing %#v: %s", *stm, err)
		}
	}

	doc, err := parser.reqDocument(blank("document"))
	if err != nil {
		t.Errorf("Unexpected error while getting document: %s", err)
		t.FailNow()
	}
	if doc.Meta.LineStart != 1 {
		t.Errorf("Wrong meta linestart at document. Found %d (expected %d)", doc.Meta.LineStart, 1)
	}
	sv := spdx.Str("SPDX-1.2", spdx.NewMetaL(2))
	if doc.SpecVersion.Val != sv.Val && doc.SpecVersion.Meta.LineStart == sv.Meta.LineStart {
		t.Errorf("Wrong SpecVersion. Found %#v (expected %#v)", doc.SpecVersion, sv)
	}

	pkg, err := parser.reqPackage(blank("pkg"))
	if err != nil {
		t.Errorf("Unexpected error while getting package: %s", err)
		t.FailNow()
	}
	if pkg.Meta.LineStart != 4 {
		t.Errorf("Wrong meta linestart at package. Found %d (expected %d)", pkg.Meta.LineStart, 4)
	}
	pkgFile := spdx.Str("pkgfile.zip", spdx.NewMetaL(3))
	if pkg.FileName.Val != pkgFile.Val && pkg.FileName.Meta.LineStart == pkgFile.Meta.LineStart {
		t.Errorf("Wrong SpecVersion. Found %#v (expected %#v)", pkg.FileName, pkgFile)
	}
}
