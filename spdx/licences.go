package spdx

type AnyLicenceInfo interface {
	Value
	LicenceId() string
}

type LicenceReference struct {
	Id      ValueStr
	Licence *Licence
}

func (l LicenceReference) LicenceId() string { return l.Id.V() }
func (l LicenceReference) M() *Meta          { return l.Id.M() }
func (l LicenceReference) V() string         { return l.LicenceId() }

// LicenceReference comparison ignoring metadata
func (a LicenceReference) Equal(b LicenceReference) bool {
	return a.Id.Val == b.Id.Val && a.Licence.Equal(b.Licence)
}

func NewLicenceReference(id string, m *Meta) LicenceReference {
	return LicenceReference{Id: Str(id, m)}
}

type Licence struct {
	Id               ValueStr
	Name             ValueStr // optional
	Text             ValueStr
	isOsiApproved    ValueBool
	StandardHeader   []ValueStr // optional
	StandardTemplate ValueStr   // optional
	CrossReference   []ValueStr // optional
	Comment          ValueStr   // optional
}

func (l *Licence) LicenceId() string { return l.Id.V() }
func (l *Licence) V() string         { return l.LicenceId() }
func (l *Licence) M() *Meta          { return nil }

// Licence comparison ignoring metadata
func (a *Licence) Equal(b *Licence) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Id.Equal(b.Id) &&
		a.Name.Equal(b.Name) &&
		a.Text.Equal(b.Text) &&
		a.isOsiApproved.Val == b.isOsiApproved.Val &&
		len(a.StandardHeader) == len(b.StandardHeader) &&
		a.StandardTemplate.Equal(b.StandardTemplate) &&
		len(a.CrossReference) == len(b.CrossReference) &&
		a.Comment.Equal(b.Comment)

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
func (l *ExtractedLicensingInfo) M() *Meta          { return nil }

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
func (c ConjunctiveLicenceList) M() *Meta          { return nil }

// DisjunctiveLicenceList is a AnyLicenceInfo
type DisjunctiveLicenceList []AnyLicenceInfo

func (d DisjunctiveLicenceList) LicenceId() string { return join(d, " or ") }
func (c DisjunctiveLicenceList) V() string         { return c.LicenceId() }
func (c DisjunctiveLicenceList) M() *Meta          { return nil }
