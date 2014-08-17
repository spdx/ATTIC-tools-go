package rdf

import (
	"os"
	"testing"
)

const testFile = "testfile.rdf"

// Parse a document, write it and parse it again. The parsed documents should be the same.
func TestParseWriteParse(t *testing.T) {
	documentReader, err := os.Open(testFile)
	if err != nil {
		t.Logf("The RDF package should contain a test file called %s.", testFile)
		t.FailNow()
	}

	parsed1, err := Parse(documentReader, "rdf")
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Log("Couldn't create a pipe.")
		t.FailNow()
	}

	err = Write(w, parsed1)
	if err != nil {
		t.Logf("Write error (first write): %s", err)
		t.FailNow()
	}
	w.Close()

	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}

	parsed2, err := Parse(r, "rdf")
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}
	if !parsed1.Equal(parsed2) {
		t.Error("Documents are not the same.")
		t.FailNow()
	}

	documentReader.Close()
	r.Close()
}

// There are known problems writing other RDF formats than xmlrdf-abbrev.
// Refer to issue #46 on github: https://github.com/vladvelici/spdx-go/issues/46
func _TestWriteFormat(t *testing.T) {
	format := Fmt_turtle

	documentReader, err := os.Open(testFile)
	if err != nil {
		t.Logf("The RDF package should contain a test file called %s.", testFile)
		t.FailNow()
	}

	parsed1, err := Parse(documentReader, "rdf")
	if err != nil {
		t.FailNow()
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Log("Couldn't create a pipe.")
		t.FailNow()
	}

	err = WriteFormat(w, parsed1, format)
	if err != nil {
		t.Logf("Write error (first write): %s", err)
		t.FailNow()
	}
	w.Close()

	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}

	parsed2, err := Parse(r, format)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}

	if !parsed1.Equal(parsed2) {
		t.Error("Documents are not the same.")
		t.FailNow()
	}

	documentReader.Close()
	r.Close()
}
