package spdx

type Review struct {
	Reviewer ValueCreator // mandatory
	Date     ValueDate    // mandatory
	Comment  ValueStr     // optional
	*Meta
}

func (r *Review) M() *Meta { return r.Meta }

func (a *Review) Equal(b *Review) bool {
	return a.Reviewer.V() == b.Reviewer.V() && a.Date.V() == b.Date.V() && a.Comment.Val == b.Comment.Val
}
