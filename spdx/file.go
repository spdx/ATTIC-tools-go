package spdx

type File struct {
	Name              string
	Type              int
	Checksum          *Checksum
	LicenceConcluded  *AnyLicenceInfo
	LicenceInfoInFile *SimpleLicenceInfo
	LicenceComments   string
	CopyrightText     string
	NoticeText        string
	ArtifactOf        []*ArtifactOf
	Dependency        *File
	Contributor       string
}

type ArtifactOf struct {
	ProjectUri string
	HomePage   string
	Name       string
}
