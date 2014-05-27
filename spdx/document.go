package spdx

const DATA_LICENCE_TAG = "CC0-1.0"
const DATA_LICENCE_RDF = "http://spdx.org/licenses/CC0-1.0"

type Document struct {
	SpecVersion  string        // one
	DataLicence  string        // one
	CreationInfo *CreationInfo //
	Packages     []*Package    // spdx1.2: one, spdx2: one or more
	Comment      *string       // zero or one
}

type CreationInfo struct {
	Creator            []string // one or many
	Created            string   // one
	LicenceListVersion *string  // zero or one
	Comment            *string  // zero or one
}
