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

func sameLicence(a, b spdx.AnyLicenceInfo) bool {
	switch ta := a.(type) {
	default:
		return false
	case spdx.LicenceReference:
		if tb, ok := b.(spdx.LicenceReference); ok && tb == ta {
			return true
		}
		return false
	case spdx.DisjunctiveLicenceList:
		if tb, ok := b.(spdx.DisjunctiveLicenceList); ok && len(ta) == len(tb) {
			for i, lica := range ta {
				licb := tb[i]
				if !sameLicence(lica, licb) {
					return false
				}
			}
			return true
		}
		return false
	case spdx.ConjunctiveLicenceList:
		if tb, ok := b.(spdx.ConjunctiveLicenceList); ok && len(ta) == len(tb) {
			for i, lica := range ta {
				licb := tb[i]
				if !sameLicence(lica, licb) {
					return false
				}
			}
			return true
		}
		return false
	}
}

func joinStrSlice(a []string, glue string) string {
	if len(a) == 0 {
		return ""
	}
	result := a[0]
	for i := 1; i < len(a); i++ {
		result += glue + a[i]
	}
	return result
}

func TestUpd(t *testing.T) {
	a := ""
	f := upd(&a)
	f("world")
	if a != "world" {
		t.Fail()
	}
}

func TestUpdFail(t *testing.T) {
	a := ""
	f := upd(&a)
	f("hello")
	err := f("world")
	if err != ErrAlreadyDefined {
		t.Errorf("Different error: %s", err)
	}
}

func TestUpdList(t *testing.T) {
	arr := []string{"1", "2", "3"}
	f := updList(&arr)
	f("4")
	if len(arr) != 4 || arr[3] != "4" {
		t.Fail()
	}
}

func TestVerifCodeNoExcludes(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	f := verifCode(vc)
	err := f(value)
	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}
	if vc.Value != value {
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
	f := verifCode(vc)
	err := f(value + excludes)

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}
	if vc.Value != value {
		t.Error("Verification code value different than given value")
	}

	if !sameStrSlice(vc.ExcludedFiles, []string{"abc.txt", "file.spdx"}) {
		t.Error("Different ExcludedFiles. Elements found: %s.", vc.ExcludedFiles)
	}
}

func TestVerifCodeWithExcludes2(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	excludes := " (abc.txt, file.spdx)"
	f := verifCode(vc)

	err := f(value + excludes)

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}

	if vc.Value != value {
		t.Error("Verification code value different than given value")
	}

	if !sameStrSlice(vc.ExcludedFiles, []string{"abc.txt", "file.spdx"}) {
		t.Error("Different ExcludedFiles. Elements found: %s.", vc.ExcludedFiles)
	}
}

func TestVerifCodeInvalidExcludesNoClosedParentheses(t *testing.T) {
	vc := new(spdx.VerificationCode)
	value := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	excludes := " ("
	f := verifCode(vc)
	err := f(value + excludes)

	if err != ErrNoClosedParen {
		t.Errorf("Error should be UnclosedParentheses but found %s.", err)
	}
}

func TestChecksum(t *testing.T) {
	cksum := new(spdx.Checksum)
	val := "d6a770ba38583ed4bb4525bd96e50461655d2758"
	algo := "SHA1"

	f := checksum(cksum)
	err := f(algo + ": " + val)

	if err != nil {
		t.Errorf("Error should be nil but found %s.", err)
	}

	if cksum.Value != "d6a770ba38583ed4bb4525bd96e50461655d2758" {
		t.Errorf("Checksum value is wrong. Found: '%s'. Expected: '%s'.", cksum.Value, val)
	}

	if cksum.Algo != algo {
		t.Errorf("Algo is wrong. Found: '%s'. Expected: '%s'.", cksum.Algo, algo)
	}
}

func TestChecksumInvalid(t *testing.T) {
	cksum := new(spdx.Checksum)
	f := checksum(cksum)
	err := f("d6a770ba38583ed4bb")

	if err != ErrInvalidChecksum {
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
	expected := spdx.DisjunctiveLicenceList{spdx.NewLicenceReference("GPLv3"), spdx.NewLicenceReference("LicenseRef-1")}
	output, err := parseLicenceSet(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !sameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetAnd(t *testing.T) {
	input := "(GPLv3 and LicenseRef-1)"
	expected := spdx.ConjunctiveLicenceList{spdx.NewLicenceReference("GPLv3"), spdx.NewLicenceReference("LicenseRef-1")}
	output, err := parseLicenceSet(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !sameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetNested(t *testing.T) {
	input := "(GPLv3 or (LicenseRef-1 and LicenseRef-3) or LicenseRef-2)"

	expected := spdx.DisjunctiveLicenceList{
		spdx.NewLicenceReference("GPLv3"),
		spdx.ConjunctiveLicenceList{
			spdx.NewLicenceReference("LicenseRef-1"),
			spdx.NewLicenceReference("LicenseRef-3"),
		},
		spdx.NewLicenceReference("LicenseRef-2"),
	}

	output, err := parseLicenceSet(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !sameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetSingleValue(t *testing.T) {
	input := "one value "

	expected := spdx.NewLicenceReference("one value")

	output, err := parseLicenceSet(input)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !sameLicence(output, expected) {
		t.Errorf("\nExpected: %T : %s\nbut found %T : %s\n", expected, expected, output, output)
	}
}

func TestParseLicenceSetEmptyLicence(t *testing.T) {
	input := " "

	_, err := parseLicenceSet(input)
	if err != ErrEmptyLicence {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseLicenceStringEmptyLicence(t *testing.T) {
	input := " "

	_, err := parseLicenceString(input)
	if err != ErrEmptyLicence {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseLicenceStringUnbalancedParentheses(t *testing.T) {
	input := " (()"

	_, err := parseLicenceString(input)
	if err != ErrNoClosedParen {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestLicenceSetConjunctionAndDisjunction(t *testing.T) {
	input := "a and b or c"
	_, err := parseLicenceSet(input)
	if err != ErrConjunctionAndDisjunction {
		t.Errorf("Unexpected error: %s", err)
	}
}
