package tag

import (
	"bytes"
	"github.com/vladvelici/spdx-go/spdx"
	"testing"
)

func TestCksumStr(t *testing.T) {
	res := cksumStr(nil)
	if res != "" {
		t.Errorf("Incorrect value for nil Checksum: %s", res)
	}

	cksum := new(spdx.Checksum)
	res = cksumStr(cksum)
	if res != "" {
		t.Errorf("Incorrect value for empty Checksum: %s", res)
	}

	cksum.Algo = spdx.Str("SHA1", nil)
	cksum.Value = spdx.Str("testvalue", nil)
	res = cksumStr(cksum)
	if res != cksum.Algo.Val+": "+cksum.Value.Val {
		t.Errorf("Incorrect value for \"SHA1: testvalue\" Checksum: %s", res)
	}
}

func TestVerifCodeStr(t *testing.T) {
	res := verifCodeStr(nil)
	if res != "" {
		t.Errorf("Incorrect value for nil VerificationCode: %s", res)
	}

	verif := new(spdx.VerificationCode)
	res = verifCodeStr(verif)
	if res != "" {
		t.Errorf("Incorrect value for empty VerificationCode: %s", res)
	}

	verif.Value = spdx.Str("testcode", nil)
	res = verifCodeStr(verif)
	if res != verif.Value.Val {
		t.Errorf("Incorrect value for \"testcode\" VerificationCode: %s", res)
	}

	verif.ExcludedFiles = []spdx.ValueStr{spdx.Str("f1", nil), spdx.Str("f2", nil)}
	res = verifCodeStr(verif)
	if res != verif.Value.Val+" (Excludes: f1, f2)" {
		t.Errorf("Incorrect value for \"testcode (Excludes: f1, f2)\" VerificationCode: %s", res)
	}
}

func TestFormatterSpaces(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter(buf)

	f.lastWritten = ""
	f.spaces("FileName")
	if buf.String() != "" {
		t.Error("Newline at the beginning of document.")
	}

	f.lastWritten = commentLastWritten
	f.spaces(commentLastWritten)
	if buf.String() != "" {
		t.Error("Newline between comments.")
	}

	f.spaces("FileName")
	if buf.String() != "" {
		t.Error("Newline between comment and 'special' property.")
	}

	f.lastWritten = "PackageLicenseConcluded"
	properties := []string{"PackageName", "FileName", "LicenseID", "Reviewer", "ArtifactOfProjectName", commentLastWritten}
	for _, property := range properties {
		f.spaces(property)
		if buf.String() != "\n" {
			t.Errorf("No newline before %s", property)
		}
		buf.Reset()
	}
}
