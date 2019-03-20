package spdx

// Represents a review.
type Review struct {
	Reviewer ValueCreator
	Date     ValueDate
	Comment  ValueStr
	*Meta
}

// M returns the SPDX Review.
func (r *Review) M() *Meta { return r.Meta }

// Equal compares two Review pointers, ignoring any metadata.
func (a *Review) Equal(b *Review) bool {
	return a == b || (a != nil && b != nil &&
		a.Reviewer.V() == b.Reviewer.V() && a.Date.V() == b.Date.V() && a.Comment.Val == b.Comment.Val)
}
