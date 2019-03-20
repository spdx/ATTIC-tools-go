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

// M returns the document metadata.
func (doc *Document) M() *Meta { return doc.Meta }

// Equal checks if this document is equal to `other`. Ignores metadata. Slices
// elements (ExtractedLicences, Packages, Files and Reviews) must appear
// in the same order for this method to return true.
func (doc *Document) Equal(other *Document) bool {
	if doc == other {
		return true
	}
	if doc == nil || other == nil {
		return false
	}
	eq := doc.SpecVersion.Val == other.SpecVersion.Val &&
		doc.DataLicence.Val == other.DataLicence.Val &&
		doc.CreationInfo.Equal(other.CreationInfo) &&
		len(doc.ExtractedLicences) == len(other.ExtractedLicences) &&
		len(doc.Packages) == len(other.Packages) &&
		len(doc.Files) == len(other.Files) &&
		len(doc.Reviews) == len(other.Reviews) &&
		doc.Comment.Val == other.Comment.Val

	if !eq {
		return false
	}

	for i, lic := range doc.ExtractedLicences {
		if !lic.Equal(other.ExtractedLicences[i]) {
			return false
		}
	}
	for i, pkg := range doc.Packages {
		if !pkg.Equal(other.Packages[i]) {
			return false
		}
	}
	for i, file := range doc.Files {
		if !file.Equal(other.Files[i]) {
			return false
		}
	}
	for i, rev := range doc.Reviews {
		if !rev.Equal(other.Reviews[i]) {
			return false
		}
	}

	return true
}

// Represents the Creation Info part of a document
type CreationInfo struct {
	Creator            []ValueCreator // Creator of the document
	Created            ValueDate      // Creation date
	LicenceListVersion ValueStr       // Version of the SPDX licence list used
	Comment            ValueStr       // Creator comment
	*Meta                             // Creation Info meta
}

// M returns the creation info metadata.
func (ci *CreationInfo) M() *Meta { return ci.Meta }

// Equal checks if this CreationInfo is equal to `other`. Ignores metadata.
func (ci *CreationInfo) Equal(other *CreationInfo) bool {
	if ci == other {
		return true
	}
	if ci == nil || other == nil {
		return false
	}
	if len(ci.Creator) != len(other.Creator) {
		return false
	}
	for i, cr := range ci.Creator {
		if cr.V() != other.Creator[i].V() {
			return false
		}
	}
	return ci.Created.V() == other.Created.V() &&
		ci.LicenceListVersion.Val == other.LicenceListVersion.Val &&
		ci.Comment.Val == other.Comment.Val
}
