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

// M returns the File metadata.
func (f *File) M() *Meta { return f.Meta }

// Equal checks if this file is equal to `other`. Ignores metadata. Elements
// in slices file.Contributor, file.Dependency, file.ArtifactOf and
// file.LicenceInfoInFile must be in the same order for this method to
// return true.
func (f *File) Equal(other *File) bool {
	eq := (f == other) || (f != nil && other != nil &&
		f.Name.Val == other.Name.Val &&
		f.Type.Val == other.Type.Val &&
		f.LicenceComments.Val == other.LicenceComments.Val &&
		f.CopyrightText.Val == other.CopyrightText.Val &&
		f.Notice.Val == other.Notice.Val &&
		f.Comment.Val == other.Comment.Val &&
		f.Checksum.Equal(other.Checksum) &&
		SameLicence(f.LicenceConcluded, other.LicenceConcluded) &&
		len(f.LicenceInfoInFile) == len(other.LicenceInfoInFile) &&
		len(f.ArtifactOf) == len(other.ArtifactOf) &&
		len(f.Dependency) == len(other.Dependency) &&
		len(f.Contributor) == len(other.Contributor))
	if !eq {
		return false
	}
	for i, lic := range f.LicenceInfoInFile {
		if !SameLicence(lic, other.LicenceInfoInFile[i]) {
			return false
		}
	}
	for i, elem := range f.ArtifactOf {
		if !elem.Equal(other.ArtifactOf[i]) {
			return false
		}
	}
	for i, elem := range f.Dependency {
		if !elem.Equal(other.Dependency[i]) {
			return false
		}
	}
	for i, v := range f.Contributor {
		if v.Val == other.Contributor[i].Val {
			return false
		}
	}
	return true
}

// Represents the ArtifactOf* properties of a SPDX File.
type ArtifactOf struct {
	ProjectUri ValueStr // Project URI
	HomePage   ValueStr // Project HomePage
	Name       ValueStr // Project Name
	*Meta               // Artifact metadata.
}

// M returns the artifact metadata.
func (artif *ArtifactOf) M() *Meta { return artif.Meta }

// Equal checks if this ArtifactOf is equal to `o`. Ignores metadata.
func (a *ArtifactOf) Equal(o *ArtifactOf) bool {
	return a == o || (a != nil && o != nil &&
		a.ProjectUri.Val == o.ProjectUri.Val &&
		a.HomePage.Val == o.HomePage.Val &&
		a.Name.Val == o.Name.Val)
}
