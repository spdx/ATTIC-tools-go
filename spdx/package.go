package spdx

type Package struct {
	Name                 ValueStr          // one
	Version              ValueStr          // zero or one
	DownloadLocation     ValueStr          // one; NOASSERTION or NONE valid
	HomePage             ValueStr          // zero or one; NOASSERTION or NONE valid
	FileName             ValueStr          // zero or one
	Supplier             ValueCreator      // zero or one, NOASSERTION valid
	Originator           ValueCreator      // zero or one, NOASSERTION valid
	VerificationCode     *VerificationCode // mandatory, one
	Checksum             *Checksum         // zero or one
	SourceInfo           ValueStr          // zero or one
	LicenceConcluded     AnyLicence        // one; NOASSERTION or NONE valid
	LicenceInfoFromFiles []AnyLicence      // one or more; NOASSERTION or NONE valid; no sets allowed
	LicenceDeclared      AnyLicence        // one
	LicenceComments      ValueStr          // zero or one
	CopyrightText        ValueStr          // one
	Summary              ValueStr          // zero or one
	Description          ValueStr          // zero or one
	Files                []*File           // one or more
	*Meta
}

func (pkg *Package) M() *Meta { return pkg.Meta }

type VerificationCode struct {
	Value         ValueStr
	ExcludedFiles []ValueStr
	*Meta
}

func (vc *VerificationCode) M() *Meta { return vc.Meta }

type Checksum struct {
	Algo, Value ValueStr
	*Meta
}

func (c *Checksum) Equal(d *Checksum) bool {
	return c.Algo.Val == d.Algo.Val && c.Value.Val == d.Value.Val
}

func (c *Checksum) M() *Meta { return c.Meta }
