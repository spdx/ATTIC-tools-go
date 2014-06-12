package spdx

type File struct {
	Name              ValueStr         // mandatory
	Type              ValueStr         // optional
	Checksum          *Checksum        // mandatory
	LicenceConcluded  AnyLicenceInfo   // mandatory, NOASSERTION and NONE allowed
	LicenceInfoInFile []AnyLicenceInfo // no sets; NOASSERTION and NONE allowed
	LicenceComments   ValueStr         // optional
	CopyrightText     ValueStr         // mandatory
	Notice            ValueStr         // optional
	ArtifactOf        []*ArtifactOf    // optinal
	Dependency        []*File          // optional
	Contributor       []ValueStr       // optional
	Comment           ValueStr         //optional
}

type ArtifactOf struct {
	ProjectUri ValueStr
	HomePage   ValueStr
	Name       ValueStr
}
