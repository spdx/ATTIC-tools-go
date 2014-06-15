package spdx

type Package struct {
	Name                 ValueStr          // one
	Version              ValueStr          // zero or one
	DownloadLocation     ValueStr          // one; NOASSERTION or NONE valid
	HomePage             ValueStr          // one; NOASSERTION or NONE valid
	FileName             ValueStr          // zero or one
	Supplier             ValueCreator      // zero or one, NOASSERTION valid
	Originator           ValueCreator      // zero or one, NOASSERTION valid
	VerificationCode     *VerificationCode // mandatory, one
	Checksum             *Checksum         // zero or one
	SourceInfo           ValueStr          // zero or one
	LicenceConcluded     AnyLicenceInfo    // one; NOASSERTION or NONE valid
	LicenceInfoFromFiles []AnyLicenceInfo  // one or more; NOASSERTION or NONE valid; no sets allowed
	LicenceDeclared      AnyLicenceInfo    // one
	LicenceComments      ValueStr          // zero or one
	CopyrightText        ValueStr          // one
	Summary              ValueStr          // zero or one
	Description          ValueStr          // zero or one
	Files                []*File           // one or more
}

type VerificationCode struct {
	Value         ValueStr
	ExcludedFiles []ValueStr
}

type Checksum struct {
	Algo, Value ValueStr
}

func (c *Checksum) Equal(d *Checksum) bool {
	return c.Algo.Val == d.Algo.Val && c.Value.Val == d.Value.Val
}
