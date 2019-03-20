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

// M returns the package metadata.
func (pkg *Package) M() *Meta { return pkg.Meta }

// Equal checks if this package is equal to `other`. Ignores metadata. Elements
// in slices pkg.Files and pkg.LicenceInfoFromFiles must be in the same
// order for this method to return true.
func (pkg *Package) Equal(other *Package) bool {
	if pkg == other {
		return true
	}
	if pkg == nil || other == nil {
		return false
	}

	eq := pkg.Name.Val == other.Name.Val &&
		pkg.Version.Val == other.Version.Val &&
		len(pkg.LicenceInfoFromFiles) == len(other.LicenceInfoFromFiles) &&
		len(pkg.Files) == len(other.Files) &&
		pkg.DownloadLocation.Val == other.DownloadLocation.Val &&
		pkg.HomePage.Val == other.HomePage.Val &&
		pkg.FileName.Val == other.FileName.Val &&
		pkg.LicenceComments.Val == other.LicenceComments.Val &&
		pkg.CopyrightText.Val == other.CopyrightText.Val &&
		pkg.Summary.Val == other.Summary.Val &&
		pkg.Description.Val == other.Description.Val &&
		pkg.SourceInfo.Val == other.SourceInfo.Val &&
		pkg.Supplier.V() == other.Supplier.V() &&
		pkg.Originator.V() == other.Originator.V() &&
		pkg.Checksum.Equal(other.Checksum) &&
		pkg.VerificationCode.Equal(other.VerificationCode) &&
		SameLicence(pkg.LicenceConcluded, other.LicenceConcluded) &&
		SameLicence(pkg.LicenceDeclared, other.LicenceDeclared)

	if !eq {
		return false
	}
	for i, file := range pkg.Files {
		if file.Equal(other.Files[i]) {
			return false
		}
	}
	for i, lic := range pkg.LicenceInfoFromFiles {
		if !SameLicence(lic, other.LicenceInfoFromFiles[i]) {
			return false
		}
	}
	return true
}

// Represents a package verification code.
type VerificationCode struct {
	Value         ValueStr   // Verification code
	ExcludedFiles []ValueStr // List of excluded file names
	*Meta                    // Verification code metadata.
}

// Package metadata.
func (vc *VerificationCode) M() *Meta { return vc.Meta }

// Equal checks if two VerificationCodes are equal. Ignores metadata. elements
// in the slice ExcludedFiles must be in the same order for this method to
// return true.
func (vc *VerificationCode) Equal(other *VerificationCode) bool {
	if vc == other {
		return true
	}
	if vc == nil || other == nil {
		return false
	}
	if vc.Value.Val != other.Value.Val || len(vc.ExcludedFiles) != len(other.ExcludedFiles) {
		return false
	}
	for i, v := range vc.ExcludedFiles {
		if v.Val != other.ExcludedFiles[i].Val {
			return false
		}
	}
	return true
}

// Represents a file or package checksum.
type Checksum struct {
	Algo, Value ValueStr
	*Meta
}

// Equal compares two checksums, ignoring their metadatas.
func (c *Checksum) Equal(d *Checksum) bool {
	return c == d || (c != nil && d != nil && c.Algo.Val == d.Algo.Val && c.Value.Val == d.Value.Val)
}

// M returns the checksum metadata.
func (c *Checksum) M() *Meta { return c.Meta }
