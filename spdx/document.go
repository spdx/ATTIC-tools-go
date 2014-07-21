package spdx

const (
	DATA_LICENCE_TAG = "CC0-1.0"
	DATA_LICENCE_RDF = "http://spdx.org/licenses/CC0-1.0"
)

// Represents a SPDX Document.
type Document struct {
	SpecVersion       ValueStr            // SPDX Version
	DataLicence       ValueStr            // Should have value DATA_LICENCE_TAG
	CreationInfo      *CreationInfo       // Pointer to Creation Info element
	ExtractedLicences []*ExtractedLicence // Extracted Licences found in this doc
	Packages          []*Package          // Nested Packages
	Files             []*File             // Files referenced in this doc
	Comment           ValueStr            // Document comment
	Reviews           []*Review           // Document reviews
	*Meta                                 // Document metadata
}

// Return the document metadata.
func (doc *Document) M() *Meta { return doc.Meta }

// Represents the Creation Info part of a document
type CreationInfo struct {
	Creator            []ValueCreator // Creator of the document
	Created            ValueDate      // Creation date
	LicenceListVersion ValueStr       // Version of the SPDX licence list used
	Comment            ValueStr       // Creator comment
	*Meta                             // Creation Info meta
}

// Returns the creation info metadata.
func (ci *CreationInfo) M() *Meta { return ci.Meta }
