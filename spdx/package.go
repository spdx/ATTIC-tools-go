package spdx

type Package struct {
	Name                 string
	VersionInfo          string
	DownloadLocation     string
	HomePage             string
	FileName             string
	Supplier             []string
	Originator           []string
	VerificationCode     *VerificationCode
	Checksum             *Checksum
	SourceInfo           string
	LicenceConcluded     *AnyLicenceInfo
	LicenceInfoFromFiles *SimpleLicenceInfo
	LicenceDeclared      *AnyLicenceInfo
	LicenceComments      string
	CopyrightText        string
	Summary              string
	Description          string
}

type VerificationCode struct {
	Value         string
	ExcludedFiles []string
}

type Checksum struct {
	Algo, Value string
}
