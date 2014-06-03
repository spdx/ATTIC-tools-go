package spdx

type AnyLicenceInfo interface {
	LicenceId() string
}

type LicenceReference struct {
	Id      string
	Licence *Licence
}

func (l LicenceReference) LicenceId() string { return l.Id }

type Licence struct {
	Id               string
	Name             string // optional
	Text             string
	isOsiApproved    bool
	StandardHeader   []string // optional
	StandardTemplate string   // optional
	CrossReference   []string // optional
	Comment          string   // optional
}

func (l *Licence) LicenceId() string { return l.Id }

type ExtractedLicensingInfo struct {
	Id             string
	Name           []string // conditional. one required if the licence is not in the SPDX Licence List
	Text           string
	CrossReference []string //optional
	Comment        string   //optional
}

func (l *ExtractedLicensingInfo) LicenceId() string { return l.Id }

func join(list []AnyLicenceInfo, separator string) string {
	if len(list) == 0 {
		return "()"
	}
	res := "(" + list[0].LicenceId()
	for i := 1; i < len(list); i++ {
		res += separator + list[i].LicenceId()
	}
	res += ")"
	return res
}

// DisjunctiveLicenceList is a AnyLicenceInfo
type ConjunctiveLicenceList []AnyLicenceInfo

func (c ConjunctiveLicenceList) LicenceId() string { return join(c, " and ") }

// DisjunctiveLicenceList is a AnyLicenceInfo
type DisjunctiveLicenceList []AnyLicenceInfo

func (d DisjunctiveLicenceList) LicenceId() string { return join(d, " or ") }

// Used to specify values such as NONE and NOASSERTION
type LicenceStatus struct {
	Status string
}

func (l LicenceStatus) LicenceId() string { return l.Status }
