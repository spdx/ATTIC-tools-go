package spdx

type AnyLicence interface {
	Value
	LicenceId() string
}

type Licence struct {
	Id ValueStr
}

func (l Licence) LicenceId() string { return l.Id.V() }
func (l Licence) M() *Meta          { return l.Id.M() }
func (l Licence) V() string         { return l.LicenceId() }

// Licence comparison ignoring metadata
func (a Licence) Equal(b Licence) bool {
	return a.Id.Val == b.Id.Val
}

func NewLicence(id string, m *Meta) Licence {
	return Licence{Id: Str(id, m)}
}

type ExtractedLicence struct {
	Id             ValueStr
	Name           []ValueStr // conditional. one required if the licence is not in the SPDX Licence List
	Text           ValueStr
	CrossReference []ValueStr //optional
	Comment        ValueStr   //optional
}

func (l *ExtractedLicence) LicenceId() string { return l.Id.V() }
func (l *ExtractedLicence) V() string         { return l.LicenceId() }
func (l *ExtractedLicence) M() *Meta          { return l.Id.M() }

func join(list []AnyLicence, separator string) string {
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

// DisjunctiveLicenceSet is a AnyLicence
type ConjunctiveLicenceSet []AnyLicence

func (c ConjunctiveLicenceSet) LicenceId() string { return join(c, " and ") }
func (c ConjunctiveLicenceSet) V() string         { return c.LicenceId() }
func (c ConjunctiveLicenceSet) M() *Meta {
	for _, k := range c {
		if k != nil {
			return k.M()
		}
	}
	return nil
}

// DisjunctiveLicenceSet is a AnyLicence
type DisjunctiveLicenceSet []AnyLicence

func (c DisjunctiveLicenceSet) LicenceId() string { return join(c, " or ") }
func (c DisjunctiveLicenceSet) V() string         { return c.LicenceId() }
func (c DisjunctiveLicenceSet) M() *Meta {
	for _, k := range c {
		if k != nil {
			return k.M()
		}
	}
	return nil
}
