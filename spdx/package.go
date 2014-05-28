package spdx

type Package struct {
	Name                 string            // one
	VersionInfo          string            // zero or one
	DownloadLocation     string            // one; NOASSERTION or NONE valid
	HomePage             string            // one; NOASSERTION or NONE valid
	FileName             string            // zero or one
	Supplier             string            // zero or one, NOASSERTION valid
	Originator           string            // zero or one, NOASSERTION valid
	VerificationCode     *VerificationCode // mandatory, one
	Checksum             *Checksum         // zero or one
	SourceInfo           string            // zero or one
	LicenceConcluded     AnyLicenceInfo    // one; NOASSERTION or NONE valid
	LicenceInfoFromFiles []AnyLicenceInfo  // one or more; NOASSERTION or NONE valid; no sets allowed
	LicenceDeclared      AnyLicenceInfo    // one
	LicenceComments      string            // zero or one
	CopyrightText        string            // one
	Summary              string            // zero or one
	Description          string            // zero or one
	HasFile              []*File           // one or more
}

type VerificationCode struct {
	Value         string
	ExcludedFiles []string
}

type Checksum struct {
	Algo, Value string
}
