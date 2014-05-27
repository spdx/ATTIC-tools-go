package spdx

type AnyLicenceInfo struct {
	Value           interface{}
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
