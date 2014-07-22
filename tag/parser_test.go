package tag

import (
	"github.com/vladvelici/spdx-go/spdx"
	"testing"
)

func sameStrSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameValStrValues(a []spdx.ValueStr, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Val != b[i] {
			return false
		}
	}
	return true
}

func sameValCreatValues(a []spdx.ValueCreator, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].V() != b[i] {
			return false
		}
	}
	return true
}

func sameValSlice(a, b []spdx.Value) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Fake lexer for testing, no meta
type testLexer struct {
	i     int
	pairs []Pair
}

func (l *testLexer) Lex() bool     { l.i++; return l.i < len(l.pairs) }
func (l *testLexer) Token() *Token { return &Token{TokenPair, l.pairs[l.i], nil} }
func (l *testLexer) Err() error    { return nil }
func (l *testLexer) Line() int     { return 0 }
func l(p []Pair) lexer             { return &testLexer{-1, p} }

// Fake lexer for testing, with meta support
type testLexerTok struct {
	i      int
	tokens []*Token
	line   int
}

func (l *testLexerTok) Lex() bool {
	l.i++
	res := l.i < len(l.tokens)
	if res {
		if meta := l.tokens[l.i].Meta; meta != nil {
			l.line = meta.LineStart
		}
	}
	return res
}
func (l *testLexerTok) Token() *Token { l.tokens[l.i].Type = TokenPair; return l.tokens[l.i] }
func (l *testLexerTok) Err() error    { return nil }
func (l *testLexerTok) Line() int     { return l.line }
func lt(toks []*Token) lexer          { return &testLexerTok{-1, toks, 0} }

// Create a token from a string or string and meta.
// If a meta is given, it is used as the token's meta. Only the first meta given is used, others are ignored.
func tk(value string, metas ...*spdx.Meta) *Token {
	var meta *spdx.Meta
	if len(metas) > 0 {
		meta = metas[0]
	}
	return &Token{Pair: Pair{Value: value}, Meta: meta}
}

func sameSpdx(t *testing.T, found, expected spdx.Value) {
	if found.V() != expected.V() {
		t.Errorf("Incorrect value. Found \"%s\" but expected \"%s\"", found.V(), expected.V())
	}
	if found.M() != expected.M() {
		t.Errorf("Incorrect meta. Found %#v but expected %#v", found.M(), expected.M())
	}
}

func TestUpd(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.Str("", nil)
	f := upd(&a)
	err := f(tk("world", meta))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	sameSpdx(t, a, spdx.Str("world", meta))
}

func TestUpdFail(t *testing.T) {
	a := spdx.Str("", nil)
	f := upd(&a)
	f(tk("hello"))
	err := f(tk("world"))
	if err.Error() != MsgAlreadyDefined {
		t.Errorf("Different error: %s", err)
	}
}

func TestUpdDelay(t *testing.T) {
	called := false
	val := spdx.Str("", nil)
	delay := func(tok *Token) *spdx.ValueStr {
		if called {
			t.Error("UpdDelay - delay function called more than once")
		}
		called = true
		return &val
	}
	f := updDelay(delay)
	meta := spdx.NewMeta(3, 4)
	err := f(tk("hello", meta))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if !called {
		t.Errorf("Delay function not called.")
	}
	sameSpdx(t, val, spdx.Str("hello", meta))
	err = f(tk("world"))
	if err != nil && val.Val == "world" {
		t.Error("No update should have happened.")
	}
	if err == nil {
		t.Error("No error when there should be one.")
	}
	f(tk("third test"))
}

func TestUpdList(t *testing.T) {
	arr := []spdx.ValueStr{spdx.Str("1", nil), spdx.Str("2", nil), spdx.Str("3", nil)}
	f := updList(&arr)
	meta := spdx.NewMeta(5, 7)
	err := f(tk("4", meta))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if len(arr) != 4 || arr[3].Val != "4" || arr[3].M() != meta {
		t.Fail()
	}
}

func TestUpdListDelay(t *testing.T) {
	arr := make([]spdx.ValueStr, 0)
	called := false
	f := updListDelay(func(tok *Token) *[]spdx.ValueStr {
		if called {
			t.Error("Delay function called more than once.")
		}
		called = true
		return &arr
	})
	f(tk("zero"))
	if !called {
		t.Error("Delay function not called.")
	}
	f(tk("one"))
	f(tk("two"))
}

// Test upd*Creator
// Testing the actual parsing of the 'creator' format should happen in the spdx package

func TestUpdCreator(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueCreator("", nil)
	f := updCreator(&a)
	err := f(tk("Tool: spdx-go", meta))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	sameSpdx(t, a, spdx.NewValueCreator("Tool: spdx-go", meta))
	err = f(tk("Person: Mr. Tester", nil))
	if err == nil || err.Error() != MsgAlreadyDefined {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestUpdCreatorDelay(t *testing.T) {
	called := false
	val := spdx.NewValueCreator("", nil)
	f := updCreatorDelay(func(tok *Token) *spdx.ValueCreator {
		if called {
			t.Error("UpdDelay - delay function called more than once")
		}
		called = true
		return &val
	})
	meta := spdx.NewMeta(3, 4)
	err := f(tk("Tool: spdx-go", meta))
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}
	if !called {
		t.Errorf("Delay function not called.")
	}
	sameSpdx(t, val, spdx.NewValueCreator("Tool: spdx-go", meta))
	err = f(tk("world"))
	if err != nil && val.V() == "world" {
		t.Error("No update should have happened.")
	}
	if err == nil {
		t.Error("No error when there should be one.")
	}
	f(tk("third test"))
}

func TestUpdCreatorListDelay(t *testing.T) {
	arr := make([]spdx.ValueCreator, 0)
	called := false
	f := updCreatorListDelay(func(tok *Token) *[]spdx.ValueCreator {
		if called {
			t.Error("Delay function called more than once.")
		}
		called = true
		return &arr
	})
	f(tk("zero"))
	if !called {
		t.Error("Delay function not called.")
	}
	f(tk("one"))
	f(tk("two"))
}

// Test upd*Date

func TestUpdDate(t *testing.T) {
	meta := spdx.NewMeta(3, 4)
	a := spdx.NewValueDate("", nil)
	f := updDate(&a)
	date := "2010-02-03T00:00:00Z"
	err := f(tk(date, meta))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	sameSpdx(t, a, spdx.NewValueDate(date, meta))
	err = f(tk("should not update", nil))
	if err == nil || err.Error() != MsgAlreadyDefined {
		t.Errorf("Incorrect error %+v", err)
	}
}

func TestUpdDateDelay(t *testing.T) {
	called := false
	val := spdx.NewValueDate("", nil)
	f := updDateDelay(func(tok *Token) *spdx.ValueDate {
		if called {
			t.Error("UpdDelay - delay function called more than once")
		}
		called = true
		return &val
	})
	date := "2010-02-03T00:00:00Z"
	meta := spdx.NewMeta(3, 4)
	err := f(tk(date, meta))
	if err != nil {
		t.Errorf("Unexpected error %+v", err)
	}
	if !called {
		t.Errorf("Delay function not called.")
	}
	sameSpdx(t, val, spdx.NewValueDate(date, meta))
	err = f(tk("should not update"))
	if err != nil && val.V() == "should not update" {
		t.Error("No update should have happened.")
	}
	if err == nil {
		t.Error("No error when there should be one.")
	}
	f(tk("third test"))
}

func TestVerifCodeNoExcludes(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	f := updVerifCode(vc)
	err := f(tk(value))
	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}
	if vc.Value.Val != value {
		t.Error("Verification code value different than given value")
	}

	if len(vc.ExcludedFiles) != 0 {
		t.Errorf("ExcludedFiles should be null/empty but found %s.", vc.ExcludedFiles)
	}
}

func TestVerifCodeWithExcludes(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	excludes := " (excludes: abc.txt, file.spdx)"
	f := updVerifCode(vc)
	err := f(tk(value + excludes))

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}
	if vc.Value.Val != value {
		t.Error("Verification code value different than given value")
	}

	if !sameValStrValues(vc.ExcludedFiles, []string{"abc.txt", "file.spdx"}) {
		t.Error("Different ExcludedFiles. Elements found: %s.", vc.ExcludedFiles)
	}
}

func TestVerifCodeWithExcludes2(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	excludes := " (abc.txt, file.spdx)"
	f := updVerifCode(vc)

	err := f(tk(value + excludes))

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}

	if vc.Value.Val != value {
		t.Error("Verification code value different than given value")
	}

	if !sameValStrValues(vc.ExcludedFiles, []string{"abc.txt", "file.spdx"}) {
		t.Error("Different ExcludedFiles. Elements found: %s.", vc.ExcludedFiles)
	}
}

func TestVerifCodeInvalidExcludesNoClosedParentheses(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	excludes := " ("
	f := updVerifCode(vc)
	err := f(tk(value + excludes))

	if err.Error() != MsgNoClosedParen {
		t.Errorf("Error should be UnclosedParentheses but found %s.", err)
	}
}

func TestChecksum(t *testing.T) {
	cksum := new(spdx.Checksum)
	val := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	algo := "SHA1"

	f := updChecksum(cksum)
	err := f(tk(algo + ": " + val))

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}

	if cksum.Value.Val != val {
		t.Errorf("Checksum value is wrong. Found: '%s'. Expected: '%s'.", cksum.Value.Val, val)
	}

	if cksum.Algo.Val != algo {
		t.Errorf("Algo is wrong. Found: '%s'. Expected: '%s'.", cksum.Algo.Val, algo)
	}
}

func TestChecksumInvalid(t *testing.T) {
	cksum := new(spdx.Checksum)
	f := updChecksum(cksum)
	err := f(tk("d6a770ba38583ed4bb"))

	if err.Error() != MsgInvalidChecksum {
		t.Errorf("Invalid error found: %s.", err)
	}
}

// findMatchingParenSet tests:

func TestFindMachingParen(t *testing.T) {
	input := "a(f())d"
	//        0123456
	o, c := findMatchingParenSet(input)
	if o != 1 || c != 5 {
		t.Errorf("Expected o=1 and c=5 but have o=%d and c=%d", o, c)
	}
}

func TestFindMachingParenUnbalanced(t *testing.T) {
	input := "a(f()d"
	//        012345
	o, c := findMatchingParenSet(input)
	if o != 1 || c >= 0 {
		t.Errorf("Expected o=1 and c<0 but have o=%d and c=%d", o, c)
	}
}

func TestFindMachingParenNextElem(t *testing.T) {
	input := "a()"
	//        012
	o, c := findMatchingParenSet(input)
	if o != 1 || c != 2 {
		t.Errorf("Expected o=1 and c=2 but have o=%d and c=%d", o, c)
	}
}

func TestFindMachingParenNoParenthesis(t *testing.T) {
	input := "blahblah"
	o, c := findMatchingParenSet(input)
	if o >= 0 || c >= 0 || c >= o {
		t.Errorf("Expected c<o<0 but have o=%d and c=%d", o, c)
	}
}

func TestConjDisjAND(t *testing.T) {
	input := "a and b"
	conj, disj := conjOrDisjSet(input)
	if conj != true || disj != false {
		t.Errorf("Expected cojunctive. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestConjDisjOR(t *testing.T) {
	input := "a or b"
	conj, disj := conjOrDisjSet(input)
	if conj != false || disj != true {
		t.Errorf("Expected disjunctive. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestConjDisjBOTH(t *testing.T) {
	input := "a and b or c"
	conj, disj := conjOrDisjSet(input)
	if conj != true || disj != true {
		t.Errorf("Expected conjunctive and disjunctive. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestConjDisjParenOR(t *testing.T) {
	input := "(a and b) or c"
	conj, disj := conjOrDisjSet(input)
	if conj != false || disj != true {
		t.Errorf("Expected disjunctive. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestConjDisjMultipleParen(t *testing.T) {
	input := "(a and b and (c or d)) or (a or d)"
	conj, disj := conjOrDisjSet(input)
	if conj != false || disj != true {
		t.Errorf("Expected disjunctive. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestConjDisjNone(t *testing.T) {
	input := "GPLv3"
	conj, disj := conjOrDisjSet(input)
	if conj != false || disj != false {
		t.Errorf("Expected neither. Found: conj=%b, disj=%b", conj, disj)
	}
}

func TestLicenceSetSplit(t *testing.T) {
	input := "a and (b or c)"
	expected := []string{"a", "(b or c)"}
	output := licenceSetSplit(andSeprator, input)

	if !sameStrSlice(expected, output) {
		t.Errorf("Expected %s but found %s", expected, output)
	}
}

func TestLicenceSetSplitSameTypeInParen(t *testing.T) {
	input := "a and (b and c) and d"
	expected := []string{"a", "(b and c)", "d"}
	output := licenceSetSplit(andSeprator, input)

	if !sameStrSlice(expected, output) {
		t.Errorf("Expected %s but found %s", expected, output)
	}
}

func TestLicenceSetSplitNoParen(t *testing.T) {
	input := "a and b  and  c "
	expected := []string{"a", "b", "c"}
	output := licenceSetSplit(andSeprator, input)

	if !sameStrSlice(expected, output) {
		t.Errorf("Expected %s but found %s", expected, output)
	}
}

func TestLicenceSetSplitNoSeparator(t *testing.T) {
	input := "a"
	expected := []string{"a"}
	output := licenceSetSplit(andSeprator, input)

	if !sameStrSlice(expected, output) {
		t.Errorf("Expected %s but found %s", expected, output)
	}
}

func TestParseLicenceSetOr(t *testing.T) {
	input := "(GPLv3 or LicenseRef-1)"
	expected := spdx.NewDisjunctiveSet(nil, spdx.NewLicence("GPLv3", nil), spdx.NewLicence("LicenseRef-1", nil))
	output, err := parseLicenceSet(tk(input))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !spdx.SameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetAnd(t *testing.T) {
	input := "(GPLv3 and LicenseRef-1)"
	expected := spdx.NewConjunctiveSet(nil, spdx.NewLicence("GPLv3", nil), spdx.NewLicence("LicenseRef-1", nil))
	output, err := parseLicenceSet(tk(input))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !spdx.SameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetNested(t *testing.T) {
	input := "(GPLv3 or (LicenseRef-1 and LicenseRef-3) or LicenseRef-2)"

	expected := spdx.NewDisjunctiveSet(nil,
		spdx.NewLicence("GPLv3", nil),
		spdx.NewConjunctiveSet(nil,
			spdx.NewLicence("LicenseRef-1", nil),
			spdx.NewLicence("LicenseRef-3", nil),
		),
		spdx.NewLicence("LicenseRef-2", nil),
	)

	output, err := parseLicenceSet(tk(input))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !spdx.SameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetSingleValue(t *testing.T) {
	input := "one value "

	expected := spdx.NewLicence("one value", nil)

	output, err := parseLicenceSet(tk(input))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !spdx.SameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetEmptyLicence(t *testing.T) {
	input := " "

	_, err := parseLicenceSet(tk(input))
	if err.Error() != MsgEmptyLicence {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseLicenceStringEmptyLicence(t *testing.T) {
	input := " "

	_, err := parseLicenceString(tk(input))
	if err.Error() != MsgEmptyLicence {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseLicenceStringUnbalancedParentheses(t *testing.T) {
	input := " (()"

	_, err := parseLicenceString(tk(input))
	if err.Error() != MsgNoClosedParen {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestLicenceSetConjunctionAndDisjunction(t *testing.T) {
	input := "a and b or c"
	_, err := parseLicenceSet(tk(input))
	if err.Error() != MsgConjunctionAndDisjunction {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDoc(t *testing.T) {
	m := map[string][]string{
		"SPDXVersion":        {"1.2"},
		"DataLicense":        {"CC-1.0"},
		"DocumentComment":    {"hahaha hahaha"},
		"Creator":            {"Organization: F", "Person: Test Testaculous"},
		"Created":            {"04/05/06"},
		"CreatorComment":     {"This is some creator comment...\n blah blah"},
		"LicenseListVersion": {"1.2.3"},
	}

	// transform data to input pair
	input := make([]Pair, 0)
	for k, vals := range m {
		for _, v := range vals {
			input = append(input, Pair{k, v})
		}
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if doc.SpecVersion.Val != m["SPDXVersion"][0] {
		t.Errorf("h Invalid doc.SpecVersion: '%+v'", doc.SpecVersion)
	}
	if doc.DataLicence.Val != m["DataLicense"][0] {
		t.Errorf("Invalid doc.DataLicence: '%+v'", doc.DataLicence)
	}
	if doc.Comment.Val != m["DocumentComment"][0] {
		t.Errorf("Invalid doc.Comment: '%+v'", doc.Comment)
	}
	if !sameValCreatValues(doc.CreationInfo.Creator, m["Creator"]) {
		t.Errorf("Invalid doc.CreationInfo.Creator: (len=%d) '%+v'", len(doc.CreationInfo.Creator), doc.CreationInfo.Creator)
	}
	if doc.CreationInfo.Created.V() != m["Created"][0] {
		t.Errorf("Invalid doc.CreationInfo.Created: '%+v'", doc.CreationInfo.Created)
	}
	if doc.CreationInfo.Comment.Val != m["CreatorComment"][0] {
		t.Errorf("Invalid doc.CreationInfo.Comment: '%+v'", doc.CreationInfo.Comment)
	}
	if doc.CreationInfo.LicenceListVersion.Val != m["LicenseListVersion"][0] {
		t.Errorf("Invalid doc.LicenceListVersion: '%+v'", doc.CreationInfo.LicenceListVersion)
	}
}

func TestSamePropertyTwice(t *testing.T) {
	input := []Pair{
		{"SPDXVersion", "1.2"},
		{"SPDXVersion", "1.3"},
	}

	_, err := Parse(l(input))

	if err.Error() != MsgAlreadyDefined {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDocNestedPackage(t *testing.T) {
	cksum := spdx.Checksum{
		Algo:  spdx.Str("SHA1", nil),
		Value: spdx.Str("2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", nil),
	}
	input := []Pair{
		{"SPDXVersion", "1.2"},
		{"PackageName", "spdx-tools-go"},
		{"PackageVersion", "1.2"},
		{"PackageFileName", "spdx-tools-go.tar.gz"},
		{"PackageSupplier", "Organization: Linux Foundation"},
		{"PackageOriginator", "Organization: Linux Foundation"},
		{"PackageDownloadLocation", "http://git.spdx.org/spdx-tools-go.git"},
		{"PackageHomePage", "http://git.spdx.org/spdx-tools-go.git"},
		{"PackageSourceInfo", ":)"},
		{"PackageLicenseComments", "example comment"},
		{"PackageCopyrightText", "sample copyright text"},
		{"PackageSummary", "spdx library for go"},
		{"PackageDescription", "spdx parser library for go supporting tag and rdf formats"},
		{"PackageVerificationCode", "fdsa431fdsa43fdsa432"},
		{"PackageLicenseDeclared", "MIT1"},
		{"PackageLicenseConcluded", "MIT2"},
		// non-string comparable attributes:
		{"PackageChecksum", cksum.Algo.Val + ": " + cksum.Value.Val},
		{"PackageLicenseInfoFromFiles", "Apache2"},
		{"PackageLicenseInfoFromFiles", "MIT"},
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(doc.Packages) != 1 {
		t.Error("None or more packages than expected.")
	}
	pkg := doc.Packages[0]

	strExpected := []spdx.Value{
		doc.SpecVersion,
		pkg.Name,
		pkg.Version,
		pkg.FileName,
		pkg.Supplier,
		pkg.Originator,
		pkg.DownloadLocation,
		pkg.HomePage,
		pkg.SourceInfo,
		pkg.LicenceComments,
		pkg.CopyrightText,
		pkg.Summary,
		pkg.Description,
		pkg.VerificationCode.Value,
		pkg.LicenceDeclared,
		pkg.LicenceConcluded,
	}

	for i, val := range strExpected {
		p := input[i]
		if val.V() != p.Value {
			t.Errorf("Invalid %s: '%s'.", p.Key, val)
		}
	}

	if !pkg.Checksum.Equal(&cksum) {
		t.Errorf("Invalid PackageChecksum found: %+v \n Expected: %+v", *pkg.Checksum, cksum)
	}

	if len(pkg.LicenceInfoFromFiles) != 2 ||
		pkg.LicenceInfoFromFiles[0].LicenceId() != "Apache2" ||
		pkg.LicenceInfoFromFiles[1].LicenceId() != "MIT" {

		t.Errorf("Invalid PackageLicenseInfoFromFiles: '%s'", pkg.LicenceInfoFromFiles)
	}
}

func TestDocNestedFiles(t *testing.T) {
	cksum := spdx.Checksum{
		Algo:  spdx.Str("SHA1", nil),
		Value: spdx.Str("2fd4e1c67a2d28fced849ee1bb76e7391b93eb12", nil),
	}

	fContributor := []string{"Person: A", "Person: B", "Organization: LF"}
	fDependency := []string{"f0.txt", "f1.txt", "f2.txt"}
	licInfo := []string{"MIT", "Apache", "LicenseRef-0"}
	input := []Pair{
		{"SPDXVersion", "1.2"},
		{"PackageName", "spdx-tools-go"},

		{"FileName", "a.txt"},
		{"FileType", "BINARY"},
		{"LicenseConcluded", "MIT"},
		{"LicenseComments", "file licence comments"},
		{"FileCopyrightText", "some copyright text"},
		{"FileComment", "some file comment"},
		{"FileNotice", "example file notice"},

		// nested attributes
		{"ArtifactOfProjectName", "spdx-tools-go"},
		{"ArtifactOfProjectURI", "http://git.spdx.org/spdx-tools-go.git"},
		{"ArtifactOfProjectHomePage", "http://git.spdx.org/spdx-tools-go.git"},

		{"ArtifactOfProjectName", "spdx"},
		{"ArtifactOfProjectName", "spdx-tools"},

		// second file
		{"FileName", "b.go"},
		{"FileType", "SOURCE"},

		// non-string checks
		{"FileChecksum", cksum.Algo.Val + ": " + cksum.Value.Val},
	}

	for _, v := range fContributor {
		input = append(input, Pair{"FileContributor", v})
	}

	for _, v := range fDependency {
		input = append(input, Pair{"FileDependency", v})
	}

	for _, v := range licInfo {
		input = append(input, Pair{"LicenseInfoInFile", v})
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(doc.Packages) != 1 {
		t.Error("None or more packages than expected.")
	}
	pkg := doc.Packages[0]

	if len(doc.Files) != 2 {
		t.Errorf("Package has %s file(s). Expected %s", len(doc.Files), 2)
	}
	file1 := doc.Files[0]
	file2 := doc.Files[1]

	if len(file1.ArtifactOf) != 3 {
		t.Errorf("File 1 has %s ArtifactOf. Expected %s", len(file1.ArtifactOf), 3)
	}
	if len(file2.ArtifactOf) != 0 {
		t.Errorf("File 2 has %s ArtifactOf. Expected %s", len(file2.ArtifactOf), 0)
	}

	// check all string properties
	strExpected := []spdx.Value{
		doc.SpecVersion,
		pkg.Name,

		file1.Name,
		file1.Type,
		file1.LicenceConcluded,
		file1.LicenceComments,
		file1.CopyrightText,
		file1.Comment,
		file1.Notice,
		file1.ArtifactOf[0].Name,
		file1.ArtifactOf[0].ProjectUri,
		file1.ArtifactOf[0].HomePage,
		file1.ArtifactOf[1].Name,
		file1.ArtifactOf[2].Name,

		file2.Name,
		file2.Type,
	}

	for i, val := range strExpected {
		p := input[i]
		if val.V() != p.Value {
			t.Errorf("Invalid %s: '%s'.", p.Key, val)
		}
	}

	// same LicenceInfoInFile
	if len(file2.LicenceInfoInFile) != len(licInfo) {
		t.Errorf("Wrong file2.LicenceInfoInFile: '%+v'", file2.LicenceInfoInFile)
	} else {
		fail := false
		for i, lic := range file2.LicenceInfoInFile {
			if lic.LicenceId() != licInfo[i] {
				fail = true
				break
			}
		}
		if fail {
			t.Errorf("Wrong file2.LicenceInfoInFile: '%+v'", file2.LicenceInfoInFile)
		}
	}

	// sameContributors
	if !sameValStrValues(file2.Contributor, fContributor) {
		t.Errorf("Wrong file2.Contributor: '%+v'. Expected: '%+v'", file2.Contributor, fContributor)
	}

	if !file2.Checksum.Equal(&cksum) {
		t.Errorf("Invalid PackageChecksum: '%+v'.", *pkg.Checksum)
	}

}

func TestDocLicenceId(t *testing.T) {
	ids := [...]string{"LicenseRef-1", "LicenseRef-2"}
	xRef := [...][]string{{"A", "B", "Ca"}, {}}
	names := [...][]string{{"MIT", "Apache", "LicenseRef-0"}, {"Lic", "Example"}}

	input := make([]Pair, 0)

	for i, id := range ids {
		input = append(input, Pair{"LicenseID", id})
		for _, ref := range xRef[i] {
			input = append(input, Pair{"LicenseCrossReference", ref})
		}
		for _, name := range names[i] {
			input = append(input, Pair{"LicenseName", name})
		}
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: '%s'", err)
		t.FailNow()
	}

	if len(doc.ExtractedLicences) != len(ids) {
		t.Errorf("Wrong ExtractedLicences (len=%s): '%s'", len(doc.ExtractedLicences), doc.ExtractedLicences)
		t.FailNow()
	}

	for i, extr := range doc.ExtractedLicences {
		if extr.Id.Val != ids[i] {
			t.Errorf("(#%d) Wrong ID: '%s' (expected '%s')", i, extr.Id, ids[i])
		}
		if !sameValStrValues(extr.Name, names[i]) {
			t.Errorf("(#%d) Wrong Name: '%s' (expected '%s')", i, extr.Name, names[i])
		}
		if !sameValStrValues(extr.CrossReference, xRef[i]) {
			t.Errorf("(#%d) Wrong CrossReference: '%s' (expected '%s')", i, extr.CrossReference, xRef[i])
		}
	}
}

func TestDocInvalidProperty(t *testing.T) {
	input := []Pair{{"SomeInvalidProperty", "value"}}
	_, err := Parse(l(input))
	if err == nil {
		t.Fail()
	}
}

func TestElementsMeta(t *testing.T) {
	elements := []Pair{
		{"SPDXVersion", "1.2"},
		{"Creator", "Person: tester"},
		{"Reviewer", "Person: tester"},
		{"PackageName", "test.pkg.1"},
		{"PackageName", "test.pkg.2"},
		{"PackageLicenseConcluded", "GPLv3 or GPLv2"},
		{"PackageChecksum", "MD5: 432"},
		{"PackageVerificationCode", "fdsa"},
		{"FileName", "test.file.1"},
		{"FileName", "test.file.2"},
		{"LicenseInfoInFile", "GPLv2 and Apache2"},
		{"FileChecksum", "SHA1: 432"},
		{"ArtifactOfProjectName", "test.proj"},
		{"LicenseID", "LicenseRef-1"},
	}

	metas := make([]*spdx.Meta, len(elements))
	for i := range metas {
		metas[i] = spdx.NewMetaL(i + 1)
	}

	input := make([]*Token, len(elements))
	for i, pair := range elements {
		input[i] = &Token{Pair: pair, Meta: metas[i]}
	}

	doc, err := Parse(lt(input))
	if err != nil {
		t.Errorf("Unexpected parse error %+v", err)
		t.FailNow()
	}

	found := []*spdx.Meta{
		doc.Meta,
		doc.CreationInfo.Meta,
		doc.Reviews[0].Meta,
		doc.Packages[0].Meta,
		doc.Packages[1].Meta,
		doc.Packages[1].LicenceConcluded.M(),
		doc.Packages[1].Checksum.Meta,
		doc.Packages[1].VerificationCode.Meta,
		doc.Files[0].Meta,
		doc.Files[1].Meta,
		doc.Files[1].LicenceInfoInFile[0].M(),
		doc.Files[1].Checksum.Meta,
		doc.Files[1].ArtifactOf[0].Meta,
		doc.ExtractedLicences[0].Meta,
	}

	for i, m := range metas {
		if m != found[i] {
			t.Errorf("Wrong meta for %d:%s. Expected %+v but found %+v", i, elements[i], m, found[i])
		}
	}
}

func TestLicenceReferences(t *testing.T) {
	input := []Pair{
		{"PackageName", "test"},
		{"PackageLicenseConcluded", "LicenseRef-1"},
		{"PackageLicenseDeclared", "LicenseRef-2"},
		{"PackageLicenseInfoFromFiles", "LicenseRef-1"},
		{"PackageLicenseInfoFromFiles", "LicenseRef-2"},
		{"PackageLicenseInfoFromFiles", "(LicenseRef-1 and LicenseRef-2)"},
		{"PackageLicenseInfoFromFiles", "(LicenseRef-2 or LicenseRef-1)"},
		{"LicenseID", "LicenseRef-1"},
	}
	doc, err := Parse(l(input))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	pkg := doc.Packages[0]

	ref, ok := pkg.LicenceConcluded.(*spdx.ExtractedLicence)
	extracted := doc.ExtractedLicences[0]
	if !ok || ref != extracted {
		t.Errorf("Pointers do not match. extr %T %v and ref %T %v", extracted, extracted.LicenceId(), ref, ref)
	}

	if _, ok := pkg.LicenceDeclared.(spdx.Licence); !ok {
		t.Errorf("Wrong reference. %T %s", pkg.LicenceDeclared, pkg.LicenceDeclared.LicenceId())
	}

	lff := pkg.LicenceInfoFromFiles[0]
	ref, ok = lff.(*spdx.ExtractedLicence)
	if !ok || ref != doc.ExtractedLicences[0] {
		t.Errorf("Pointers do not match. extr %T %v and ref %T %v", extracted, extracted.LicenceId(), lff, lff.LicenceId())
	}

	if _, ok := pkg.LicenceInfoFromFiles[1].(spdx.Licence); !ok {
		t.Errorf("Wrong reference. %T %s", pkg.LicenceInfoFromFiles[1], pkg.LicenceInfoFromFiles[1].LicenceId())
	}

	conj := pkg.LicenceInfoFromFiles[2].(spdx.ConjunctiveLicenceSet)
	if _, ok := conj.Members[0].(*spdx.ExtractedLicence); !ok {
		t.Errorf("Conjunctive set not ok")
	}

	disj := pkg.LicenceInfoFromFiles[3].(spdx.DisjunctiveLicenceSet)
	if _, ok := disj.Members[1].(*spdx.ExtractedLicence); !ok {
		t.Errorf("Disjunctive set not ok")
	}
}

func TestLicenceAlreadyDefined(t *testing.T) {
	input := []Pair{
		{"PackageName", "some package"},
		{"PackageLicenseDeclared", "decl1"},
		{"PackageLicenseDeclared", "decl2"},
	}
	_, err := Parse(l(input))
	if err.Error() != MsgAlreadyDefined {
		t.Errorf("Unexpected error '%s'.", err)
	}
}

func TestDocLicenceConjAndDisj(t *testing.T) {
	input := []Pair{
		{"PackageName", "some package"},
		{"PackageLicenseDeclared", "(decl1 and decl2 or decl3)"},
	}
	_, err := Parse(l(input))
	if err.Error() != MsgConjunctionAndDisjunction {
		t.Errorf("Unexpected error '%s'.", err)
	}
}

func TestDocLicenceInvalidSet(t *testing.T) {
	input := []Pair{
		{"PackageName", "some package"},
		{"PackageLicenseDeclared", "a and ()"},
	}
	_, err := Parse(l(input))
	if err.Error() != MsgEmptyLicence {
		t.Errorf("(conjunctive) Unexpected error '%s'.", err)
	}

	input = []Pair{
		{"PackageName", "some package"},
		{"PackageLicenseDeclared", "a or ()"},
	}
	_, err = Parse(l(input))
	if err.Error() != MsgEmptyLicence {
		t.Errorf("(disjunctive) Unexpected error '%s'.", err)
	}

}

func TestChecksumAlreadyDefined(t *testing.T) {
	input := []Pair{
		{"PackageName", "some package"},
		{"PackageChecksum", "SHA1: d6a770ba38583ed4bb4525bd96e50471655d2758"},
		{"PackageChecksum", "SHA1: d6a770ba38583ed4bb4525bd96e50471655d2758"},
	}
	_, err := Parse(l(input))
	if err.Error() != MsgAlreadyDefined {
		t.Errorf("Unexpected error '%s'.", err)
	}
}

func TestEmptyDocumentElements(t *testing.T) {
	input := []Pair{
		{"SPDXVersion", "1.2"},
	}
	doc, err := Parse(l(input))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	if doc.CreationInfo != nil {
		t.Errorf("Document creation info not nil: %#v", doc.CreationInfo)
	}

	if doc.Packages != nil {
		t.Errorf("Packages not nil: %#v", doc.Packages)
	}

	if doc.Files != nil {
		t.Errorf("Files not nil: %#v", doc.Files)
	}

	if doc.Reviews != nil {
		t.Errorf("Reviews are not nil: %#v", doc.Reviews)
	}

	if doc.ExtractedLicences != nil {
		t.Errorf("ExtractedLicences are not nil: %#v", doc.ExtractedLicences)
	}

}

func TestEmptyPackageElements(t *testing.T) {
	input := []Pair{
		{"PackageName", "test package"},
	}

	doc, err := Parse(l(input))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	pkg := doc.Packages[0]

	if pkg.Checksum != nil {
		t.Errorf("Package checksum not nil: %#v", pkg.Checksum)
	}

	if pkg.VerificationCode != nil {
		t.Errorf("Package VerificationCode not nil: %#v", pkg.VerificationCode)
	}

	if pkg.Files != nil {
		t.Errorf("Package Files not nil: %#v", pkg.Files)
	}

	if pkg.LicenceInfoFromFiles != nil {
		t.Errorf("licence info from files not nil: %#v", pkg.LicenceInfoFromFiles)
	}
}

func TestEmptyFileElements(t *testing.T) {
	input := []Pair{
		{"FileName", "test file"},
	}

	doc, err := Parse(l(input))
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	f := doc.Files[0]

	if f.Checksum != nil {
		t.Errorf("File checksum not nil: %#v", f.Checksum)
	}

	if f.ArtifactOf != nil {
		t.Errorf("ArtifactOf not nil: %#v", f.ArtifactOf)
	}

	if f.Dependency != nil {
		t.Errorf("Dependency not nil: %#v", f.Dependency)
	}

	if f.Contributor != nil {
		t.Errorf("File contributor not nil: %#v", f.Contributor)
	}

	if f.LicenceInfoInFile != nil {
		t.Errorf("LiceceInfoInFile not nil %#v", f.LicenceInfoInFile)
	}
}

func TestVerifCodeAlreadyDefined(t *testing.T) {
	input := []Pair{
		{"PackageName", "some package"},
		{"PackageVerificationCode", "6a770ba38583ed4bb4525bd96e50471655d2758"},
		{"PackageVerificationCode", "d6a770ba38583ed4bb4525bd96e50471655d2758"},
	}
	_, err := Parse(l(input))
	if err.Error() != MsgAlreadyDefined {
		t.Errorf("Unexpected error '%s'.", err)
	}
}

func TestReviewer(t *testing.T) {

	reviews := []spdx.Review{
		{spdx.NewValueCreator("a", nil), spdx.NewValueDate("b", nil), spdx.Str("c", nil), nil},
		{spdx.NewValueCreator("d", nil), spdx.NewValueDate("e", nil), spdx.Str("f", nil), nil},
	}

	input := make([]Pair, 0, 6)
	for _, rev := range reviews {
		input = append(input, Pair{"Reviewer", rev.Reviewer.V()})
		input = append(input, Pair{"ReviewDate", rev.Date.V()})
		input = append(input, Pair{"ReviewComment", rev.Comment.V()})
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		t.FailNow()
	}

	if len(doc.Reviews) != len(reviews) {
		t.Errorf("Invalid reviews (len=%d): '%v'\n (expected (len=%d): '%v')", len(doc.Reviews), doc.Reviews, len(reviews), reviews)
		t.FailNow()
	}

	for i, rev := range doc.Reviews {
		if !rev.Equal(&reviews[i]) {
			t.Errorf("Invalid reviews (len=%d): '%+v'\n (expected (len=%d): '%+v')", len(doc.Reviews), doc.Reviews, len(reviews), reviews)
			t.FailNow()
		}
	}
}

func TestFileDependency(t *testing.T) {
	input := []Pair{
		{"FileName", "file1.txt"},
		{"FileDependency", "file2.txt"},
		{"FileDependency", "file3.txt"},
		{"FileName", "file2.txt"},
		{"FileName", "file3.txt"},
	}

	doc, err := Parse(l(input))

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		t.FailNow()
	}

	if len(doc.Files) != 3 {
		t.Errorf("Invalid no. of files. Found %d: %+v", len(doc.Files), doc.Files)
		t.FailNow()
	}

	file1 := doc.Files[0]

	if len(file1.Dependency) != 2 {
		t.Logf("Wrong file dependency for file1. Found %d files: %s", len(file1.Dependency), file1.Dependency)
		t.FailNow()
	}

	if file1.Dependency[0] != doc.Files[1] || file1.Dependency[1] != doc.Files[2] {
		t.Errorf("Wrong dependencies. 1) Expected: %+v but found %+v\n2)Expected: %+v but found %+v\n", doc.Files[1], file1.Dependency[0], doc.Files[2], file1.Dependency[1])
	}
}
