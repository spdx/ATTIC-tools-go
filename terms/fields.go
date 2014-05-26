package terms

// import "rdfs"

const (
	FT_BINARY  = iota
	FT_SOURCE  = iota
	FT_ARCHIVE = iota
	FT_OTHER   = iota
)

type Document struct {
	SpecVersion      string
	DataLicence      string
	CreationInfo     *CreationInfo
	DescribesPackage []*Package
	// Comment      rdfs.Comment
}

type CreationInfo struct {
	Creator            []string
	Created            string
	LicenceListVersion string
	// Comment rdfs.Comment
}

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

type VerificationCode struct {
	Value         string
	ExcludedFiles []string
}

type Checksum struct {
	Algo, Value string
}

type AnyLicenceInfo struct {
	Licence         *Licence
	Extracted       *ExtractedLicensingInfo
	ConjunctiveList *ConjunctiveLicenceList
	DisjunctiveList *DisjunctiveLicenceList
}

type Licence struct {
	Id               string
	Name             string
	Text             string
	isOsiApproved    bool
	StandardHeader   string
	StandardTemplate string

	// SeeAlso  rdfs.SeeAlso
	// Comment rdfs.Comment
}

type ExtractedLicensingInfo struct {
	Id   string
	Name []string
	Text string
	// SeeAlso rdfs.SeeAlso
	// Comment rdfs.Comment
}

type ConjunctiveLicenceList struct {
	Members []*AnyLicenceInfo
}

type DisjunctiveLicenceList struct {
	Members []*AnyLicenceInfo
}

type SimpleLicenceInfo struct {
	Licence   *Licence
	Extracted *ExtractedLicensingInfo
}

type ArtifactOf struct {
	ProjectUri string
	HomePage   string
	Name       string
}
