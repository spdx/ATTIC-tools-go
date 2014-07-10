package spdx

import "strings"

type AnyLicence interface {
	Value
	LicenceId() string
}

type Licence struct{ ValueStr }

func (l Licence) LicenceId() string    { return l.V() }
func (l Licence) Equal(b Licence) bool { return l.ValueStr.Equal(b.ValueStr) }

// Returns whether the licence is a reference or a is supposed to be in the SPDX Licence List.
// Does not check if the licence actually is in the licence list (use InList() for that).
func (l Licence) IsReference() bool {
	return isLicIdRef(l.V())
}

// Checks whether the licence is in the SPDX Licence List.
// It always looks up the SPDX Licence List index.
func (l Licence) InList() bool {
	return CheckLicence(l.V())
}

// Creates a new Licence.
func NewLicence(id string, m *Meta) Licence {
	return Licence{Str(id, m)}
}

type ExtractedLicence struct {
	Id             ValueStr
	Name           []ValueStr // conditional. one required if the licence is not in the SPDX Licence List
	Text           ValueStr
	CrossReference []ValueStr //optional
	Comment        ValueStr   //optional
	*Meta
}

func (l *ExtractedLicence) LicenceId() string { return l.Id.V() }
func (l *ExtractedLicence) V() string         { return l.LicenceId() }
func (l *ExtractedLicence) M() *Meta          { return l.Meta }

// Abstract licence set representation
type LicenceSet struct {
	Members []AnyLicence
	*Meta
}

func (s *LicenceSet) Add(lic AnyLicence) { s.Members = append(s.Members, lic) }

// DisjunctiveLicenceSet
type ConjunctiveLicenceSet LicenceSet

func NewConjunctiveSet(meta *Meta, lics ...AnyLicence) ConjunctiveLicenceSet {
	return ConjunctiveLicenceSet{lics, meta}
}
func (c ConjunctiveLicenceSet) LicenceId() string   { return join(c.Members, " and ") }
func (c ConjunctiveLicenceSet) V() string           { return c.LicenceId() }
func (c ConjunctiveLicenceSet) M() *Meta            { return c.Meta }
func (c *ConjunctiveLicenceSet) Add(lic AnyLicence) { c.Members = append(c.Members, lic) }

// DisjunctiveLicenceSet
type DisjunctiveLicenceSet LicenceSet

func NewDisjunctiveSet(meta *Meta, lics ...AnyLicence) DisjunctiveLicenceSet {
	return DisjunctiveLicenceSet{lics, meta}
}
func (c DisjunctiveLicenceSet) LicenceId() string   { return join(c.Members, " or ") }
func (c DisjunctiveLicenceSet) V() string           { return c.LicenceId() }
func (c DisjunctiveLicenceSet) M() *Meta            { return c.Meta }
func (c *DisjunctiveLicenceSet) Add(lic AnyLicence) { c.Members = append(c.Members, lic) }

// Useful functions for working with licences

// Join the IDs for given licences by separator. Similar
// to strings.Join but for []AnyLicence.
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

// Returns whether the given ID is a Licence Reference ID (starts with LicenseRef-).
// Does not check if the string after "LicenseRef-" satisfies the requirements of any SPDX version.
// It is case-insensitive.
func isLicIdRef(id string) bool {
	return strings.HasPrefix(strings.ToLower(id), "licenseref-")
}
