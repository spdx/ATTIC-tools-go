package spdx

type File struct {
	Name              ValueStr         // mandatory
	Type              ValueStr         // optional
	Checksum          *Checksum        // mandatory
	LicenceConcluded  AnyLicence   // mandatory, NOASSERTION and NONE allowed
	LicenceInfoInFile []AnyLicence // no sets; NOASSERTION and NONE allowed
	LicenceComments   ValueStr         // optional
	CopyrightText     ValueStr         // mandatory, NOASSERTION and NONE allowed
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
