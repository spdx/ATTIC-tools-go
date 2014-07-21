package spdx

// Represents a SPDX Package.
type Package struct {
	Name                 ValueStr          // Package name.
	Version              ValueStr          // Package version.
	DownloadLocation     ValueStr          // Package download location. NOASSERTION and NONE are allowed.
	HomePage             ValueStr          // Package homepage; NOASSERTION and NONE are allowed.
	FileName             ValueStr          // Package filename
	Supplier             ValueCreator      // Package supplier. NOASSERTION is allowed.
	Originator           ValueCreator      // Package originator. NOASSERTION is allowed.
	VerificationCode     *VerificationCode // Package verification code.
	Checksum             *Checksum         // Package Checksum.
	SourceInfo           ValueStr          // Package source info.
	LicenceConcluded     AnyLicence        // Package concluded lincence. NOASSERTION and NONE are allowed.
	LicenceInfoFromFiles []AnyLicence      // Licence info from files. NOASSERTION and NONE are allowed. No sets allowed.
	LicenceDeclared      AnyLicence        // Package licence declared.
	LicenceComments      ValueStr          // Licence comments.
	CopyrightText        ValueStr          // Package copyright text.
	Summary              ValueStr          // Package summary.
	Description          ValueStr          // Package description.
	Files                []*File           // Package files.
	*Meta                                  // Package metadata.
}

// Returns the package metadata.
func (pkg *Package) M() *Meta { return pkg.Meta }

// Represents a package verification code.
type VerificationCode struct {
	Value         ValueStr   // Verification code
	ExcludedFiles []ValueStr // List of excluded file names
	*Meta                    // Verification code metadata.
}

// Package metadata.
func (vc *VerificationCode) M() *Meta { return vc.Meta }

// Represents a file or package checksum.
type Checksum struct {
	Algo, Value ValueStr
	*Meta
}

// Compares two checksums, ignoring their metadatas.
func (c *Checksum) Equal(d *Checksum) bool {
	return c.Algo.Val == d.Algo.Val && c.Value.Val == d.Value.Val
}

// Returns the checksum metadata.
func (c *Checksum) M() *Meta { return c.Meta }
