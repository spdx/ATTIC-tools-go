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

func TestUpd(t *testing.T) {
	a := "hello"
	f := upd(&a)
	f("world")
	if a != "world" {
		t.Fail()
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
