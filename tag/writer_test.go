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

func TestToken(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter(buf)

	tests := map[string]*Token{
		"# comment\n":         CommentTok("comment"),
		"PackageName: test\n": PairTok("PackageName", "test"),
	}

	for expected, tok := range tests {
		f.lastWritten = ""
		f.Token(tok)
		if result := buf.String(); result != expected {
			t.Errorf("%#v: expected %#v but found %#v", *tok, expected, result)
		}
		buf.Reset()
	}

	if err := f.Token(&Token{Type: tokenKey}); err == nil {
		t.Error("No error for wrong token type.")
	}
}

func TestComment(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter(buf)

	tests := []Pair{
		{"", "# \n"},
		{"#", "## \n"},
		{"text", "# text\n"},
		{" text", "# text\n"},
		{"#text", "## text\n"},
		{"# text", "## text\n"},
		{"##text", "### text\n"},
	}

	for _, p := range tests {
		f.Comment(p.Key)
		if res := buf.String(); res != p.Value {
			t.Errorf("Incorrect comment for %#v. Printed %#v but expected %#v", p.Key, res, p.Value)
		}
		buf.Reset()
	}

	if f.lastWritten != commentLastWritten {
		t.Errorf("f.lastWritten is %s instead of %s.", f.lastWritten, commentLastWritten)
	}
}

func TestProperty(t *testing.T) {
	buf := new(bytes.Buffer)
	f := NewFormatter(buf)

	tests := map[string]Pair{
		"": {},
		"FileName: testfile\n":              {"FileName", "testfile"},
		"LicenseConcluded: NOASSERTION\n":   {"LicenseConcluded", "NOASSERTION"},
		"LicenseConcluded: NONE\n":          {"LicenseConcluded", "NONE"},
		"DocumentComment: NONE\n":           {"DocumentComment", "NONE"},
		"DocumentComment: NOASSERTION\n":    {"DocumentComment", "NOASSERTION"},
		"DocumentComment: <text>a</text>\n": {"DocumentComment", "a"},
		"PackageName: <text>a\nb</text>\n":  {"PackageName", "a\nb"},
	}

	for expected, p := range tests {
		f.lastWritten = ""
		f.Property(p.Key, p.Value)
		if res := buf.String(); res != expected {
			t.Errorf("Incorrect result for %#v. Printed %#v but expected %#v", p, res, expected)
		}
		buf.Reset()
	}
}
