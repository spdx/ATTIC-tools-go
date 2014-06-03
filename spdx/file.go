package spdx

type File struct {
	Name              string           // mandatory
	Type              string           // optional
	Checksum          *Checksum        // mandatory
	LicenceConcluded  AnyLicenceInfo   // mandatory, NOASSERTION and NONE allowed
	LicenceInfoInFile []AnyLicenceInfo // no sets; NOASSERTION and NONE allowed
	LicenceComments   string           // optional
	CopyrightText     string           // mandatory
	Notice            string           // optional
	ArtifactOf        []*ArtifactOf    // optinal
	Dependency        []*File          // optional
	Contributor       []string         // optional
}

type ArtifactOf struct {
	ProjectUri string
	HomePage   string
	Name       string
}
