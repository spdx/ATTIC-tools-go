package spdx

const DATA_LICENCE_TAG = "CC0-1.0"
const DATA_LICENCE_RDF = "http://spdx.org/licenses/CC0-1.0"

type Document struct {
	SpecVersion          ValueStr                  // one
	DataLicence          ValueStr                  // one
	CreationInfo         *CreationInfo             //
	ExtractedLicenceInfo []*ExtractedLicensingInfo // optional
	Packages             []*Package                // spdx1.2: one, spdx2: one or more
	Files                []*File                   // mandatory
	Comment              ValueStr                  // optional
	Reviews              []*Review                 // optional
	OtherLicences        map[string]*Licence       // licences that are not in the spec and are referenced to
}

type CreationInfo struct {
	Creator            []ValueStr // one or many
	Created            ValueStr   // one
	LicenceListVersion ValueStr   // zero or one
	Comment            ValueStr   // zero or one
}
