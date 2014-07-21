package spdx

// Represents a SPDX File.
type File struct {
	Name              ValueStr      // File name.
	Type              ValueStr      // File type.
	Checksum          *Checksum     // File Checksum.
	LicenceConcluded  AnyLicence    // Licence Concluded. NOASSERTION and NONE values allowed
	LicenceInfoInFile []AnyLicence  // Licence Info in File. NOASSERTION and NONE values allowed
	LicenceComments   ValueStr      // Licence comments.
	CopyrightText     ValueStr      // File copyright text NOASSERTION and NONE allowed.
	Notice            ValueStr      // File notice.
	ArtifactOf        []*ArtifactOf // A list of artifacts
	Dependency        []*File       // File dependecies.
	Contributor       []ValueStr    // File contributors.
	Comment           ValueStr      // File comments.
	*Meta                           // File metadata.
}

// Returns the File metadata.
func (f *File) M() *Meta { return f.Meta }

// Represents the ArtifactOf* properties of a SPDX File.
type ArtifactOf struct {
	ProjectUri ValueStr // Project URI
	HomePage   ValueStr // Project HomePage
	Name       ValueStr // Project Name
	*Meta               // Artifact metadata.
}

// Returns the artifact metadata.
func (artif *ArtifactOf) M() *Meta { return artif.Meta }
