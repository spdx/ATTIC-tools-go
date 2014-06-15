package spdx

type Review struct {
	Reviewer ValueCreator // mandatory
	Date     ValueStr     // mandatory
	Comment  ValueStr     // optional
}

func (a *Review) Equal(b *Review) bool {
	return a.Reviewer.V() == b.Reviewer.V() && a.Date.Val == b.Date.Val && a.Comment.Val == b.Comment.Val
}
