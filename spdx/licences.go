package spdx

type AnyLicenceInfo interface {
	Value
	LicenceId() string
}

type LicenceReference struct {
	Id ValueStr
}

func (l LicenceReference) LicenceId() string { return l.Id.V() }
func (l LicenceReference) M() *Meta          { return l.Id.M() }
func (l LicenceReference) V() string         { return l.LicenceId() }

// LicenceReference comparison ignoring metadata
func (a LicenceReference) Equal(b LicenceReference) bool {
	return a.Id.Val == b.Id.Val
}

func NewLicenceReference(id string, m *Meta) LicenceReference {
	return LicenceReference{Id: Str(id, m)}
}

type ExtractedLicensingInfo struct {
	Id             ValueStr
	Name           []ValueStr // conditional. one required if the licence is not in the SPDX Licence List
	Text           ValueStr
	CrossReference []ValueStr //optional
	Comment        ValueStr   //optional
}

func (l *ExtractedLicensingInfo) LicenceId() string { return l.Id.V() }
func (l *ExtractedLicensingInfo) V() string         { return l.LicenceId() }
func (l *ExtractedLicensingInfo) M() *Meta          { return l.Id.M() }

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
func (c ConjunctiveLicenceList) V() string         { return c.LicenceId() }
func (c ConjunctiveLicenceList) M() *Meta {
	for _, k := range c {
		if k != nil {
			return k.M()
		}
	}
	return nil
}

// DisjunctiveLicenceList is a AnyLicenceInfo
type DisjunctiveLicenceList []AnyLicenceInfo

func (c DisjunctiveLicenceList) LicenceId() string { return join(c, " or ") }
func (c DisjunctiveLicenceList) V() string         { return c.LicenceId() }
func (c DisjunctiveLicenceList) M() *Meta {
	for _, k := range c {
		if k != nil {
			return k.M()
		}
	}
	return nil
}
