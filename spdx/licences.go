package spdx

import "strings"

// Every licence struct implements this interface.
type AnyLicence interface {
	Value
	LicenceId() string
}

// Represents a licence in the SPDX Licence list, a licence reference that
// couldn't be changed to it's corresponding ExtractedLicence pointer or
// one of the NONE or NOASSERTION constants.
type Licence struct{ ValueStr }

// LicenceId gets the licence ID.
func (l Licence) LicenceId() string { return l.V() }

// Equal compares two licences ignoring their metadata.
func (l Licence) Equal(b Licence) bool { return l.ValueStr.Equal(b.ValueStr) }

// IsReference returns whether the licence is a reference or a is supposed to be in the SPDX Licence List.
// Does not check if the licence actually is in the licence list (use InList() for that).
func (l Licence) IsReference() bool {
	return isLicIdRef(l.V())
}

// InList checks whether the licence is in the SPDX Licence List.
// It always looks up the SPDX Licence List index.
func (l Licence) InList() bool {
	return CheckLicence(l.V())
}

// Creates a new Licence.
func NewLicence(id string, m *Meta) Licence {
	return Licence{Str(id, m)}
}

// Represents an Extracted Licence.
type ExtractedLicence struct {
	Id             ValueStr
	Name           []ValueStr
	Text           ValueStr
	CrossReference []ValueStr
	Comment        ValueStr
	*Meta
}

// LicenceId returns the licence ID.
func (l *ExtractedLicence) LicenceId() string { return l.Id.V() }
func (l *ExtractedLicence) V() string         { return l.LicenceId() }
func (l *ExtractedLicence) M() *Meta          { return l.Meta }

// Equal checks if this ExtractedLicence is equal to `other`. Ignores metadata.
// Slice elements must be in the same order for this function to return true.
func (l *ExtractedLicence) Equal(other *ExtractedLicence) bool {
	if l == other {
		return true
	}
	if l == nil || other == nil {
		return false
	}
	eq := l.Id.Val == other.Id.Val &&
		len(l.Name) == len(other.Name) &&
		l.Text.Val == other.Text.Val &&
		len(l.CrossReference) == len(other.CrossReference) &&
		l.Comment.Val == other.Comment.Val
	if !eq {
		return false
	}
	for i, v := range l.Name {
		if v.Val != other.Name[i].Val {
			return false
		}
	}
	for i, v := range l.CrossReference {
		if v.Val != other.CrossReference[i].Val {
			return false
		}
	}
	return true
}

// Abstract licence set representation. Both ConjunctiveLicenceSet and
// DisjunctiveLicenceSet are aliases for LicenceSet.
type LicenceSet struct {
	Members []AnyLicence
	*Meta
}

func (s *LicenceSet) Add(lic AnyLicence) { s.Members = append(s.Members, lic) }
func (s *LicenceSet) M() *Meta           { return s.Meta }

// DisjunctiveLicenceSet
type ConjunctiveLicenceSet LicenceSet

func NewConjunctiveSet(meta *Meta, lics ...AnyLicence) ConjunctiveLicenceSet {
	return ConjunctiveLicenceSet{lics, meta}
}
func (c ConjunctiveLicenceSet) LicenceId() string { return join(c.Members, " and ") }
func (c ConjunctiveLicenceSet) V() string         { return c.LicenceId() }
func (c ConjunctiveLicenceSet) M() *Meta          { return c.Meta }

// Add a licence to the set.
func (c *ConjunctiveLicenceSet) Add(lic AnyLicence) { c.Members = append(c.Members, lic) }

// DisjunctiveLicenceSet
type DisjunctiveLicenceSet LicenceSet

func NewDisjunctiveSet(meta *Meta, lics ...AnyLicence) DisjunctiveLicenceSet {
	return DisjunctiveLicenceSet{lics, meta}
}
func (c DisjunctiveLicenceSet) LicenceId() string { return join(c.Members, " or ") }
func (c DisjunctiveLicenceSet) V() string         { return c.LicenceId() }
func (c DisjunctiveLicenceSet) M() *Meta          { return c.Meta }

// Add a licence to the set.
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

// Returns whether the given ID is a Licence Reference ID (starts with LicenseRef).
// Does not check if the string after "LicenseRef" satisfies the requirements of any SPDX version.
// It is case-insensitive.
func isLicIdRef(id string) bool {
	return strings.HasPrefix(strings.ToLower(id), "licenseref")
}

// SameLicence compares two licences. Returns `true` if `a` and `b` are the same,
// returns `false` otherwise. In case of licence sets, it recursively
// applies this function on each element. The licences in sets must be
// in the same order for this function to return true. Ignores metadata.
func SameLicence(a, b AnyLicence) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch ta := a.(type) {
	default:
		return false
	case Licence:
		if tb, ok := b.(Licence); ok && ta.Equal(tb) {
			return true
		}
		return false
	case DisjunctiveLicenceSet:
		if tb, ok := b.(DisjunctiveLicenceSet); ok && len(ta.Members) == len(tb.Members) {
			for i, lica := range ta.Members {
				licb := tb.Members[i]
				if !SameLicence(lica, licb) {
					return false
				}
			}
			return true
		}
		return false
	case ConjunctiveLicenceSet:
		if tb, ok := b.(ConjunctiveLicenceSet); ok && len(ta.Members) == len(tb.Members) {
			for i, lica := range ta.Members {
				licb := tb.Members[i]
				if !SameLicence(lica, licb) {
					return false
				}
			}
			return true
		}
		return false
	case *ExtractedLicence:
		if tb, ok := b.(*ExtractedLicence); ok && ta.Equal(tb) {
			return true
		}
		return false
	}
}
