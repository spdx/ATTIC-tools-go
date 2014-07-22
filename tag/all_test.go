package tag

import (
	"bytes"
	"strings"
	"testing"
)

// Extract from the Java Library SPDX Document
func getDocumentString() string {
	return `SPDXVersion: SPDX-1.2
DataLicense: CC0-1.0
DocumentComment: <text>This is a sample spreadsheet</text>

Creator: Person: Gary O'Neall
Creator: Organization: Source Auditor Inc.
Creator: Tool: SourceAuditor-V1.2
Created: 2010-02-03T00:00:00Z
CreatorComment: <text>This is an example of an SPDX spreadsheet format</text>

Reviewer: Person: Joe Reviewer
ReviewDate: 2010-02-10T00:00:00Z
ReviewComment: <text>This is just an example.  Some of the non-standard licenses look like they are actually BSD 3 clause licenses</text>

Reviewer: Person: Suzanne Reviewer
ReviewDate: 2011-03-13T00:00:00Z
ReviewComment: <text>Another example reviewer.</text>

PackageName: SPDX Translator
PackageVersion: Version 0.9.2
PackageDownloadLocation: http://www.spdx.org/tools
PackageSummary: <text>SPDX Translator utility</text>
PackageSourceInfo: Version 1.0 of the SPDX Translator application
PackageFileName: spdxtranslator-1.0.zip
PackageSupplier: Organization:Linux Foundation
PackageOriginator: Organization:SPDX
PackageChecksum: SHA1: 2fd4e1c67a2d28fced849ee1bb76e7391b93eb12
PackageVerificationCode: 4e3211c67a2d28fced849ee1bb76e7391b93feba (SpdxTranslatorSpdx.rdf, SpdxTranslatorSpdx.txt)
PackageDescription: <text>This utility translates and SPDX RDF XML document to a spreadsheet, translates a spreadsheet to an SPDX RDF XML document and translates an SPDX RDFa document to an SPDX RDF XML document.</text>

PackageCopyrightText: <text> Copyright 2010, 2011 Source Auditor Inc.</text>

PackageLicenseDeclared: (LicenseRef-3 AND LicenseRef-4 AND Apache-2.0 AND MPL-1.1 AND LicenseRef-1 AND LicenseRef-2)
PackageLicenseConcluded: (LicenseRef-3 AND LicenseRef-4 AND Apache-1.0 AND Apache-2.0 AND MPL-1.1 AND LicenseRef-1 AND LicenseRef-2)

PackageLicenseInfoFromFiles: Apache-1.0
PackageLicenseInfoFromFiles: LicenseRef-3
PackageLicenseInfoFromFiles: Apache-2.0
PackageLicenseInfoFromFiles: LicenseRef-4
PackageLicenseInfoFromFiles: LicenseRef-2
PackageLicenseInfoFromFiles: LicenseRef-1
PackageLicenseInfoFromFiles: MPL-1.1
PackageLicenseComments: <text>The declared license information can be found in the NOTICE file at the root of the archive file</text>

FileName: src/org/spdx/parser/DOAPProject.java
FileType: SOURCE
FileChecksum: SHA1: 2fd4e1c67a2d28fced849ee1bb76e7391b93eb12
LicenseConcluded: Apache-2.0
LicenseInfoInFile: Apache-2.0
FileCopyrightText: <text>Copyright 2010, 2011 Source Auditor Inc.</text>`

}

// Parse a document, write it and parse it again. The parsed documents should be the same.
func TestParseWriteParse(t *testing.T) {
	document := getDocumentString()
	documentReader := strings.NewReader(document)
	parsed1, err := Build(documentReader)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}

	buffer := new(bytes.Buffer)
	err = Write(buffer, parsed1)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}
	parsed2, err := Build(buffer)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		t.FailNow()
	}

	if !parsed1.Equal(parsed2) {
		t.Error("Documents are not the same.")
	}
}
