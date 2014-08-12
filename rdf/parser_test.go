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

	// create a fake builder
	fakeVal := "initial"
	fakeBuilder := &builder{
		ptr: &fakeVal,
		t:   typeDocument, // the type is ignored anyway
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

	// on Uri node, ArtifactOf.ProjectUri must be updated to node's value
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

	// AnyLicence varieties
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

	// unknown type
	if _, err := parser.setType(blank("some_unused_name"), blank("this_type_is_unknown"), nil); err == nil {
		t.Error("Unknown type didn't return an error.")
	}

}
